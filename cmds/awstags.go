package cmds

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	awstypes "github.com/alphauslabs/blue-sdk-go/api/aws"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func AwsGetTagsCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Read AWS tags-based costs",
		Long: `Read AWS tags-based costs. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadTagCosts.
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
						"tagDetails",
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
					td := v.TagId
					if td != "" {
						dec, err := base64.StdEncoding.DecodeString(td)
						if err == nil {
							td = string(dec)
						}
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
							td,
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			var stream cost.Cost_ReadTagCostsClient

			switch {
			case rawInput != "":
				var in cost.ReadTagCostsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				stream, err = client.ReadTagCosts(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				logger.Error("not yet implemented, see --raw-input")
				return
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
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadTagCosts")
	return cmd
}

func AwsGetNonTagsCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "get-nontag",
		Short: "Read AWS nontag-based costs",
		Long: `Read AWS nontag-based costs. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadNonTagCosts.
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

			var stream cost.Cost_ReadNonTagCostsClient

			switch {
			case rawInput != "":
				var in cost.ReadNonTagCostsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				stream, err = client.ReadNonTagCosts(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				logger.Error("not yet implemented, see --raw-input")
				return
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
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadTagCosts")
	return cmd
}

func AwsTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "awstags",
		Short: "Subcommand for AWS [non]tags-related operations",
		Long:  `Subcommand for AWS [non]tags-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		AwsGetTagsCmd(),
		AwsGetNonTagsCmd(),
	)

	return cmd
}
