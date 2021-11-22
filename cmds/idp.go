package cmds

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alphauslabs/blue-sdk-go/iam/v1"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func ListIdpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IdPs",
		Long:  `List IdPs.`,
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
			resp, err := client.ListIdentityProviders(ctx, &iam.ListIdentityProvidersRequest{})
			if err != nil {
				fnerr(err)
				return
			}

			hdrs := []string{"ID", "NAME", "TYPE"}

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
					for _, d := range resp.Data {
						row := []string{d.Id, d.Name, d.Type}
						logger.Infof("%v --> %v", row, params.OutFile)
						wf.Write(row)
					}
				}
			case params.OutFmt == "json":
				for _, d := range resp.Data {
					b, _ := json.Marshal(d)
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

				for _, d := range resp.Data {
					table.Append([]string{d.Id, d.Name, d.Type})
				}

				table.Render()
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func CreateIdpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name> <path-to-metadata-file>",
		Short: "Create a new IdP entry",
		Long:  `Create a new IdP entry.`,
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

			if len(args) < 2 {
				fnerr(fmt.Errorf("name and metadata file required"))
				return
			}

			meta, err := ioutil.ReadFile(args[1])
			if err != nil {
				fnerr(err)
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
			_, err = client.CreateIdentityProvider(ctx, &iam.CreateIdentityProviderRequest{
				Name:     args[0],
				Type:     "saml",
				Metadata: string(meta),
			})

			if err != nil {
				fnerr(err)
				return
			}

			logger.Infof("IdP %v created.", args[0])
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func DelIdpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <id>",
		Short: "Delete IdP",
		Long:  `Delete IdP entry.`,
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
			_, err = client.DeleteIdentityProvider(ctx, &iam.DeleteIdentityProviderRequest{Id: args[0]})
			if err != nil {
				fnerr(err)
				return
			}

			logger.Infof("deleted: %v", args[0])
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func IdpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "idp",
		Short: "Subcommand for Identity Provider-related operations",
		Long:  `Subcommand for Identity Provider-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		ListIdpsCmd(),
		CreateIdpsCmd(),
		DelIdpCmd(),
	)

	return cmd
}
