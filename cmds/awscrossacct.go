package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/alphauslabs/blue-sdk-go/admin/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func CreateDefaultCrossAcctAccessInfo() *cobra.Command {
	var (
		region string
	)

	cmd := &cobra.Command{
		Use:   "create <account>",
		Short: "Create a default cross-account access",
		Long: `Create a default cross-account access. You will be presented with link to a CloudFormation deployment.
You can deploy the template manually as well using StackSets if you prefer, in which case, you have to
deploy manually. The command will work all the same, although you have to run for each target account.`,
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
				fnerr(fmt.Errorf("account is required"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.AdminService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := admin.NewClient(ctx, &admin.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.GetDefaultBillingInfoTemplateUrl(ctx, &admin.GetDefaultBillingInfoTemplateUrlRequest{
				Region: region,
			})

			if err != nil {
				fnerr(err)
				return
			}

			fmt.Println("Open the link below in your browser and deploy:")
			fmt.Println(resp.LaunchUrl)
			var rep string
			fmt.Print("Confirm successful deployment? [Y/n]: ")
			fmt.Scanln(&rep)

			switch strings.ToLower(rep) {
			case "n":
				return
			case "":
				fallthrough
			case "y":
				fmt.Println("Validating access...")
				resp, err := client.CreateDefaultBillingInfoRole(ctx, &admin.CreateDefaultBillingInfoRoleRequest{
					Target: args[0],
				})

				if err != nil {
					fnerr(err)
					return
				}

				b, _ := json.Marshal(resp)
				fmt.Println(string(b))
			default:
				fnerr(fmt.Errorf("unknown reply"))
				return
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&region, "region", region, "optional, the AWS region code (i.e. 'us-east-1')to deploy the CloudFormation template")
	return cmd
}

func GetDefaultCrossAcctAccessInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account>",
		Short: "Get cross-account access information",
		Long:  `Get cross-account access information.`,
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
				fnerr(fmt.Errorf("account is required"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.AdminService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := admin.NewClient(ctx, &admin.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.GetDefaultBillingInfo(ctx, &admin.GetDefaultBillingInfoRequest{Target: args[0]})
			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFile != "" && params.OutFmt == "csv":
				logger.Info("format not supported yet")
			default:
				b, _ := json.Marshal(resp)
				logger.Info(string(b))
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func DelDefaultCrossAcctAccessInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <account>",
		Short: "Remove cross-account access",
		Long:  `Remove cross-account access. This does not delete the CloudFormation stack.`,
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
				fnerr(fmt.Errorf("account is required"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.AdminService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := admin.NewClient(ctx, &admin.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			_, err = client.DeleteDefaultBillingInfoRole(ctx, &admin.DeleteDefaultBillingInfoRoleRequest{Target: args[0]})
			if err != nil {
				fnerr(err)
				return
			}

			logger.Info("cross-account access removed")
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func CrossAcctAccessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "xacct",
		Short: "Subcommand for AWS cross-account access operations",
		Long:  `Subcommand for AWS cross-account access operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		CreateDefaultCrossAcctAccessInfo(),
		GetDefaultCrossAcctAccessInfo(),
		DelDefaultCrossAcctAccessInfo(),
	)

	return cmd
}
