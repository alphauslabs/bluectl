package cmds

import (
	"context"
	"encoding/json"

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
			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, "blue")
			if err != nil {
				logger.Error(err)
				return
			}

			client, err := iam.NewClient(ctx, &iam.ClientOptions{Conn: mycon})
			if err != nil {
				logger.Error(err)
				return
			}

			defer client.Close()
			resp, err := client.WhoAmI(ctx, &iam.WhoAmIRequest{})
			if err != nil {
				logger.Error(err)
				return
			}

			b, _ := json.Marshal(resp)
			logger.Info(string(b))
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
