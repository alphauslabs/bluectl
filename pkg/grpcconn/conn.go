package grpcconn

import (
	"context"
	"strings"

	"github.com/alphauslabs/blue-sdk-go/conn"
	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
)

func GetConnection(ctx context.Context, svcname string) (*conn.GrpcClientConn, error) {
	sess := session.New(
		session.WithLoginUrl(params.AuthUrl),
		session.WithClientId(params.ClientId),
		session.WithClientSecret(params.ClientSecret),
	)

	tgt := session.BlueEndpoint
	if strings.Contains(params.AuthUrl, "next") {
		tgt = session.BlueEndpointNext
	}

	return conn.New(ctx,
		conn.WithSession(sess),
		conn.WithTarget(tgt),
		conn.WithTargetService(svcname),
	)
}
