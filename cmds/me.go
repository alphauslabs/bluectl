package cmds

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"

	"github.com/alphauslabs/blue-sdk-go/blueaws/v1"
	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/loginurl"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func MeCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "me",
		Short: "Get information of me as a user",
		Long:  `Get information of me as a user.`,
		Run: func(cmd *cobra.Command, args []string) {
			var opts []grpc.DialOption
			creds := credentials.NewTLS(&tls.Config{})
			opts = append(opts, grpc.WithTransportCredentials(creds))
			opts = append(opts, grpc.WithBlock())
			opts = append(opts, grpc.WithPerRPCCredentials(
				session.NewRpcCredentials(session.RpcCredentialsInput{
					LoginUrl:     loginurl.LoginUrl(),
					ClientId:     params.ClientId,
					ClientSecret: params.ClientSecret,
				}),
			))

			conn, err := grpc.DialContext(context.Background(), session.BlueAwsEndpoint, opts...)
			if err != nil {
				log.Fatalln("DialContext failed:", err)
			}

			defer conn.Close()
			client := blueaws.NewBlueAwsClient(conn)
			resp, err := client.Me(context.Background(), &blueaws.MeRequest{})
			if err != nil {
				log.Fatalln("Me failed:", err)
			}

			b, _ := json.Marshal(resp)
			logger.Info(string(b))
		},
	}

	c.Flags().SortFlags = false
	return c
}
