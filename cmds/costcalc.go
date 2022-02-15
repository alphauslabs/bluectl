package cmds

import (
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func CostCalculationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c10s",
		Short: "Subcommand for calculations",
		Long:  `Subcommand for calculations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	// cmd.AddCommand(ListChannelsCmd())
	return cmd
}
