package wave

import (
	"os"

	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/me"
	"github.com/spf13/cobra"
)

func MeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get my information as a Wave user",
		Long:  `Get my information as a Wave user.`,
		Run: func(cmd *cobra.Command, args []string) {
			info, err := me.Me(
				cmd.Parent().Annotations["loginurl"],
				cmd.Parent().Annotations["clientid"],
				cmd.Parent().Annotations["clientsecret"],
			)

			if err != nil {
				logger.Error(err)
				os.Exit(1)
			}

			logger.Info(info)
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}
