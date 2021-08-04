package cmds

import (
	"context"
	"encoding/json"
	"os"

	"github.com/alphauslabs/blue-sdk-go/iam/v1"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func WhoAmICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Get my information as a user",
		Long:  `Get my information as a user.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.IamService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := iam.NewClient(ctx, &iam.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.WhoAmI(ctx, &iam.WhoAmIRequest{})
			if err != nil {
				fnerr(err)
				return
			}

			b, _ := json.Marshal(resp)
			logger.Info(string(b))
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
