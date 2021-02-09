package cmds

import (
	"fmt"
	"log"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/loginurl"
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
			s := session.New(
				session.WithLoginUrl(loginurl.LoginUrl()),
				session.WithClientId(params.ClientId),
				session.WithClientSecret(params.ClientSecret),
			)

			token, err := s.AccessToken()
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Print(token)
		},
	}

	c.Flags().SortFlags = false
	return c
}
