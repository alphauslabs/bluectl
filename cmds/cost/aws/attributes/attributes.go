package attributes

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	awstypes "github.com/alphauslabs/blue-sdk-go/api/aws"
	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var (
		rawInput string
		colWidth int
	)

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get AWS cost attributes",
		Long: `Get AWS cost attributes. At the moment, we recommend you to use the --raw-input flag to take advantage
of the API's full features described in https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCostAttributes.
Note that this will invalidate all the other flags.`,
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
					wf.Write([]string{
						"groupId",
						"account",
						"productCode",
						"serviceCode",
						"region",
						"zone",
						"usageType",
						"instanceType",
						"operation",
						"invoiceId",
						"description",
						"resourceId",
						"tags",
						"costCategories",
					})
				case "json":
				default:
					fnerr(fmt.Errorf("unsupported output format"))
					return
				}
			}

			fnWriteFile := func(v *awstypes.CostAttribute) {
				b, _ := json.Marshal(v)
				fmt.Println(string(b))
				if params.OutFile != "" {
					var tags, cc string
					if v.Tags != nil {
						b, _ := json.Marshal(v.Tags)
						tags = string(b)
					}

					if v.CostCategories != nil {
						b, _ := json.Marshal(v.CostCategories)
						cc = string(b)
					}

					switch params.OutFmt {
					case "csv":
						wf.Write([]string{
							v.GroupId,
							v.Account,
							v.ProductCode,
							v.ServiceCode,
							v.Region,
							v.Zone,
							v.UsageType,
							v.InstanceType,
							v.Operation,
							v.InvoiceId,
							v.Description,
							v.ResourceId,
							tags,
							cc,
						})
					case "json":
						fmt.Fprintf(f, "%v\n", string(b))
					}
				}
			}

			type colT struct {
				name   string
				title  string
				enable bool
				val    interface{}
				vfmt   string
			}

			var stream cost.Cost_ReadCostAttributesClient
			var in cost.ReadCostAttributesRequest
			cols := []string{}
			refCols := []colT{
				{name: "account", title: "ACCOUNT", enable: true},
				{name: "productCode", title: "SERVICE"},
				{name: "serviceCode", title: "SERVICECODE"},
				{name: "region", title: "REGION"},
				{name: "zone", title: "ZONE"},
				{name: "usageType", title: "USAGE_TYPE"},
				{name: "instanceType", title: "INSTANCE_TYPE"},
				{name: "operation", title: "OPERATION"},
				{name: "invoiceId", title: "INVOICE_ID"},
				{name: "description", title: "DESCRIPTION"},
				{name: "resourceId", title: "RESOURCE_ID"},
				{name: "tags", title: "TAGS"},
				{name: "costCategories", title: "COST_CATEGORIES"},
			}

			switch {
			case rawInput != "":
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				if in.Vendor == "" {
					in.Vendor = "aws"
				}

				if in.AwsOptions != nil {
					if in.AwsOptions.Dimensions != "" {
						gbcs := strings.Split(in.AwsOptions.Dimensions, ",")
						for _, gbc := range gbcs {
							for i, rc := range refCols {
								if rc.name == gbc {
									refCols[i].enable = true
									break
								}
							}
						}
					} else {
						for i := range refCols {
							refCols[i].enable = true
						}
					}
				} else {
					for i := range refCols {
						refCols[i].enable = true
					}
				}

				stream, err = client.ReadCostAttributes(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				logger.Info("please use --raw-input for now, sorry")
				return
			}

			for _, v := range refCols {
				if v.enable {
					cols = append(cols, v.title)
				}
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetAutoFormatHeaders(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetColWidth(colWidth)
			table.SetBorder(false)
			table.SetHeaderLine(false)
			table.SetColumnSeparator("")
			table.SetTablePadding("  ")
			table.SetNoWhiteSpace(true)
			table.SetHeader(cols)
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
					fnWriteFile(v.Aws)
				default:
					render = true
					refCols[0].val = v.Aws.Account
					refCols[1].val = v.Aws.ProductCode
					refCols[2].val = v.Aws.ServiceCode
					refCols[3].val = v.Aws.Region
					refCols[4].val = v.Aws.Zone
					refCols[5].val = v.Aws.UsageType
					refCols[6].val = v.Aws.InstanceType
					refCols[7].val = v.Aws.Operation
					refCols[8].val = v.Aws.InvoiceId
					refCols[9].val = v.Aws.Description
					refCols[10].val = v.Aws.ResourceId
					refCols[11].val = v.Aws.Tags
					refCols[12].val = v.Aws.CostCategories
					row := []string{}
					for _, rc := range refCols {
						if rc.enable {
							if (rc.name == "tags" || rc.name == "costCategories") && rc.val != nil {
								ms := []string{}
								m := rc.val.(map[string]string)
								for k, v := range m {
									ms = append(ms, fmt.Sprintf("%v:%v", k, v))
								}

								jms := strings.Join(ms, ",")
								row = append(row, fmt.Sprintf("%v", jms))
							} else {
								row = append(row, fmt.Sprintf("%v", rc.val))
							}
						}
					}

					fmt.Printf("\033[2K\rrecv:%v...", row)
					table.Append(row)
				}
			}

			if render {
				fmt.Printf("\033[2K\r") // reset cursor
				table.Render()
			}

			if params.OutFile != "" {
				logger.Infof("data written to %v in %v format", params.OutFile, params.OutFmt)
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_ReadCostAttributes")
	cmd.Flags().IntVar(&colWidth, "col-width", 30, "set column width, applies to table-based outputs only")
	return cmd
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attributes",
		Short: "Cost attributes subcommand",
		Long:  `Cost attributes subcommand.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(GetCmd())
	return cmd
}
