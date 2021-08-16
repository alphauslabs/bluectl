package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alphauslabs/blue-sdk-go/api"
	awstypes "github.com/alphauslabs/blue-sdk-go/api/aws"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
	"github.com/spf13/cobra"
)

func AwsGetCostsCmd() *cobra.Command {
	var (
		rawInput              string
		costtype              string
		id                    string
		start                 string
		end                   string
		includeTags           bool
		includeCostCategories bool
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
						"billingGroupId",
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
							v.BillingGroupId,
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
							fmt.Sprintf("%.f", v.ExchangeRate),
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

			var stream cost.Cost_ReadCostsClient

			switch {
			case rawInput != "":
				var in cost.ReadCostsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
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

				in := cost.ReadCostsRequest{
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
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCosts")
	cmd.Flags().StringVar(&costtype, "type", "account", "type of cost to read: all, account, billinggroup")
	cmd.Flags().StringVar(&id, "id", id, "account id or billing group id, depending on --type, skipped if 'all'")
	cmd.Flags().StringVar(&start, "start", time.Now().UTC().Format("200601")+"01", "yyyymmdd: start date to stream data; default: first day of the current month (UTC)")
	cmd.Flags().StringVar(&end, "end", time.Now().UTC().Format("20060102"), "yyyymmdd: end date to stream data; default: current date (UTC)")
	cmd.Flags().BoolVar(&includeTags, "include-tags", includeTags, "if true, include tags in the stream")
	cmd.Flags().BoolVar(&includeCostCategories, "include-costcategories", includeCostCategories, "if true, include cost categories in the stream")
	return cmd
}

func AwsGetAdjustmentsCmd() *cobra.Command {
	var (
		rawInput string
		id       string
		start    string
		end      string
		costtype string
	)

	cmd := &cobra.Command{
		Use:   "get-adjustments",
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
						"name",
						"billingGroupId",
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

			fnWriteFile := func(v *awstypes.Adjustment) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.BillingGroupId,
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

func AwsCalculateCostsCmd() *cobra.Command {
	var (
		rawInput string
		wait     bool
	)

	cmd := &cobra.Command{
		Use:   "calculate",
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

func AwsCostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "awscost [id]",
		Short: "Subcommand for AWS costs-related operations",
		Long:  `Subcommand for AWS costs-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		AwsGetCostsCmd(),
		AwsGetAdjustmentsCmd(),
		AwsCalculateCostsCmd(),
	)

	return cmd
}
