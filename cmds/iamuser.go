package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alphauslabs/blue-sdk-go/iam/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ListIamUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subusers",
		Long:  `List subusers.`,
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
			stream, err := client.ListUsers(ctx, &iam.ListUsersRequest{})
			if err != nil {
				fnerr(err)
				return
			}

			hdrs := []string{"ID", "PARENT"}

			switch {
			case params.OutFile != "" && params.OutFmt == "csv":
				if params.OutFile != "" {
					var f *os.File
					var wf *csv.Writer
					f, err = os.Create(params.OutFile)
					if err != nil {
						fnerr(err)
						return
					}

					wf = csv.NewWriter(f)
					defer func() {
						wf.Flush()
						f.Close()
					}()

					wf.Write(hdrs)
					for {
						v, err := stream.Recv()
						if err == io.EOF {
							break
						}

						if err != nil {
							fnerr(err)
							return
						}

						row := []string{v.Id, v.Parent}
						logger.Infof("%v --> %v", row, params.OutFile)
						wf.Write(row)
					}
				}
			case params.OutFmt == "json":
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
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetColWidth(100)
				table.SetBorder(false)
				table.SetHeaderLine(false)
				table.SetColumnSeparator("")
				table.SetTablePadding("  ")
				table.SetNoWhiteSpace(true)
				table.SetHeader(hdrs)

				for {
					v, err := stream.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						fnerr(err)
						return
					}

					table.Append([]string{v.Id, v.Parent})
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func GetIamUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get subuser information",
		Long:  `Get subuser information.`,
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
				fnerr(fmt.Errorf("id is required"))
				return
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
			resp, err := client.GetUser(ctx, &iam.GetUserRequest{Id: args[0]})
			if err != nil {
				fnerr(err)
				return
			}

			hdrs := []string{"ID", "PARENT", "METADATA"}

			switch {
			case params.OutFile != "" && params.OutFmt == "csv":
				if params.OutFile != "" {
					var f *os.File
					var wf *csv.Writer
					f, err = os.Create(params.OutFile)
					if err != nil {
						fnerr(err)
						return
					}

					wf = csv.NewWriter(f)
					defer func() {
						wf.Flush()
						f.Close()
					}()

					wf.Write(hdrs)
					for k, v := range resp.Metadata {
						m := fmt.Sprintf("%v: %v", k, v)
						row := []string{resp.Id, resp.Parent, m}
						logger.Infof("%v --> %v", row, params.OutFile)
						wf.Write(row)
					}
				}
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				logger.Info(string(b))
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetColWidth(100)
				table.SetBorder(false)
				table.SetHeaderLine(false)
				table.SetColumnSeparator("")
				table.SetTablePadding("  ")
				table.SetNoWhiteSpace(true)
				table.SetHeader(hdrs)

				for k, v := range resp.Metadata {
					m := fmt.Sprintf("%v: %v", k, v)
					row := []string{resp.Id, resp.Parent, m}
					table.Append(row)
				}

				table.Render()
			}

		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func IamUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iamuser",
		Short: "Subcommand for IAM user-related operations",
		Long:  `Subcommand for IAM user-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		ListIamUsersCmd(),
		GetIamUserCmd(),
	)

	return cmd
}
