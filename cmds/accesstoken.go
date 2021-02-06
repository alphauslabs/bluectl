package cmds

import (
	"fmt"
	"os"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/logger"
	"github.com/alphauslabs/bluectl/params"
	"github.com/spf13/cobra"
)

func AccessTokenCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "access-token",
		Short: "Get access token for (Ripple/Wave)",
		Long: `Get access token for (Ripple/Wave). By default, it will look for the following
environment variables:

  ALPHAUS_CLIENT_ID
  ALPHAUS_CLIENT_SECRET`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			loginUrl := session.LoginUrlRipple
			if params.Target == "wave" {
				loginUrl = session.LoginUrlWave
			}

			s := session.New(
				session.WithLoginUrl(loginUrl),
				session.WithClientId(params.ClientId),
				session.WithClientSecret(params.ClientSecret),
			)

			token, err := s.AccessToken()
			if err != nil {
				logger.Error(err)
				ret = 1
				return
			}

			fmt.Print(token)
		},
	}

	c.Flags().SortFlags = false
	return c
}
