package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alphauslabs/blue-sdk-go/admin/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
	"github.com/spf13/cobra"
)

func CreateDefaultCostAccessInfo() *cobra.Command {
	var (
		silent bool
	)

	cmd := &cobra.Command{
		Use:   "create <account> [type]",
		Short: "Create default cost access",
		Long: `Create default cost access. You will be presented with link to a CloudFormation deployment based on type.

Valid values for the optional [type] are:
  apionly - Read-only access to cost information without CUR setup.
  s3only  - Setup S3 bucket compatible for CUR definition export. Useful if you prefer a different
            region other than the default.

You can deploy the template manually as well using StackSets if you prefer, in which case, you have to
deploy manually. The command will work all the same, although you have to run for each target account.

For Wave(Pro) accounts, we recommended you to deploy this stack (type=apionly). This will allow us to
query a more accurate billing-related information such as your Reserved Instances, Savings Plans, etc.
through the AWS API. Currently, we only do a best-effort detection of these information from the parent
CUR, which is not always accurate.

The stack template will create an IAM role with read-only access to your billing-related information.
If you want to audit the template, it is publicly available from the link below:

  default:
  https://alphaus-cloudformation-templates.s3.ap-northeast-1.amazonaws.com/alphauscurexportdef-v1.yml

  apionly:
  https://alphaus-cloudformation-templates.s3.ap-northeast-1.amazonaws.com/alphausdefaultcostaccess-v1.yml

  s3only:
  https://alphaus-cloudformation-templates.s3.ap-northeast-1.amazonaws.com/alphauscurexportbucket-v1.yml`,
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

			var s3only bool
			req := admin.GetDefaultCostAccessTemplateUrlRequest{}
			if len(args) >= 2 {
				switch args[1] {
				case "": // empty is default (valid)
				case "apionly":
					req.Type = args[1]
				case "s3only":
					req.Type = args[1]
					s3only = true
				default:
					fnerr(fmt.Errorf("unknown type: %v", args[1]))
					return
				}
			}

			defer client.Close()
			resp, err := client.GetDefaultCostAccessTemplateUrl(ctx, &req)
			if err != nil {
				fnerr(err)
				return
			}

			fmt.Println("Open the link below in your browser and deploy:")
			fmt.Println(resp.LaunchUrl)
			if s3only {
				fmt.Println("\nTo use the deployed bucket, rerun this command with the default type (empty) then select the 'USE_EXISTING' parameter in your CloudFormation console.")
				return
			}

			var rep string
			if !silent {
				fmt.Print("Confirm successful deployment? [Y/n]: ")
				fmt.Scanln(&rep)
			}

			switch strings.ToLower(rep) {
			case "n":
				return
			case "":
				fallthrough
			case "y":
				fmt.Println("Validating access...")
				resp, err := client.CreateDefaultCostAccess(ctx, &admin.CreateDefaultCostAccessRequest{
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
	cmd.Flags().BoolVar(&silent, "silent", silent, "if true, no input required (non-interactive)")
	return cmd
}

func ListDefaultCostAccessInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List default cost access information",
		Long:  `List default cost access information.`,
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
			stream, err := client.ListDefaultCostAccess(ctx, &admin.ListDefaultCostAccessRequest{})
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

				b, _ := json.Marshal(v)
				logger.Info(string(b))
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func GetDefaultCostAccessInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <account>",
		Short: "Get default cost access information",
		Long:  `Get default cost access information.`,
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
			resp, err := client.GetDefaultCostAccess(ctx, &admin.GetDefaultCostAccessRequest{
				Target: args[0],
			})

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

func UpdateDefaultCostAccessInfo() *cobra.Command {
	var (
		wait bool
	)

	cmd := &cobra.Command{
		Use:   "update <account>",
		Short: "Update default cost access",
		Long: `Update default cost access. Recommended when the status is 'outdated', which means there is an
update to the CloudFormation template.`,
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
			resp, err := client.UpdateDefaultCostAccess(ctx, &admin.UpdateDefaultCostAccessRequest{
				Target: args[0],
			})

			if err != nil {
				fnerr(err)
				return
			}

			logger.Infof("operation=%v", resp.Name)

			if wait {
				quit, cancel := context.WithCancel(context.Background())
				done := make(chan struct{}, 1)

				// Interrupt handler.
				go func() {
					sigch := make(chan os.Signal)
					signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
					<-sigch
					cancel()
				}()

				go func() {
					for {
						q := context.WithValue(quit, struct{}{}, nil)
						op, err := ops.WaitForOperation(q, ops.WaitForOperationInput{
							Name: resp.Name,
						})

						if err != nil {
							logger.Error(err)
							done <- struct{}{}
							return
						}

						if op != nil {
							if op.Done {
								done <- struct{}{}
								return
							}
						}
					}
				}()

				logger.Infof("wait for %v, this could take some time...", resp.Name)

				select {
				case <-done:
				case <-quit.Done():
					logger.Info("interrupted")
				}
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVar(&wait, "wait", wait, "wait for the update to finish")
	return cmd
}

func DelDefaultCostAccessInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <account>",
		Short: "Remove default cost access",
		Long:  `Remove default cost access. This does not delete the CloudFormation stack.`,
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
			_, err = client.DeleteDefaultCostAccess(ctx, &admin.DeleteDefaultCostAccessRequest{
				Target: args[0],
			})

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
		Short: "Subcommand for AWS cost access operations",
		Long:  `Subcommand for AWS cost access operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		CreateDefaultCostAccessInfo(),
		ListDefaultCostAccessInfo(),
		GetDefaultCostAccessInfo(),
		UpdateDefaultCostAccessInfo(),
		DelDefaultCostAccessInfo(),
	)

	return cmd
}
