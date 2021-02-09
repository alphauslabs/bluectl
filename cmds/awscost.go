package cmds

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/awscost/v1"
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
			switch typ {
			case "all":
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

					log.Println(v.Date.AsTime().Format(time.RFC3339), v, err)
				}
			case "company":
			case "billinggroup":
			default:
				fnerr(fmt.Errorf("type unsupported: %v", typ))
				return
			}
		},
	}

	c.Flags().SortFlags = false
	c.Flags().StringVar(&typ, "type", "account", "type of cost to stream: all, account, company, billinggroup")
	return c
}
