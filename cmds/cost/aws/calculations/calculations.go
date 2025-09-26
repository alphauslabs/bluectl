package calculations

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	protosinternal "github.com/alphauslabs/blue-internal-go/protos"
	"github.com/alphauslabs/blue-sdk-go/api"
	"github.com/alphauslabs/blue-sdk-go/billing/v1"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/cmds/cost/aws/calculations/schedule"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func RunCmd() *cobra.Command {
	var (
		rawInput string
		wait     bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Trigger an ondemand AWS costs calculation",
		Long:  `Trigger an ondemand AWS costs calculation.`,
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
			var resp *protosinternal.Operation

			switch {
			case rawInput != "":
				var in cost.CalculateCostsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				resp, err = client.CalculateCosts(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				resp, err = client.CalculateCosts(ctx, &cost.CalculateCostsRequest{
					Vendor: "aws",
				})

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
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_CalculateCosts")
	cmd.Flags().BoolVar(&wait, "wait", wait, "if true, wait for the operation to finish")
	return cmd
}

func ListRunningCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-running [month]",
		Short: "List AWS accounts that are still processing",
		Long: `List AWS accounts that are still processing. The format for [month] is yyyymm.
If [month] is not provided, it defaults to the current UTC month.`,
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

			mm := time.Now().UTC().Format("200601")
			if len(args) > 0 {
				_, err := time.Parse("200601", args[0])
				if err != nil {
					fnerr(err)
				} else {
					mm = args[0]
				}
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
					wf.Write([]string{"month", "account", "date", "started"})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWrite := func(v *cost.ListCalculatorRunningAccountsResponse) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.Aws.Month,
							v.Aws.Account,
							v.Aws.Date,
							v.Aws.Started,
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			stream, err := client.ListCalculatorRunningAccounts(ctx,
				&cost.ListCalculatorRunningAccountsRequest{
					Vendor: "aws",
					Month:  mm,
				},
			)

			if err != nil {
				fnerr(err)
				return
			}

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
			table.SetHeader([]string{"MONTH", "ACCOUNT", "DATE", "STARTED"})
			var render bool

			for {
				v, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err != nil {
					fnerr(err)
					return
				}

				switch {
				case params.OutFile != "":
					fnWrite(v)
				default:
					render = true
					row := []string{
						v.Aws.Month,
						v.Aws.Account,
						v.Aws.Date,
						v.Aws.Started,
					}

					fmt.Printf("\033[2K\rrecv:%v...", row)
					table.Append(row)
				}
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}

			if render {
				fmt.Printf("\033[2K\r") // reset cursor
				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func ListHistoryCmd() *cobra.Command {
	var (
		rawInput string
	)

	// Same copy in backend
	type CalculateCostsMeta struct {
		OrgId    string   `json:"orgId"`
		Month    string   `json:"month"`
		Status   string   `json:"status"`
		GroupIds []string `json:"groupIds"`
		Created  string   `json:"created"`
		Updated  string   `json:"updated"`
	}

	cmd := &cobra.Command{
		Use:   "list-history",
		Short: "Query AWS calculation history",
		Long:  `Query AWS calculation history.`,
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
			hdrs := []string{"NAME", "MONTH", "GROUPS", "UPDATED", "CREATED", "STATUS", "DONE", "RESULT"}
			var resp *cost.ListCalculationsHistoryResponse

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
				var in cost.ListCalculationsHistoryRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				resp, err = client.ListCalculationsHistory(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				resp, err = client.ListCalculationsHistory(ctx, &cost.ListCalculationsHistoryRequest{
					Vendor: "aws",
				})

				if err != nil {
					fnerr(err)
					return
				}
			}

			for _, op := range resp.Aws.Operations {
				var sm structpb.Struct
				op.Metadata.UnmarshalTo(&sm)
				meta := sm.AsMap()
				var result string
				switch op.Result.(type) {
				case *protosinternal.Operation_Response:
					result = "success"
				case *protosinternal.Operation_Error:
					terr := op.Result.(*protosinternal.Operation_Error)
					result = terr.Error.String()
				}

				switch {
				case params.OutFmt == "csv" && params.OutFile != "":
					b, _ := json.Marshal(meta)
					var cm CalculateCostsMeta
					json.Unmarshal(b, &cm)
					wf.Write([]string{
						op.Name,
						cm.Month,
						strings.Join(cm.GroupIds, ","),
						cm.Updated,
						cm.Created,
						cm.Status,
						fmt.Sprintf("%v", op.Done),
						result,
					})
				case params.OutFmt == "json":
					var m map[string]interface{}
					b, _ := json.Marshal(op)
					json.Unmarshal(b, &m)

					// Make metadata more readable.
					v := m["metadata"]
					vv := v.(map[string]interface{})
					vv["value"] = meta

					// Make result more readable.
					switch op.Result.(type) {
					case *protosinternal.Operation_Response:
						var res api.KeyValue
						tres := op.Result.(*protosinternal.Operation_Response)
						anypb.UnmarshalTo(tres.Response, &res, proto.UnmarshalOptions{})
						v := m["Result"]
						vv := v.(map[string]interface{})
						vvv := vv["Response"].(map[string]interface{})
						vvv["value"] = res
					}

					b, _ = json.Marshal(m)
					fmt.Println(string(b))
				default:
					render = true
					b, _ := json.Marshal(meta)
					var cm CalculateCostsMeta
					json.Unmarshal(b, &cm)
					table.Append([]string{
						op.Name,
						cm.Month,
						strings.Join(cm.GroupIds, ","),
						cm.Updated,
						cm.Created,
						cm.Status,
						fmt.Sprintf("%v", op.Done),
						result,
					})
				}
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
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ListCalculationsHistory")
	return cmd
}

func ListDailyRunHistoryCmd() *cobra.Command {
	var (
		red   = color.New(color.FgRed).SprintFunc()
		month string
	)

	cmd := &cobra.Command{
		Use:   "list-runhistory [yyyymm] [billingInternalId]",
		Short: "Query AWS daily run history for all accounts",
		Long: `Query AWS daily run history for all accounts. The default output format is:

billingInternalId/billingGroupId (yyyymm):
  accountId: timestamp=timestamp, trigger='cur|invoice[, after=true]'

Timestamps are ordered with the topmost as most recent. 'cur'-triggered means this calculation was
triggered by updates to the CUR while 'invoice' means by a manual invoice request.`,
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

			month = time.Now().UTC().Format("200601")
			if len(args) > 0 {
				mm, err := time.Parse("200601", args[0])
				if err != nil {
					fnerr(err)
					return
				} else {
					month = mm.Format("200601")
				}
			}

			var groupId string
			if len(args) >= 2 {
				groupId = args[1]
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.BillingService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := billing.NewClient(ctx, &billing.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			stream, err := client.ListAwsDailyRunHistory(ctx, &billing.ListAwsDailyRunHistoryRequest{
				Month:   month,
				GroupId: groupId,
			})

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

					wf.Write([]string{
						"billingInternalId",
						"billingGroupId",
						"month",
						"account",
						"timestamp",
						"trigger",
						"after",
					})

					for {
						v, err := stream.Recv()
						if err == io.EOF {
							break
						}

						if err != nil {
							fnerr(err)
							return
						}

						if len(v.Accounts) == 0 {
							continue
						}

						for _, acct := range v.Accounts {
							if len(acct.History) > 0 {
								var itr int
								var updated bool // after invoice
								for _, h := range acct.History {
									itr++
									if h.Trigger == "invoice" {
										if itr > 1 {
											updated = true
										}
										break
									}
								}

								for _, h := range acct.History {
									if updated && h.Trigger == "invoice" {
										updated = false
									}

									var row []string
									if updated && h.Trigger != "invoice" {
										row = []string{
											v.BillingInternalId,
											v.BillingGroupId,
											v.Month,
											acct.AccountId,
											h.Timestamp,
											h.Trigger,
											"yes",
										}
									} else {
										row = []string{
											v.BillingInternalId,
											v.BillingGroupId,
											v.Month,
											acct.AccountId,
											h.Timestamp,
											h.Trigger,
											"",
										}
									}

									logger.Infof("%v --> %v", row, params.OutFile)
									wf.Write(row)
								}
							}
						}
					}
				}
			case params.OutFmt == "json":
				logger.Info("format not supported yet")
			default:
				for {
					v, err := stream.Recv()
					if err == io.EOF {
						break
					}

					if err != nil {
						fnerr(err)
						return
					}

					if len(v.Accounts) == 0 {
						continue
					}

					fmt.Printf("%v/%v (%v)\n", v.BillingInternalId, v.BillingGroupId, v.Month)
					for _, acct := range v.Accounts {
						if len(acct.History) > 0 {
							var itr int
							var updated bool // after invoice
							for _, h := range acct.History {
								itr++
								if h.Trigger == "invoice" {
									if itr > 1 {
										updated = true
									}
									break
								}
							}

							for _, h := range acct.History {
								if updated && h.Trigger == "invoice" {
									updated = false
								}

								if updated && h.Trigger != "invoice" {
									fmt.Printf(red("  %v: timestamp=%v, trigger=%v, after=true\n"),
										acct.AccountId, h.Timestamp, h.Trigger)
								} else {
									fmt.Printf("  %v: timestamp=%v, trigger=%v\n",
										acct.AccountId, h.Timestamp, h.Trigger)
								}
							}
						}
					}
				}
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calculation",
		Short: "Calculations subcommand",
		Long:  `Calculations subcommand.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		RunCmd(),
		ListRunningCmd(),
		ListHistoryCmd(),
		ListDailyRunHistoryCmd(),
		schedule.Cmd(),
	)

	return cmd
}
