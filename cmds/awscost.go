package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/awscost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/loginurl"
	"github.com/spf13/cobra"
)

func AwsCostCmd() *cobra.Command {
	var (
		typ string
	)

	c := &cobra.Command{
		Use:   "awscost [id]",
		Short: "Stream your AWS usage-based costs",
		Long: `Stream your AWS usage-based costs based on the type. If --type is 'all', [id] is discarded.
If 'account', it should be an AWS account id. If 'company', it should be a company id.
If 'billinggroup', it should be a billing group id.`,
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
			client, err := awscost.NewClient(ctx, awscost.WithLoginUrl(loginurl.LoginUrl()))
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
						"usageAmount",
						"cost",
					})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWriteFile := func(name string, v *awscost.Cost) {
				if params.OutFile != "" {
					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							name,
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
							fmt.Sprintf("%.9f", v.UsageAmount),
							fmt.Sprintf("%.9f", v.Cost),
						})
					case "json":
						b, _ := json.Marshal(v)
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			switch typ {
			case "all":
				stream, err := client.StreamReadCosts(ctx,
					&awscost.StreamReadCostsRequest{
						Name: args[0],
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

					fnWriteFile(args[0], v)
				}
			case "account":
				stream, err := client.StreamReadAccountCosts(ctx,
					&awscost.StreamReadAccountCostsRequest{
						Name: args[0],
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

					fnWriteFile(args[0], v)
				}
			case "company":
				stream, err := client.StreamReadCompanyCosts(ctx,
					&awscost.StreamReadCompanyCostsRequest{
						Name: args[0],
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

					fnWriteFile(args[0], v)
				}
			case "billinggroup":
				stream, err := client.StreamReadBillingGroupCosts(ctx,
					&awscost.StreamReadBillingGroupCostsRequest{
						Name: args[0],
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

					fnWriteFile(args[0], v)
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

	c.Flags().SortFlags = false
	c.Flags().StringVar(&typ, "type", "account", "type of cost to stream: all, account, company, billinggroup")
	return c
}
