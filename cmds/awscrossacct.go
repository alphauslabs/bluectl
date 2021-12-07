package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alphauslabs/blue-sdk-go/admin/v1"
	"github.com/alphauslabs/blue-sdk-go/api"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
			resp, err := client.GetDefaultBillingInfo(ctx, &admin.GetDefaultBillingInfoRequest{
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

func UpdateDefaultCrossAcctAccessInfo() *cobra.Command {
	var (
		wait bool
	)

	cmd := &cobra.Command{
		Use:   "update <account>",
		Short: "Update cross-account access",
		Long: `Update cross-account access. Recommended when the status is 'outdated', which means there is an
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
			resp, err := client.UpdateDefaultBillingInfoRole(ctx, &admin.UpdateDefaultBillingInfoRoleRequest{
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
								if v, ok := op.Result.(*api.Operation_Response); ok {
									var r api.KeyValue
									anypb.UnmarshalTo(v.Response, &r, proto.UnmarshalOptions{})
									logger.Infof("%v=%v", r.Key, r.Value)
								}

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
		UpdateDefaultCrossAcctAccessInfo(),
		DelDefaultCrossAcctAccessInfo(),
	)

	return cmd
}
