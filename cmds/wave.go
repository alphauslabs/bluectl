package cmds

import (
	"os"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/cmds/wave"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func WaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wave",
		Short: "Wave-specific subcommands",
		Long:  `Wave-specific subcommands.`,
		Annotations: map[string]string{
			"loginurl":     session.LoginUrlWave,
			"loginurlbeta": session.LoginUrlWaveNext,
			"clientid":     params.WaveClientId,
			"clientsecret": params.WaveClientSecret,
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("check subcmds, run with -h")
			os.Exit(1)
		},
	}

	cmd.AddCommand(
		wave.MeCmd(),
	)

	cmd.Flags().SortFlags = false
	return cmd
}
