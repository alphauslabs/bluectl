package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	awstypes "github.com/alphauslabs/blue-sdk-go/api/aws"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func AwsFeesCmd() *cobra.Command {
	var (
		rawInput string
		start    string
		end      string
		costtype string
	)

	cmd := &cobra.Command{
		Use:   "aws-adjustments [id]",
		Short: "Read AWS adjustment costs",
		Long: `Read AWS adjustment costs based on the type. If --type is 'all', [id] is discarded.
If 'account', it should be an AWS account id. If 'billinggroup', it should be a billing group id.

You can also use the --raw-input flag and provide the full JSON string input described in
https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadAdjustments. This option is provided
so you can utilize all the API features not yet supported by the other flags. Note that
this will invalidate all other flags.`,
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
			client, err := cost.NewClient(ctx)
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

			fnWriteFile := func(name string, v *awstypes.Adjustment) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							name,
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

				stream, err = client.ReadAdjustments(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				if costtype != "all" {
					if len(args) == 0 {
						fnerr(fmt.Errorf("id is required"))
						return
					}
				}

				var ts, te time.Time
				if start != "" {
					ts, err = time.Parse("2006-01-02", start)
					if err != nil {
						fnerr(err)
						return
					}
				}

				if end != "" {
					te, err = time.Parse("2006-01-02", end)
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
					in.AccountId = args[0]
				case "billinggroup":
					in.BillingInternalId = args[0]
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

				fnWriteFile("all", v.Aws)
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadAdjustments")
	cmd.Flags().StringVar(&costtype, "type", "account", "type of cost to stream: all, account, billinggroup")
	cmd.Flags().StringVar(&start, "start", time.Now().UTC().Format("2006-01")+"-01", "yyyy-mm-dd: start date to stream data; default: first day of the current month (UTC)")
	cmd.Flags().StringVar(&end, "end", time.Now().UTC().Format("2006-01-02"), "yyyy-mm-dd: end date to stream data; default: current date (UTC)")
	return cmd
}
