package cmds

import (
	"os"

	"github.com/alphauslabs/bluectl/cmds/ripple"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func RippleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ripple",
		Short: "Ripple-specific subcommands",
		Long:  `Ripple-specific subcommands.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("check subcmds, run with -h")
			os.Exit(1)
		},
	}

	cmd.AddCommand(
		ripple.RippleTestCmd(),
	)

	return cmd
}
