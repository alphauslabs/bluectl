package awsbilling

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/billing/v1"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func CalculationHistoryCmd() *cobra.Command {
	var (
		month string
	)

	cmd := &cobra.Command{
		Use:   "calhist [yyyymm]",
		Short: "Query calculation history for all accounts",
		Long:  `Query calculation history for all accounts.`,
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

				fmt.Printf("billingInternalId: %v\n", v.BillingInternalId)
				for _, acct := range v.Accounts {
					if len(acct.History) > 0 {
						fmt.Printf("  accountId: %v\n", acct.AccountId)
						for _, h := range acct.History {
							fmt.Printf("    timestamp=%v, trigger=%v\n", h.Timestamp, h.Trigger)
						}
					}
				}
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
