package session

import (
	"context"
	"fmt"
)

type tokenAuth struct {
	loginUrl     string
	clientId     string
	clientSecret string
	authType     string // default: Bearer
	accessToken  string // use directly if set
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	s := New(
		WithLoginUrl(t.loginUrl),
		WithClientId(t.clientId),
		WithClientSecret(t.clientSecret),
	)

	token := t.accessToken
	if token == "" {
		var err error
		token, err = s.AccessToken()
		if err != nil {
			return map[string]string{}, err
		}
	}

	authType := "Bearer"
	if t.authType != "" {
		authType = t.authType
	}

	ftoken := fmt.Sprintf("%s %s", authType, token)
	return map[string]string{"authorization": ftoken}, nil
}

func (tokenAuth) RequireTransportSecurity() bool { return false }

type RpcCredentialsInput struct {
	LoginUrl     string
	ClientId     string
	ClientSecret string
	AuthType     string // default: Bearer
	AccessToken  string // use directly if non-empty; disregard others
}

func NewRpcCredentials(in ...RpcCredentialsInput) tokenAuth {
	var authType, accessToken string
	sess := New()
	loginUrl := sess.LoginUrl()
	clientId := sess.ClientId()
	clientSecret := sess.ClientSecret()
	if len(in) > 0 {
		authType = in[0].AuthType
		accessToken = in[0].AccessToken
		if in[0].LoginUrl != "" {
			loginUrl = in[0].LoginUrl
		}

		if in[0].ClientId != "" {
			clientId = in[0].ClientId
		}

		if in[0].ClientSecret != "" {
			clientSecret = in[0].ClientSecret
		}
	}

	return tokenAuth{
		loginUrl:     loginUrl,
		clientId:     clientId,
		clientSecret: clientSecret,
		authType:     authType,
		accessToken:  accessToken,
	}
}
