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
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func AwsCostCmd() *cobra.Command {
	var (
		typ                   string
		start                 string
		end                   string
		includeTags           bool
		includeCostCategories bool
	)

	cmd := &cobra.Command{
		Use:   "aws-costs [id]",
		Short: "Read AWS usage-based costs",
		Long: `Read AWS usage-based costs based on the type. If --type is 'all', [id] is discarded.
If 'account', it should be an AWS account id. If 'billinggroup', it should be a billing group id.`,
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

			if typ != "all" {
				if len(args) == 0 {
					fnerr(fmt.Errorf("id is required"))
					return
				}
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, "cost")
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
						"productCode",
						"serviceCode",
						"region",
						"zone",
						"usageType",
						"instanceType",
						"operation",
						"invoiceId",
						"description",
						"tags",
						"costCategories",
						"usageAmount",
						"cost",
					})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWriteFile := func(name string, v *awstypes.Cost) {
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
							name,
							v.BillingGroupId,
							v.Account,
							v.Date.AsTime().Format(time.RFC3339),
							v.ProductCode,
							v.ServiceCode,
							v.Region,
							v.Zone,
							v.UsageType,
							v.InstanceType,
							v.Operation,
							v.InvoiceId,
							v.Description,
							tags,
							cc,
							fmt.Sprintf("%.9f", v.Usage),
							fmt.Sprintf("%.9f", v.Cost),
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
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

			switch typ {
			case "all":
				stream, err := client.ReadCosts(ctx,
					&cost.ReadCostsRequest{
						Vendor:    "aws",
						StartTime: ts.Format("20060102"),
						EndTime:   te.Format("20060102"),
						AwsOptions: &cost.AwsOptions{
							IncludeTags:           includeTags,
							IncludeCostCategories: includeCostCategories,
						},
					},
				)

				if err != nil {
					fnerr(err)
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

					fnWriteFile("all", v.Aws)
				}
			case "account":
				stream, err := client.ReadAccountCosts(ctx,
					&cost.ReadAccountCostsRequest{
						Vendor:    "aws",
						Name:      args[0],
						StartTime: ts.Format("20060102"),
						EndTime:   te.Format("20060102"),
						AwsOptions: &cost.AwsOptions{
							IncludeTags:           includeTags,
							IncludeCostCategories: includeCostCategories,
						},
					},
				)

				if err != nil {
					fnerr(err)
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

					fnWriteFile(args[0], v.Aws)
				}
			case "billinggroup":
				stream, err := client.ReadBillingGroupCosts(ctx,
					&cost.ReadBillingGroupCostsRequest{
						Vendor:    "aws",
						Name:      args[0],
						StartTime: ts.Format("20060102"),
						EndTime:   te.Format("20060102"),
						AwsOptions: &cost.AwsOptions{
							IncludeTags:           includeTags,
							IncludeCostCategories: includeCostCategories,
						},
					},
				)

				if err != nil {
					fnerr(err)
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

					fnWriteFile(args[0], v.Aws)
				}
			default:
				fnerr(fmt.Errorf("type unsupported: %v", typ))
				return
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&typ, "type", "account", "type of cost to read: all, account, billinggroup")
	cmd.Flags().StringVar(&start, "start", time.Now().UTC().Format("2006-01")+"-01", "yyyy-mm-dd: start date to stream data; default: first day of the current month (UTC)")
	cmd.Flags().StringVar(&end, "end", time.Now().UTC().Format("2006-01-02"), "yyyy-mm-dd: end date to stream data; default: current date (UTC)")
	cmd.Flags().BoolVar(&includeTags, "include-tags", includeTags, "if true, include tags in the stream")
	cmd.Flags().BoolVar(&includeCostCategories, "include-costcategories", includeCostCategories, "if true, include cost categories in the stream")
	return cmd
}
