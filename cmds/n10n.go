package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alphauslabs/blue-sdk-go/admin/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func ListChannelsCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notification channels",
		Long:  `List notification channels.`,
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
			con, err := grpcconn.GetConnection(ctx, grpcconn.AdminService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := admin.NewClient(ctx, &admin.ClientOptions{Conn: con})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var resp *admin.ListNotificationChannelsResponse
			switch {
			case rawInput != "":
				var in admin.ListNotificationChannelsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				resp, err = client.ListNotificationChannels(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				resp, err = client.ListNotificationChannels(ctx, &admin.ListNotificationChannelsRequest{})
				if err != nil {
					fnerr(err)
					return
				}
			}

			switch {
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				if err != nil {
					fnerr(err)
					return
				}

				fmt.Printf("%v", string(b))
			default:
				var m admin.ListNotificationChannelsResponse
				b, _ := json.Marshal(resp)
				err = yaml.Unmarshal(b, &m)
				if err != nil {
					fnerr(err)
					return
				}

				b, _ = yaml.Marshal(m)
				fmt.Printf("%v", string(b))
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Admin/Admin_ListNotificationChannels")
	return cmd
}

func ChannelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "Subcommand for notification channels",
		Long:  `Subcommand for notification channels.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(ListChannelsCmd())
	return cmd
}

func NotificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "n10n",
		Short: "Subcommand for notifications",
		Long:  `Subcommand for notifications.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(ChannelsCmd())
	return cmd
}
