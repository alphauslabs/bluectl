package cmds

import (
	"github.com/alphauslabs/bluectl/cmds/awsbilling"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func AwsBillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "awsbill",
		Short: "Subcommand for AWS billing-related operations",
		Long:  `Subcommand for AWS billing-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		awsbilling.CalculationHistoryCmd(),
	)

	return cmd
}
