package ripple

import (
	"fmt"
	"log"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/spf13/cobra"
)

func AccessTokenCmd() *cobra.Command {
	var (
		beta bool
	)

	cmd := &cobra.Command{
		Use:   "access-token",
		Short: "Get access token for Ripple",
		Long: `Get access token for Ripple. By default, it will look for the following
environment variables:

  ALPHAUS_RIPPLE_CLIENT_ID
  ALPHAUS_RIPPLE_CLIENT_SECRET`,
		Run: func(cmd *cobra.Command, args []string) {
			var s *session.Session
			switch {
			case beta:
				s = session.New(
					session.WithLoginUrl(cmd.Parent().Annotations["loginurlbeta"]),
					session.WithClientId(cmd.Parent().Annotations["clientid"]),
					session.WithClientSecret(cmd.Parent().Annotations["clientsecret"]),
				)
			default:
				s = session.New(
					session.WithLoginUrl(cmd.Parent().Annotations["loginurl"]),
					session.WithClientId(cmd.Parent().Annotations["clientid"]),
					session.WithClientSecret(cmd.Parent().Annotations["clientsecret"]),
				)
			}

			token, err := s.AccessToken()
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Print(token)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVar(&beta, "beta", beta, "if true, access beta version (next)")
	return cmd
}
