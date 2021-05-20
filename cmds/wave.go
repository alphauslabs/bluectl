package cmds

import (
	"os"

	"github.com/alphauslabs/bluectl/cmds/wave"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func WaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wave",
		Short: "Wave-specific subcommands",
		Long:  `Wave-specific subcommands.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("check subcmds, run with -h")
			os.Exit(1)
		},
	}

	cmd.AddCommand(
		wave.WaveTestCmd(),
	)

	cmd.Flags().SortFlags = false
	return cmd
}
