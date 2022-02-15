package cmds

import (
	"fmt"
	"os"

	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
)

func AccessTokenCmd() *cobra.Command {
	var (
		beta     bool
		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Get access token for Ripple/Wave(Pro) authentication.",
		Long:  `Get access token for Ripple/Wave(Pro) authentication. See global flags for more information on the default environment variables.`,
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
				if params.AuthUrl == "" {
					params.AuthUrl = session.LoginUrlRippleNext
				}
			default:
				if params.AuthUrl == "" {
					params.AuthUrl = session.LoginUrlRipple
				}
			}

			o = append(o, session.WithLoginUrl(params.AuthUrl))
			o = append(o, session.WithClientId(params.ClientId))
			o = append(o, session.WithClientSecret(params.ClientSecret))
			s = session.New(o...)

			// Get actual access token.
			token, err := s.AccessToken()
			if err != nil {
				logger.Error(err)
				os.Exit(1)
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
