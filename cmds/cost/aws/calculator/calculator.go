package calculator

import (
	"github.com/alphauslabs/bluectl/cmds/cost/aws/calculator/costmods"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calculator",
		Short: "Cost calculator subcommands",
		Long:  `Cost calculator subcommands.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(costmods.Cmd())
	return cmd
}
