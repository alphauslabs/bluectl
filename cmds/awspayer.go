package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alphauslabs/blue-sdk-go/api"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
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
		rawInput string
		month    string
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
			var f *os.File
			var wf *csv.Writer
			hdrs := []string{"PAYER", "MONTH", "TIMESTAMP"}
			var stream cost.Cost_GetPayerAccountImportHistoryClient

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
			var render bool

			if params.OutFile != "" {
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

				switch params.OutFmt {
				case "csv":
					wf.Write(hdrs)
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			switch {
			case rawInput != "":
				var in cost.GetPayerAccountImportHistoryRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				stream, err = client.GetPayerAccountImportHistory(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				if len(args) == 0 {
					fnerr(fmt.Errorf("id is required"))
					return
				}

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

				stream, err = client.GetPayerAccountImportHistory(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			}

			fnWrite := func(v *cost.GetPayerAccountImportHistoryResponse) {
				switch {
				case params.OutFile != "" && params.OutFmt == "csv":
					for _, t := range v.Timestamps {
						wf.Write([]string{v.Id, v.Month, t})
					}
				case params.OutFmt == "json":
					b, _ := json.Marshal(v)
					fmt.Println(string(b))
				default:
					render = true
					for _, t := range v.Timestamps {
						table.Append([]string{v.Id, v.Month, t})
					}
				}
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

				fnWrite(v)
			}

			if render {
				table.Render()
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_GetPayerAccountImportHistory")
	cmd.Flags().StringVar(&month, "month", time.Now().UTC().Format("200601"), "import month (UTC), fmt: yyyymm")
	return cmd
}

func ImportCursCmd() *cobra.Command {
	var (
		rawInput string
		month    string
		wait     bool
	)

	cmd := &cobra.Command{
		Use:   "import-curs [id1[,id2,id...]]",
		Short: "Trigger an ondemand import of all (or input) CUR files",
		Long:  `Trigger an ondemand import of all (or input) CUR files.`,
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
			var resp *api.Operation

			switch {
			case rawInput != "":
				var in cost.ImportCurFilesRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				resp, err = client.ImportCurFiles(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				in := cost.ImportCurFilesRequest{}
				if len(args) > 0 {
					in.Filter = args[0]
				}

				resp, err = client.ImportCurFiles(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			}

			b, _ := json.Marshal(resp)
			logger.Info(string(b))

			if wait {
				func() {
					defer func(begin time.Time) {
						logger.Info("duration:", time.Since(begin))
					}(time.Now())

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
									logger.Infof("[%v] done", resp.Name)
									done <- struct{}{}
									return
								}
							}
						}
					}()

					logger.Infof("wait for [%v], this could take some time...", resp.Name)

					select {
					case <-done:
					case <-quit.Done():
						logger.Info("interrupted")
					}
				}()
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ImportCurFiles")
	cmd.Flags().StringVar(&month, "month", time.Now().UTC().Format("200601"), "import month (UTC), fmt: yyyymm")
	cmd.Flags().BoolVar(&wait, "wait", wait, "if true, wait for the operation to finish")
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
		ImportCursCmd(),
	)

	return cmd
}
