package cmds

import (
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func IamUsersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iamuser",
		Short: "Subcommand for IAM user-related operations",
		Long:  `Subcommand for IAM user-related operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
