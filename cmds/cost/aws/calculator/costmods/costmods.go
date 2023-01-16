package costmods

import (
	"context"
	"io"
	"os"

	"github.com/alphauslabs/blue-sdk-go/cost/v1"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	var (
		rawInput string
		colWidth int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cost modifiers",
		Long:  `List cost modifiers.`,
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
			stream, err := client.ListCalculatorCostModifiers(ctx, &cost.ListCalculatorCostModifiersRequest{
				Vendor: "aws",
			})

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

				logger.Info(v)
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
		Use:   "mods",
		Short: "Cost modifiers subcommand",
		Long:  `Cost modifiers subcommand.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		ListCmd(),
	)

	return cmd
}
