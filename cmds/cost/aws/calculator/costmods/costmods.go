package costmods

import (
	"context"
	"encoding/json"
	"fmt"
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

func CreateCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a cost modifier",
		Long:  `Create a cost modifier.`,
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

			if rawInput == "" {
				fnerr(fmt.Errorf("--raw-input is required for this cmd"))
			}

			var in cost.CreateCalculatorCostModifierRequest
			err := json.Unmarshal([]byte(rawInput), &in)
			if err != nil {
				fnerr(fmt.Errorf("--raw-input is invalid"))
			}

			in.Vendor = "aws"
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
			resp, err := client.CreateCalculatorCostModifier(ctx, &in)
			if err != nil {
				fnerr(err)
				return
			}

			logger.Info(resp)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Cost/Cost_CreateCalculatorCostModifier")
	return cmd
}

func DelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <id>",
		Short: "Delete a cost modifier",
		Long:  `Delete a cost modifier.`,
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
			_, err = client.DeleteCalculatorCostModifier(ctx, &cost.DeleteCalculatorCostModifierRequest{
				Vendor: "aws",
				Id:     args[0],
			})

			if err != nil {
				fnerr(err)
				return
			}

			logger.Infof("%v deleted", args[0])
		},
	}

	cmd.Flags().SortFlags = false
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
		CreateCmd(),
		DelCmd(),
	)

	return cmd
}
