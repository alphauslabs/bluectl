package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ListPayersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registered payer accounts",
		Long:  `List registered payer accounts.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			stream, err := client.ListPayerAccounts(ctx, &cost.ListPayerAccountsRequest{
				Vendor: "aws",
			})

			if err != nil {
				fnerr(err)
				return
			}

			hdrs := []string{"ID", "NAME"}

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

						row := []string{v.Id, v.Name}
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

					table.Append([]string{v.Id, v.Name})
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func GetPayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Query a registered payer account",
		Long:  `Query a registered payer account.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.GetPayerAccount(ctx, &cost.GetPayerAccountRequest{
				Vendor: "aws",
				Id:     args[0],
			})

			if err != nil {
				fnerr(err)
				return
			}

			hdrs := []string{"ID", "NAME", "METADATA"}

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
					for _, v := range resp.Metadata {
						m := fmt.Sprintf("%v: %v", v.Key, v.Value)
						wf.Write([]string{resp.Id, resp.Name, m})
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

				for _, v := range resp.Metadata {
					m := fmt.Sprintf("%v: %v", v.Key, v.Value)
					table.Append([]string{resp.Id, resp.Name, m})
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func CurImportHistoryCmd() *cobra.Command {
	var (
		month string
	)

	cmd := &cobra.Command{
		Use:   "get-curhistory <id>",
		Short: "Query an AWS management account's CUR import history",
		Long:  `Query an AWS management account's CUR import history.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.CostService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := cost.NewClient(ctx, &cost.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			mm, err := time.Parse("200601", month)
			if err != nil {
				fnerr(err)
				return
			}

			in := cost.GetPayerAccountImportHistoryRequest{
				Vendor: "aws",
				Id:     args[0],
				Month:  mm.Format("200601"),
			}

			resp, err := client.GetPayerAccountImportHistory(ctx, &in)
			if err != nil {
				fnerr(err)
				return
			}

			hdrs := []string{"PAYER", "MONTH", "TIMESTAMP"}

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
					for _, v := range resp.Timestamps {
						row := []string{resp.Id, resp.Month, v}
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

				for _, v := range resp.Timestamps {
					table.Append([]string{resp.Id, resp.Month, v})
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&month, "month", time.Now().UTC().Format("200601"), "import month (UTC), fmt: yyyymm")
	return cmd
}

func AwsPayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "awspayer",
		Short: "Subcommand for AWS management account-related operations",
		Long:  `Subcommand for AWS management account-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		ListPayersCmd(),
		GetPayerCmd(),
		CurImportHistoryCmd(),
	)

	return cmd
}
