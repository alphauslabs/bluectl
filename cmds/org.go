package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alphauslabs/blue-sdk-go/org/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func OrgRegisterCmd() *cobra.Command {
	var (
		passwd string
		desc   string
	)

	cmd := &cobra.Command{
		Use:   "create <email>",
		Short: "Create a new organization",
		Long:  `Create a new organization.`,
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

			if len(args) == 0 {
				fnerr(fmt.Errorf("<email> cannot be empty"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.OrgService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := org.NewClient(ctx, &org.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.CreateOrg(ctx, &org.CreateOrgRequest{})
			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				logger.Info(string(b))
			default:
				logger.Info(resp)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&passwd, "password", passwd, "your org password")
	cmd.Flags().StringVar(&desc, "description", desc, "your org description (e.g. org name)")
	return cmd
}

func OrgGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Print information about your organization",
		Long:  `Print information about your organization.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.OrgService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := org.NewClient(ctx, &org.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.GetOrg(ctx, &org.GetOrgRequest{})
			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				logger.Info(string(b))
			default:
				logger.Info(resp)
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func OrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Subcommand for Organization operations",
		Long:  `Subcommand for Organization operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		OrgRegisterCmd(),
		OrgGetCmd(),
	)

	return cmd
}
