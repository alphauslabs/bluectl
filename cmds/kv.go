package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/alphauslabs/blue-sdk-go/kvstore/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func KvScanCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "scan [like]",
		Short: "Scan keys in your store",
		Long: `Scan keys in your store. If [like] is provided, it is translated as SQL's LIKE operator.
For example, 'scan %pattern%'. Return all keys by default.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.KvStoreService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := kvstore.NewClient(ctx, &kvstore.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			var stream kvstore.KvStore_ScanClient
			var f *os.File
			var wf *csv.Writer
			hdrs := []string{"KEY", "VALUE"}

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
				var in kvstore.ScanRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				stream, err = client.Scan(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				var in kvstore.ScanRequest
				if len(args) >= 1 {
					in.Like = args[0]
				}

				stream, err = client.Scan(ctx, &in)
				if err != nil {
					fnerr(err)
					return
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

				switch {
				case params.OutFile != "" && params.OutFmt == "csv":
					wf.Write([]string{v.Key, v.Value})
				case params.OutFmt == "json":
					b, _ := json.Marshal(v)
					fmt.Println(string(b))
				default:
					render = true
					table.Append([]string{v.Key, v.Value})
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
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/KvStore/KvStore_Scan")
	return cmd
}

func KvReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <key>",
		Short: "Read a key:value",
		Long:  `Read a key:value.`,
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
				fnerr(fmt.Errorf("<key> cannot be empty"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.KvStoreService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := kvstore.NewClient(ctx, &kvstore.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			resp, err := client.Read(ctx, &kvstore.ReadRequest{Key: args[0]})
			if err != nil {
				fnerr(err)
				return
			}

			switch {
			case params.OutFmt == "json":
				b, _ := json.Marshal(resp)
				logger.Info(string(b))
			default:
				logger.Info(resp.Value)
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func KvWriteCmd() *cobra.Command {
	var (
		fromFile string
	)

	cmd := &cobra.Command{
		Use:   "write <key> [value]",
		Short: "Write a new (or update an existing) key:value",
		Long:  `Write a new (or update an existing) key:value. If the --from-file is provided, [value] will be discarded.`,
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
				fnerr(fmt.Errorf("<key> cannot be empty"))
				return
			}

			var value string // empty allowed
			if len(args) >= 2 {
				value = args[1]
			}

			if fromFile != "" {
				b, err := ioutil.ReadFile(fromFile)
				if err != nil {
					fnerr(err)
					return
				}

				value = string(b)
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.KvStoreService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := kvstore.NewClient(ctx, &kvstore.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			_, err = client.Write(ctx, &kvstore.KeyValue{
				Key:   args[0],
				Value: value,
			})

			if err != nil {
				fnerr(err)
				return
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&fromFile, "from-file", fromFile, "path to file, use contents as value")
	return cmd
}

func KvDelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <key>",
		Short: "Delete a key:value",
		Long:  `Delete a key:value. To delete all your keys, use '-' as the input key.`,
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
				fnerr(fmt.Errorf("<key> cannot be empty"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.KvStoreService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := kvstore.NewClient(ctx, &kvstore.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			_, err = client.Delete(ctx, &kvstore.DeleteRequest{Key: args[0]})
			if err != nil {
				fnerr(err)
				return
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func KvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kv",
		Short: "Subcommand for KvStore operations",
		Long:  `Subcommand for KvStore operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		KvScanCmd(),
		KvReadCmd(),
		KvWriteCmd(),
		KvDelCmd(),
	)

	return cmd
}
