package awsbilling

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/billing/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func CostsDriftCmd() *cobra.Command {
	var (
		month string
	)

	cmd := &cobra.Command{
		Use:   "costsdrift [yyyymm] [billingInternalId]",
		Short: "Query differences, if any, between invoice and latest costs",
		Long:  `Query differences, if any, between invoice and latest costs.`,
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

			month = time.Now().UTC().Format("200601")
			if len(args) >= 1 {
				mm, err := time.Parse("200601", args[0])
				if err != nil {
					fnerr(err)
					return
				} else {
					month = mm.Format("200601")
				}
			}

			var comp string
			if len(args) >= 2 {
				comp = args[1]
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.BillingService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := billing.NewClient(ctx, &billing.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			stream, err := client.ListUsageCostsDrift(ctx, &billing.ListUsageCostsDriftRequest{
				Vendor:            "aws",
				BillingInternalId: comp,
				Month:             month,
			})

			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFile != "" && params.OutFmt == "csv":
				if params.OutFile != "" {
					var f *os.File
					var wf *csv.Writer
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

					wf.Write([]string{
						"billingInternalId",
						"billingGroupId",
						"account",
						"month",
						"snapshot",
						"current",
						"diff",
						"drift",
					})

					for {
						v, err := stream.Recv()
						if err == io.EOF {
							break
						}

						if err != nil {
							fnerr(err)
							return
						}

						row := []string{
							v.BillingInternalId,
							v.BillingGroupId,
							v.Account,
							month,
							fmt.Sprintf("%f", v.Snapshot),
							fmt.Sprintf("%f", v.Current),
							fmt.Sprintf("%f", v.Diff),
						}

						logger.Infof("%v --> %v", row, params.OutFile)
						wf.Write(row)
					}
				}
			case params.OutFmt == "json":
				logger.Info("format not supported yet")
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				table.SetColumnAlignment([]int{
					tablewriter.ALIGN_LEFT,
					tablewriter.ALIGN_LEFT,
					tablewriter.ALIGN_LEFT,
					tablewriter.ALIGN_LEFT,
					tablewriter.ALIGN_RIGHT,
					tablewriter.ALIGN_RIGHT,
					tablewriter.ALIGN_RIGHT,
				})

				table.SetColWidth(100)
				table.SetBorder(false)
				table.SetHeaderLine(false)
				table.SetColumnSeparator("")
				table.SetTablePadding("  ")
				table.SetNoWhiteSpace(true)
				table.SetHeader([]string{
					"INTERNAL_ID",
					"BILLING_GROUP_ID",
					"ACCOUNT",
					"MONTH",
					"SNAPSHOT",
					"CURRENT",
					"DIFF",
				})

				for {
					v, err := stream.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						fnerr(err)
						return
					}

					vf := func(f float64) string {
						if f == 0 {
							return "%f"
						} else {
							return "%.9f"
						}
					}

					row := []string{
						v.BillingInternalId,
						v.BillingGroupId,
						v.Account,
						month,
						fmt.Sprintf(vf(v.Snapshot), v.Snapshot),
						fmt.Sprintf(vf(v.Current), v.Current),
						fmt.Sprintf(vf(v.Diff), math.Abs(v.Diff)),
					}

					fmt.Printf("\033[2K\rrecv:%v...", row)
					table.Append(row)
				}

				fmt.Printf("\033[2K\r") // reset cursor
				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
