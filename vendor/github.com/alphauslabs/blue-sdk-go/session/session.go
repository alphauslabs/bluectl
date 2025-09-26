package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	LoginUrlRipple     = "https://login.alphaus.cloud/ripple/access_token"
	LoginUrlWave       = "https://login.alphaus.cloud/access_token"
	LoginUrlRippleNext = "https://loginnext.alphaus.cloud/ripple/access_token"
	LoginUrlWaveNext   = "https://loginnext.alphaus.cloud/access_token"
)

type Option interface {
	Apply(*Session)
}

type withClientId string

func (w withClientId) Apply(s *Session) { s.clientId = string(w) }
func WithClientId(v string) Option      { return withClientId(v) }

type withClientSecret string

func (w withClientSecret) Apply(s *Session) { s.clientSecret = string(w) }
func WithClientSecret(v string) Option      { return withClientSecret(v) }

type withGrantType string

func (w withGrantType) Apply(s *Session) { s.grantType = string(w) }
func WithGrantType(v string) Option      { return withGrantType(v) }

type withScope string

func (w withScope) Apply(s *Session) { s.scope = string(w) }
func WithScope(v string) Option      { return withScope(v) }

type withUsername string

func (w withUsername) Apply(s *Session) { s.username = string(w) }
func WithUsername(v string) Option      { return withUsername(v) }

type withPassword string

func (w withPassword) Apply(s *Session) { s.password = string(w) }
func WithPassword(v string) Option      { return withPassword(v) }

type withLoginUrl string

func (w withLoginUrl) Apply(s *Session) { s.loginUrl = string(w) }
func WithLoginUrl(v string) Option      { return withLoginUrl(v) }

type withHttpClient struct{ client *http.Client }

func (w withHttpClient) Apply(s *Session)  { s.httpClient = w.client }
func WithHttpClient(v *http.Client) Option { return withHttpClient{v} }

type Session struct {
	clientId     string
	clientSecret string
	grantType    string
	scope        string
	username     string
	password     string
	loginUrl     string
	httpClient   *http.Client
}

func (s *Session) ClientId() string     { return s.clientId }
func (s *Session) ClientSecret() string { return s.clientSecret }
func (s *Session) GrantType() string    { return s.grantType }
func (s *Session) Scope() string        { return s.scope }
func (s *Session) Username() string     { return s.username }
func (s *Session) Password() string     { return s.password }
func (s *Session) LoginUrl() string     { return s.loginUrl }

// AccessToken returns the access token after successful authentication to Blue API.
func (s *Session) AccessToken() (string, error) {
	var err error
	var token string
	var body []byte

	form := url.Values{}
	form.Add("client_id", s.ClientId())
	form.Add("client_secret", s.ClientSecret())
	form.Add("grant_type", s.GrantType())
	form.Add("scope", s.Scope())
	if s.GrantType() == "password" {
		form.Add("username", s.Username())
		form.Add("password", s.Password())
	}

	var resp *http.Response
	switch {
	case s.httpClient != nil:
		resp, err = s.httpClient.PostForm(s.LoginUrl(), form)
		if err != nil {
			return token, err
		}
	default:
		httpClient := &http.Client{Timeout: 60 * time.Second}
		resp, err = httpClient.PostForm(s.LoginUrl(), form)
		if err != nil {
			return token, err
		}
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if (resp.StatusCode / 100) != 2 {
		return token, fmt.Errorf(resp.Status)
	}

	var m map[string]interface{}
	if err = json.Unmarshal(body, &m); err != nil {
		return token, err
	}

	t, found := m["access_token"]
	if !found {
		return token, fmt.Errorf("cannot find access token")
	}

	token = fmt.Sprintf("%s", t)
	return token, nil
}

// New returns a Session object for Blue API authentication.
func New(o ...Option) *Session {
	id, secret, user, pass, loginUrl := GetLocalCreds()
	gt := "client_credentials"
	if user != "" && pass != "" {
		gt = "password"
	}

	s := &Session{
		loginUrl:     loginUrl,
		clientId:     id,
		clientSecret: secret,
		grantType:    gt,
		scope:        "openid",
		username:     user,
		password:     pass,
	}

	for _, opt := range o {
		opt.Apply(s)
	}

	return s
}

// GetLocalCreds returns caller's id, secret, user, password, and login url.
func GetLocalCreds() (string, string, string, string, string) {
	// Default environment variables.
	id := os.Getenv("ALPHAUS_CLIENT_ID")
	secret := os.Getenv("ALPHAUS_CLIENT_SECRET")
	user := os.Getenv("ALPHAUS_USERNAME")
	pass := os.Getenv("ALPHAUS_PASSWORD")
	loginUrl := os.Getenv("ALPHAUS_AUTH_URL")
	if loginUrl == "" {
		loginUrl = LoginUrlRipple
	}

	func() {
		if id != "" && secret != "" {
			return
		}

		// Then Ripple environment variables.
		id = os.Getenv("ALPHAUS_RIPPLE_CLIENT_ID")
		secret = os.Getenv("ALPHAUS_RIPPLE_CLIENT_SECRET")
		if id != "" && secret != "" {
			if loginUrl == "" {
				loginUrl = LoginUrlRipple
			}

			user = os.Getenv("ALPHAUS_RIPPLE_USERNAME")
			pass = os.Getenv("ALPHAUS_RIPPLE_PASSWORD")
			return
		}

		// Finally, Wave environment variables.
		id = os.Getenv("ALPHAUS_WAVE_CLIENT_ID")
		secret = os.Getenv("ALPHAUS_WAVE_CLIENT_SECRET")
		if id != "" && secret != "" {
			if loginUrl == "" {
				loginUrl = LoginUrlWave
			}

			user = os.Getenv("ALPHAUS_WAVE_USERNAME")
			pass = os.Getenv("ALPHAUS_WAVE_PASSWORD")
			loginUrl = LoginUrlWave
			return
		}
	}()

	return id, secret, user, pass, loginUrl
}
