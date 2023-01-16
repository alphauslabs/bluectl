package cost

import (
	"github.com/alphauslabs/bluectl/cmds/cost/aws"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func CostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost",
		Short: "Subcommand for cost-related operations",
		Long:  `Subcommand for cost-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(aws.Cmd())
	return cmd
}
