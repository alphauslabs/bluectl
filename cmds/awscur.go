package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func CurImportHistoryCmd() *cobra.Command {
	var (
		month string
	)

	cmd := &cobra.Command{
		Use:   "aws-curhistory <id>",
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
				fnerr(fmt.Errorf("<id> can't be empty"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, "cost")
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

					wf.Write([]string{"payer", "month", "timestamps"})
					for _, v := range resp.Timestamps {
						wf.Write([]string{resp.Id, resp.Month, v})
					}
				}
				fallthrough
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				logger.SetCleanOutput()
				logger.Info(string(b))
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAutoFormatHeaders(false)
				table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetColWidth(100)
				table.SetHeader([]string{"payer", "month", "timestamps"})

				for i, v := range resp.Timestamps {
					var rows []string
					if i == 0 {
						rows = []string{resp.Id, resp.Month, v}
					} else {
						rows = []string{"", "", v}
					}

					table.Append(rows)
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&month, "month", time.Now().UTC().Format("200601"), "import month (UTC), fmt: yyyymm")
	return cmd
}
