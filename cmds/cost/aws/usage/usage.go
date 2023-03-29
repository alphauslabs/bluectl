package usage

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	awstypes "github.com/alphauslabs/blue-sdk-go/api/aws"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

type Flags struct {
	RawInput              string
	CostType              string
	Id                    string
	Start                 string
	End                   string
	IncludeTags           bool
	IncludeCostCategories bool
	ColWidth              int
}

func get(cmd *cobra.Command, args []string, fl *Flags) {
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
	case fl.RawInput != "":
		err := json.Unmarshal([]byte(fl.RawInput), &in)
		if err != nil {
			fnerr(err)
			return
		}

		if in.Vendor == "" {
			in.Vendor = "aws"
		}

		if in.AwsOptions == nil {
			// Let's default to monthly services.
			in.AwsOptions = &cost.ReadCostsRequestAwsOptions{
				GroupByColumns: "productCode",
				GroupByMonth:   true,
			}
		}

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

		stream, err = client.ReadCosts(ctx, &in)
		if err != nil {
			fnerr(err)
			return
		}
	default:
		if fl.CostType != "all" {
			if fl.Id == "" {
				fnerr(fmt.Errorf("id is required"))
				return
			}
		}

		var ts, te time.Time
		if fl.Start != "" {
			ts, err = time.Parse("20060102", fl.Start)
			if err != nil {
				fnerr(err)
				return
			}

		}

		if fl.End != "" {
			te, err = time.Parse("20060102", fl.End)
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
				IncludeTags:           fl.IncludeTags,
				IncludeCostCategories: fl.IncludeCostCategories,
			},
		}

		switch fl.CostType {
		case "account":
			in.AccountId = fl.Id
		case "billinggroup":
			in.GroupId = fl.Id
		default:
			fnerr(fmt.Errorf("type unsupported: %v", fl.CostType))
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
	table.SetColWidth(fl.ColWidth)
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
}

func GetCmd() *cobra.Command {
	fl := Flags{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read AWS usage-based costs",
		Long: `Read AWS usage-based costs. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCosts.
Note that this will invalidate all the other flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			get(cmd, args, &fl)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&fl.RawInput, "raw-input", fl.RawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCosts")
	cmd.Flags().StringVar(&fl.CostType, "type", "account", "type of cost to read: all, account, billinggroup")
	cmd.Flags().StringVar(&fl.Id, "id", fl.Id, "account id or billing group id, depending on --type, skipped if 'all'")
	cmd.Flags().StringVar(&fl.Start, "start", time.Now().UTC().Format("200601")+"01", "yyyymmdd: start date to stream data; default: first day of the current month (UTC)")
	cmd.Flags().StringVar(&fl.End, "end", time.Now().UTC().Format("20060102"), "yyyymmdd: end date to stream data; default: current date (UTC)")
	cmd.Flags().BoolVar(&fl.IncludeTags, "include-tags", fl.IncludeTags, "if true, include tags in the stream")
	cmd.Flags().BoolVar(&fl.IncludeCostCategories, "include-costcategories", fl.IncludeCostCategories, "if true, include cost categories in the stream")
	cmd.Flags().IntVar(&fl.ColWidth, "col-width", 30, "set column width, applies to table-based outputs only")
	return cmd
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage",
		Short: "AWS-specific usage-based cost subcommand",
		Long:  `AWS-specific usage-based cost subcommand.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(GetCmd())
	return cmd
}
