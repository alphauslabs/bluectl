package loginurl

import (
	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
)

func LoginUrl() string {
	if params.Target == "wave" {
		return session.LoginUrlWave
	}

	return session.LoginUrlRipple
}
