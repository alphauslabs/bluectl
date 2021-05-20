package cmds

import (
	"os"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/cmds/ripple"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func RippleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ripple",
		Short: "Ripple-specific subcommands",
		Long:  `Ripple-specific subcommands.`,
		Annotations: map[string]string{
			"loginurl":     session.LoginUrlRipple,
			"loginurlbeta": session.LoginUrlRippleNext,
			"clientid":     params.RippleClientId,
			"clientsecret": params.RippleClientSecret,
		},
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("check subcmds, run with -h")
			os.Exit(1)
		},
	}

	cmd.AddCommand(
		ripple.MeCmd(),
		ripple.AccessTokenCmd(),
		ripple.AwsCostCmd(),
		ripple.AwsFeesCmd(),
	)

	cmd.Flags().SortFlags = false
	return cmd
}
