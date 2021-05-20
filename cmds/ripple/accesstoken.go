package ripple

import (
	"fmt"
	"log"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/spf13/cobra"
)

func AccessTokenCmd() *cobra.Command {
	var (
		beta     bool
		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "access-token",
		Short: "Get access token for Ripple",
		Long:  `Get access token for Ripple. See global flags for more information on the default environment variables.`,
		Run: func(cmd *cobra.Command, args []string) {
			var s *session.Session
			var o []session.Option
			if username != "" || password != "" {
				o = append(o, session.WithGrantType("password"))
				o = append(o, session.WithUsername(username))
				o = append(o, session.WithPassword(password))
			}

			switch {
			case beta:
				o = append(o, session.WithLoginUrl(cmd.Parent().Annotations["loginurlbeta"]))
			default:
				o = append(o, session.WithLoginUrl(cmd.Parent().Annotations["loginurl"]))
			}

			o = append(o, session.WithClientId(cmd.Parent().Annotations["clientid"]))
			o = append(o, session.WithClientSecret(cmd.Parent().Annotations["clientsecret"]))
			s = session.New(o...)
			token, err := s.AccessToken()
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Print(token)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVar(&beta, "beta", beta, "if true, access beta version (next)")
	cmd.Flags().StringVar(&username, "username", username, "if provided, 'password' grant type is implied")
	cmd.Flags().StringVar(&password, "password", password, "if provided, 'password' grant type is implied")
	return cmd
}
