package ripple

import (
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func RippleTestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test",
		Long:  `Test here.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("Hello")
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
