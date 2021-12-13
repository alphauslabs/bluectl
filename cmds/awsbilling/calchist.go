package awsbilling

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/billing/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func CalculationHistoryCmd() *cobra.Command {
	var (
		red   = color.New(color.FgRed).SprintFunc()
		month string
	)

	cmd := &cobra.Command{
		Use:   "calchist [yyyymm]",
		Short: "Query calculation history for all accounts",
		Long: `Query calculation history for all accounts.
The default output format is:

billingInternalId/billingGroupId (yyyymm):
  accountId: timestamp=timestamp, trigger='cur|invoice'

Timestamps are ordered with the topmost as most recent. 'cur'-triggered means this calculation was
triggered by updates to the CUR while 'invoice' means by a manual invoice request.`,
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
			if len(args) > 0 {
				mm, err := time.Parse("200601", args[0])
				if err != nil {
					fnerr(err)
					return
				} else {
					month = mm.Format("200601")
				}
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
			stream, err := client.ListAwsCalculationHistory(ctx, &billing.ListAwsCalculationHistoryRequest{
				Month: month,
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
						"month",
						"account",
						"timestamp",
						"trigger",
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

						if len(v.Accounts) == 0 {
							continue
						}

						for _, acct := range v.Accounts {
							if len(acct.History) > 0 {
								for _, h := range acct.History {
									row := []string{
										v.BillingInternalId,
										v.BillingGroupId,
										v.Month,
										acct.AccountId,
										h.Timestamp,
										h.Trigger,
									}

									logger.Infof("%v --> %v", row, params.OutFile)
									wf.Write(row)
								}
							}
						}
					}
				}
			case params.OutFmt == "json":
				logger.Info("format not supported yet")
			default:
				for {
					v, err := stream.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						fnerr(err)
						return
					}

					if len(v.Accounts) == 0 {
						continue
					}

					fmt.Printf("%v/%v (%v)\n", v.BillingInternalId, v.BillingGroupId, v.Month)
					for _, acct := range v.Accounts {
						if len(acct.History) > 0 {
							var itr int
							var updated bool // after invoice
							for _, h := range acct.History {
								itr++
								if h.Trigger == "invoice" {
									if itr > 1 {
										updated = true
									}
									break
								}
							}

							for _, h := range acct.History {
								if updated && h.Trigger == "invoice" {
									fmt.Printf(red("  %v: timestamp=%v, trigger=%v\n"),
										acct.AccountId, h.Timestamp, h.Trigger)
								} else {
									fmt.Printf("  %v: timestamp=%v, trigger=%v\n",
										acct.AccountId, h.Timestamp, h.Trigger)
								}
							}
						}
					}
				}
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
