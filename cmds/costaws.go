package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alphauslabs/blue-sdk-go/api"
	awstypes "github.com/alphauslabs/blue-sdk-go/api/aws"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func CostAwsAttributesGetCmd() *cobra.Command {
	var (
		rawInput string
		colWidth int
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get AWS cost attributes",
		Long: `Get AWS cost attributes. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCostAttributes.
Note that this will invalidate all the other flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var f *os.File
			var wf *csv.Writer

			if params.OutFile != "" {
				f, err = os.Create(params.OutFile)
				if err != nil {
					fnerr(err)
					return
				}

				wf = csv.NewWriter(f)
				defer func() {
					wf.Flush()
					f.Close()
				}()

				switch params.OutFmt {
				case "csv":
					wf.Write([]string{
						"groupId",
						"account",
						"productCode",
						"serviceCode",
						"region",
						"zone",
						"usageType",
						"instanceType",
						"operation",
						"invoiceId",
						"description",
						"resourceId",
						"tags",
						"costCategories",
					})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWriteFile := func(v *awstypes.CostAttribute) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					var tags, cc string
					if v.Tags != nil {
						b, _ := json.Marshal(v.Tags)
						tags = string(b)
					}

					if v.CostCategories != nil {
						b, _ := json.Marshal(v.CostCategories)
						cc = string(b)
					}

					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.GroupId,
							v.Account,
							v.ProductCode,
							v.ServiceCode,
							v.Region,
							v.Zone,
							v.UsageType,
							v.InstanceType,
							v.Operation,
							v.InvoiceId,
							v.Description,
							v.ResourceId,
							tags,
							cc,
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			type colT struct {
				name   string
				title  string
				enable bool
				val    interface{}
				vfmt   string
			}

			var stream cost.Cost_ReadCostAttributesClient
			var in cost.ReadCostAttributesRequest
			cols := []string{}
			refCols := []colT{
				{name: "account", title: "ACCOUNT", enable: true},
				{name: "productCode", title: "SERVICE"},
				{name: "serviceCode", title: "SERVICECODE"},
				{name: "region", title: "REGION"},
				{name: "zone", title: "ZONE"},
				{name: "usageType", title: "USAGE_TYPE"},
				{name: "instanceType", title: "INSTANCE_TYPE"},
				{name: "operation", title: "OPERATION"},
				{name: "invoiceId", title: "INVOICE_ID"},
				{name: "description", title: "DESCRIPTION"},
				{name: "resourceId", title: "RESOURCE_ID"},
				{name: "tags", title: "TAGS"},
				{name: "costCategories", title: "COST_CATEGORIES"},
			}

			switch {
			case rawInput != "":
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				if in.AwsOptions != nil {
					if in.AwsOptions.Dimensions != "" {
						gbcs := strings.Split(in.AwsOptions.Dimensions, ",")
						for _, gbc := range gbcs {
							for i, rc := range refCols {
								if rc.name == gbc {
									refCols[i].enable = true
									break
								}
							}
						}
					} else {
						for i := range refCols {
							refCols[i].enable = true
						}
					}
				}

				stream, err = client.ReadCostAttributes(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				logger.Info("please use --raw-input for now, sorry")
				return
			}

			for _, v := range refCols {
				if v.enable {
					cols = append(cols, v.title)
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetColWidth(colWidth)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetTablePadding("  ")
			table.SetNoWhiteSpace(true)
			table.SetHeader(cols)
			var render bool

			for {
				v, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err != nil {
					fnerr(err)
					return
				}

				switch {
				case params.OutFile != "":
					fnWriteFile(v.Aws)
				default:
					render = true
					refCols[0].val = v.Aws.Account
					refCols[1].val = v.Aws.ProductCode
					refCols[2].val = v.Aws.ServiceCode
					refCols[3].val = v.Aws.Region
					refCols[4].val = v.Aws.Zone
					refCols[5].val = v.Aws.UsageType
					refCols[6].val = v.Aws.InstanceType
					refCols[7].val = v.Aws.Operation
					refCols[8].val = v.Aws.InvoiceId
					refCols[9].val = v.Aws.Description
					refCols[10].val = v.Aws.ResourceId
					refCols[11].val = v.Aws.Tags
					refCols[12].val = v.Aws.CostCategories
					row := []string{}
					for _, rc := range refCols {
						if rc.enable {
							if (rc.name == "tags" || rc.name == "costCategories") && rc.val != nil {
								ms := []string{}
								m := rc.val.(map[string]string)
								for k, v := range m {
									ms = append(ms, fmt.Sprintf("%v:%v", k, v))
								}

								jms := strings.Join(ms, ",")
								row = append(row, fmt.Sprintf("%v", jms))
							} else {
								row = append(row, fmt.Sprintf("%v", rc.val))
							}
						}
					}

					fmt.Printf("\033[2K\rrecv:%v...", row)
					table.Append(row)
				}
			}

			if render {
				fmt.Printf("\033[2K\r") // reset cursor
				table.Render()
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCostAttributes")
	cmd.Flags().IntVar(&colWidth, "col-width", 30, "set column width, applies to table-based outputs only")
	return cmd
}

func CostAwsAttributesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attributes",
		Short: "AWS-specific cost attributes subcommand",
		Long:  `AWS-specific cost attributes subcommand.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(CostAwsAttributesGetCmd())
	return cmd
}

func CostAwsGetCmd() *cobra.Command {
	var (
		rawInput              string
		costtype              string
		id                    string
		start                 string
		end                   string
		includeTags           bool
		includeCostCategories bool
		colWidth              int
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read AWS usage-based costs",
		Long: `Read AWS usage-based costs. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCosts.
Note that this will invalidate all the other flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var f *os.File
			var wf *csv.Writer

			if params.OutFile != "" {
				f, err = os.Create(params.OutFile)
				if err != nil {
					fnerr(err)
					return
				}

				wf = csv.NewWriter(f)
				defer func() {
					wf.Flush()
					f.Close()
				}()

				switch params.OutFmt {
				case "csv":
					wf.Write([]string{
						"groupId",
						"account",
						"date",
						"productCode",
						"serviceCode",
						"region",
						"zone",
						"usageType",
						"instanceType",
						"operation",
						"invoiceId",
						"description",
						"resourceId",
						"tags",
						"costCategories",
						"usageAmount",
						"cost",
						"baseCurrency",
						"exchangeRate",
						"targetCost",
						"targetCurrency",
						"effectiveCost",
						"targetEffectiveCost",
						"amortizedCost",
						"targetAmortizedCost",
					})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWriteFile := func(v *awstypes.Cost) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					var tags, cc string
					if v.Tags != nil {
						b, _ := json.Marshal(v.Tags)
						tags = string(b)
					}

					if v.CostCategories != nil {
						b, _ := json.Marshal(v.CostCategories)
						cc = string(b)
					}

					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.GroupId,
							v.Account,
							v.Date,
							v.ProductCode,
							v.ServiceCode,
							v.Region,
							v.Zone,
							v.UsageType,
							v.InstanceType,
							v.Operation,
							v.InvoiceId,
							v.Description,
							v.ResourceId,
							tags,
							cc,
							fmt.Sprintf("%.9f", v.Usage),
							fmt.Sprintf("%.9f", v.Cost),
							v.BaseCurrency,
							fmt.Sprintf("%.9f", v.ExchangeRate),
							fmt.Sprintf("%.9f", v.TargetCost),
							v.TargetCurrency,
							fmt.Sprintf("%.9f", v.EffectiveCost),
							fmt.Sprintf("%.9f", v.TargetEffectiveCost),
							fmt.Sprintf("%.9f", v.AmortizedCost),
							fmt.Sprintf("%.9f", v.TargetAmortizedCost),
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			type colT struct {
				name   string
				title  string
				enable bool
				val    interface{}
				vfmt   string
			}

			var stream cost.Cost_ReadCostsClient
			var in cost.ReadCostsRequest
			cols := []string{}
			refCols := []colT{
				{name: "group", title: "GROUP"},
				{name: "account", title: "ACCOUNT"},
				{name: "date", title: "DATE", enable: true},
				{name: "productCode", title: "SERVICE"},
				{name: "serviceCode", title: "SERVICECODE"},
				{name: "region", title: "REGION"},
				{name: "zone", title: "ZONE"},
				{name: "usageType", title: "USAGE_TYPE"},
				{name: "instanceType", title: "INSTANCE_TYPE"},
				{name: "operation", title: "OPERATION"},
				{name: "invoiceId", title: "INVOICE_ID"},
				{name: "description", title: "DESCRIPTION"},
				{name: "resourceId", title: "RESOURCE_ID"},
				{name: "tags", title: "TAGS"},
				{name: "costCategories", title: "COST_CATEGORIES"},
				{name: "usage", title: "USAGE", enable: true, vfmt: "%.10f"},
				{name: "cost", title: "COST", enable: true, vfmt: "%.10f"},
			}

			switch {
			case rawInput != "":
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				if in.AwsOptions != nil {
					if in.AwsOptions.GroupByColumns != "" {
						refCols[0].enable = true // group
						refCols[1].enable = true // account
						gbcs := strings.Split(in.AwsOptions.GroupByColumns, ",")
						for _, gbc := range gbcs {
							for i, rc := range refCols {
								if rc.name == gbc {
									refCols[i].enable = true
									break
								}
							}
						}
					} else {
						for i := range refCols {
							refCols[i].enable = true
						}
					}

					refCols[13].enable = in.AwsOptions.IncludeTags
					refCols[14].enable = in.AwsOptions.IncludeCostCategories
				}

				stream, err = client.ReadCosts(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				if costtype != "all" {
					if id == "" {
						fnerr(fmt.Errorf("id is required"))
						return
					}
				}

				var ts, te time.Time
				if start != "" {
					ts, err = time.Parse("20060102", start)
					if err != nil {
						fnerr(err)
						return
					}

				}

				if end != "" {
					te, err = time.Parse("20060102", end)
					if err != nil {
						fnerr(err)
						return
					}
				}

				in = cost.ReadCostsRequest{
					Vendor:    "aws",
					StartTime: ts.Format("20060102"),
					EndTime:   te.Format("20060102"),
					AwsOptions: &cost.ReadCostsRequestAwsOptions{
						IncludeTags:           includeTags,
						IncludeCostCategories: includeCostCategories,
					},
				}

				switch costtype {
				case "account":
					in.AccountId = id
				case "billinggroup":
					in.GroupId = id
				default:
					fnerr(fmt.Errorf("type unsupported: %v", costtype))
					return
				}

				stream, err = client.ReadCosts(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			}

			for _, v := range refCols {
				if v.enable {
					cols = append(cols, v.title)
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetColWidth(colWidth)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetTablePadding("  ")
			table.SetNoWhiteSpace(true)
			table.SetHeader(cols)
			var render bool

			for {
				v, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err != nil {
					fnerr(err)
					return
				}

				switch {
				case params.OutFile != "":
					fnWriteFile(v.Aws)
				default:
					render = true
					refCols[0].val = v.Aws.GroupId
					refCols[1].val = v.Aws.Account
					refCols[2].val = v.Aws.Date
					refCols[3].val = v.Aws.ProductCode
					refCols[4].val = v.Aws.ServiceCode
					refCols[5].val = v.Aws.Region
					refCols[6].val = v.Aws.Zone
					refCols[7].val = v.Aws.UsageType
					refCols[8].val = v.Aws.InstanceType
					refCols[9].val = v.Aws.Operation
					refCols[10].val = v.Aws.InvoiceId
					refCols[11].val = v.Aws.Description
					refCols[12].val = v.Aws.ResourceId
					refCols[13].val = v.Aws.Tags
					refCols[14].val = v.Aws.CostCategories
					refCols[15].val = v.Aws.Usage
					refCols[16].val = v.Aws.Cost
					row := []string{}
					for _, rc := range refCols {
						if rc.enable {
							vfmt := "%v"
							if rc.vfmt != "" {
								vfmt = rc.vfmt
							}

							row = append(row, fmt.Sprintf(vfmt, rc.val))
						}
					}

					fmt.Printf("\033[2K\rrecv:%v...", row)
					table.Append(row)
				}
			}

			if render {
				fmt.Printf("\033[2K\r") // reset cursor
				table.Render()
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCosts")
	cmd.Flags().StringVar(&costtype, "type", "account", "type of cost to read: all, account, billinggroup")
	cmd.Flags().StringVar(&id, "id", id, "account id or billing group id, depending on --type, skipped if 'all'")
	cmd.Flags().StringVar(&start, "start", time.Now().UTC().Format("200601")+"01", "yyyymmdd: start date to stream data; default: first day of the current month (UTC)")
	cmd.Flags().StringVar(&end, "end", time.Now().UTC().Format("20060102"), "yyyymmdd: end date to stream data; default: current date (UTC)")
	cmd.Flags().BoolVar(&includeTags, "include-tags", includeTags, "if true, include tags in the stream")
	cmd.Flags().BoolVar(&includeCostCategories, "include-costcategories", includeCostCategories, "if true, include cost categories in the stream")
	cmd.Flags().IntVar(&colWidth, "col-width", 30, "set column width, applies to table-based outputs only")
	return cmd
}

func CostAwsAdjustmentsGetCmd() *cobra.Command {
	var (
		rawInput string
		id       string
		start    string
		end      string
		costtype string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read AWS adjustment costs",
		Long: `Read AWS adjustment costs. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadAdjustments.
Note that this will invalidate all the other flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var f *os.File
			var wf *csv.Writer

			if params.OutFile != "" {
				f, err = os.Create(params.OutFile)
				if err != nil {
					fnerr(err)
					return
				}

				wf = csv.NewWriter(f)
				defer func() {
					wf.Flush()
					f.Close()
				}()

				switch params.OutFmt {
				case "csv":
					wf.Write([]string{
						"groupId",
						"account",
						"date",
						"type",
						"productCode",
						"description",
						"cost",
						"baseCurrency",
						"exchangeRate",
						"targetCost",
						"targetCurrency",
					})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWriteFile := func(v *awstypes.Cost) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.GroupId,
							v.Account,
							v.Date,
							v.Type,
							v.ProductCode,
							v.Description,
							fmt.Sprintf("%.9f", v.Cost),
							v.BaseCurrency,
							fmt.Sprintf("%f", v.ExchangeRate),
							fmt.Sprintf("%.9f", v.TargetCost),
							v.TargetCurrency,
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			var stream cost.Cost_ReadAdjustmentsClient

			switch {
			case rawInput != "":
				var in cost.ReadAdjustmentsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				stream, err = client.ReadAdjustments(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				if costtype != "all" {
					if id == "" {
						fnerr(fmt.Errorf("--id is required"))
						return
					}
				}

				var ts, te time.Time
				if start != "" {
					ts, err = time.Parse("20060102", start)
					if err != nil {
						fnerr(err)
						return
					}
				}

				if end != "" {
					te, err = time.Parse("20060102", end)
					if err != nil {
						fnerr(err)
						return
					}
				}

				in := cost.ReadAdjustmentsRequest{
					Vendor:    "aws",
					StartTime: ts.Format("20060102"),
					EndTime:   te.Format("20060102"),
				}

				switch costtype {
				case "account":
					in.AccountId = id
				case "billinggroup":
					in.GroupId = id
				default:
					fnerr(fmt.Errorf("type unsupported: %v", costtype))
					return
				}

				stream, err = client.ReadAdjustments(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			}

			for {
				v, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err != nil {
					fnerr(err)
					return
				}

				fnWriteFile(v.Aws)
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadAdjustments")
	cmd.Flags().StringVar(&id, "id", id, "account id or billing group id, depending on --type, skipped if 'all'")
	cmd.Flags().StringVar(&costtype, "type", "account", "type of cost to stream: all, account, billinggroup")
	cmd.Flags().StringVar(&start, "start", time.Now().UTC().Format("200601")+"01", "yyyymmdd: start date to stream data; default: first day of the current month (UTC)")
	cmd.Flags().StringVar(&end, "end", time.Now().UTC().Format("20060102"), "yyyymmdd: end date to stream data; default: current date (UTC)")
	return cmd
}

func CostAwsAdjustmentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adjustments",
		Short: "AWS-specific cost adjustments subcommand",
		Long:  `AWS-specific cost adjustments subcommand.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(CostAwsAdjustmentsGetCmd())
	return cmd
}

func CostAwsCalculationsRunCmd() *cobra.Command {
	var (
		rawInput string
		wait     bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Trigger an ondemand AWS costs calculation",
		Long:  `Trigger an ondemand AWS costs calculation.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var resp *api.Operation

			switch {
			case rawInput != "":
				var in cost.CalculateCostsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				resp, err = client.CalculateCosts(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				resp, err = client.CalculateCosts(ctx, &cost.CalculateCostsRequest{
					Vendor: "aws",
				})

				if err != nil {
					fnerr(err)
					return
				}
			}

			b, _ := json.Marshal(resp)
			logger.Info(string(b))

			if wait {
				func() {
					defer func(begin time.Time) {
						logger.Info("duration:", time.Since(begin))
					}(time.Now())

					quit, cancel := context.WithCancel(context.Background())
					done := make(chan struct{}, 1)

					// Interrupt handler.
					go func() {
						sigch := make(chan os.Signal)
						signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
						<-sigch
						cancel()
					}()

					go func() {
						for {
							q := context.WithValue(quit, struct{}{}, nil)
							op, err := ops.WaitForOperation(q, ops.WaitForOperationInput{
								Name: resp.Name,
							})

							if err != nil {
								logger.Error(err)
								done <- struct{}{}
								return
							}

							if op != nil {
								if op.Done {
									logger.Infof("[%v] done", resp.Name)
									done <- struct{}{}
									return
								}
							}
						}
					}()

					logger.Infof("wait for [%v], this could take some time...", resp.Name)

					select {
					case <-done:
					case <-quit.Done():
						logger.Info("interrupted")
					}
				}()
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_CalculateCosts")
	cmd.Flags().BoolVar(&wait, "wait", wait, "if true, wait for the operation to finish")
	return cmd
}

func CostAwsCalculationsListRunningCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-running [month]",
		Short: "List accounts that are still processing",
		Long: `List accounts that are still processing. The format for [month] is yyyymm.
If [month] is not provided, it defaults to the current UTC month.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			mm := time.Now().UTC().Format("200601")
			if len(args) > 0 {
				_, err := time.Parse("200601", args[0])
				if err != nil {
					fnerr(err)
				} else {
					mm = args[0]
				}
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var f *os.File
			var wf *csv.Writer

			if params.OutFile != "" {
				f, err = os.Create(params.OutFile)
				if err != nil {
					fnerr(err)
					return
				}

				wf = csv.NewWriter(f)
				defer func() {
					wf.Flush()
					f.Close()
				}()

				switch params.OutFmt {
				case "csv":
					wf.Write([]string{"month", "account", "date", "started"})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWrite := func(v *cost.ListCalculatorRunningAccountsResponse) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.Aws.Month,
							v.Aws.Account,
							v.Aws.Date,
							v.Aws.Started,
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			stream, err := client.ListCalculatorRunningAccounts(ctx,
				&cost.ListCalculatorRunningAccountsRequest{
					Vendor: "aws",
					Month:  mm,
				},
			)

			if err != nil {
				fnerr(err)
				return
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetColWidth(100)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetTablePadding("  ")
			table.SetNoWhiteSpace(true)
			table.SetHeader([]string{"MONTH", "ACCOUNT", "DATE", "STARTED"})
			var render bool

			for {
				v, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err != nil {
					fnerr(err)
					return
				}

				switch {
				case params.OutFile != "":
					fnWrite(v)
				default:
					render = true
					row := []string{
						v.Aws.Month,
						v.Aws.Account,
						v.Aws.Date,
						v.Aws.Started,
					}

					fmt.Printf("\033[2K\rrecv:%v...", row)
					table.Append(row)
				}
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}

			if render {
				fmt.Printf("\033[2K\r") // reset cursor
				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func CostAwsCalculationsListHistoryCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "list-history",
		Short: "Query AWS calculation history",
		Long:  `Query AWS calculation history.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var f *os.File
			var wf *csv.Writer
			hdrs := []string{"NAME", "MONTH", "GROUPS", "UPDATED", "CREATED", "STATUS", "DONE", "RESULT"}
			var resp *cost.ListCalculationsHistoryResponse

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetColWidth(100)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetTablePadding("  ")
			table.SetNoWhiteSpace(true)
			table.SetHeader(hdrs)
			var render bool

			if params.OutFile != "" {
				f, err = os.Create(params.OutFile)
				if err != nil {
					fnerr(err)
					return
				}

				wf = csv.NewWriter(f)
				defer func() {
					wf.Flush()
					f.Close()
				}()

				switch params.OutFmt {
				case "csv":
					wf.Write(hdrs)
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			switch {
			case rawInput != "":
				var in cost.ListCalculationsHistoryRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				resp, err = client.ListCalculationsHistory(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				resp, err = client.ListCalculationsHistory(ctx, &cost.ListCalculationsHistoryRequest{
					Vendor: "aws",
				})

				if err != nil {
					fnerr(err)
					return
				}
			}

			for _, op := range resp.Aws.Operations {
				var meta api.OperationAwsCalculateCostsMetadataV1
				anypb.UnmarshalTo(op.Metadata, &meta, proto.UnmarshalOptions{})
				var result string
				switch op.Result.(type) {
				case *api.Operation_Response:
					var res api.KeyValue
					tres := op.Result.(*api.Operation_Response)
					anypb.UnmarshalTo(tres.Response, &res, proto.UnmarshalOptions{})
					result = fmt.Sprintf("%v", res.Value)
				case *api.Operation_Error:
					terr := op.Result.(*api.Operation_Error)
					result = terr.Error.String()
				}

				switch {
				case params.OutFmt == "csv" && params.OutFile != "":
					wf.Write([]string{
						op.Name,
						meta.Month,
						strings.Join(meta.GroupIds, ","),
						meta.Updated,
						meta.Created,
						meta.Status,
						fmt.Sprintf("%v", op.Done),
						result,
					})
				case params.OutFmt == "json":
					var m map[string]interface{}
					b, _ := json.Marshal(op)
					json.Unmarshal(b, &m)

					// Make metadata more readable.
					v := m["metadata"]
					vv := v.(map[string]interface{})
					vv["value"] = meta

					// Make result more readable.
					switch op.Result.(type) {
					case *api.Operation_Response:
						var res api.KeyValue
						tres := op.Result.(*api.Operation_Response)
						anypb.UnmarshalTo(tres.Response, &res, proto.UnmarshalOptions{})
						v := m["Result"]
						vv := v.(map[string]interface{})
						vvv := vv["Response"].(map[string]interface{})
						vvv["value"] = res
					}

					b, _ = json.Marshal(m)
					fmt.Println(string(b))
				default:
					render = true
					table.Append([]string{
						op.Name,
						meta.Month,
						strings.Join(meta.GroupIds, ","),
						meta.Updated,
						meta.Created,
						meta.Status,
						fmt.Sprintf("%v", op.Done),
						result,
					})
				}
			}

			if render {
				table.Render()
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ListCalculationsHistory")
	return cmd
}

func CostAwsCalculationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c10s",
		Short: "Subcommand for AWS-specific calculations",
		Long:  `Subcommand for AWS-specific calculations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		CostAwsCalculationsRunCmd(),
		CostAwsCalculationsListRunningCmd(),
		CostAwsCalculationsListHistoryCmd(),
	)

	return cmd
}

func CostAwsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS-specific cost-related subcommands",
		Long:  `AWS-specific cost-related subcommands.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		CostAwsAttributesCmd(),
		CostAwsGetCmd(),
		CostAwsAdjustmentsCmd(),
		CostAwsCalculationsCmd(),
	)

	return cmd
}
