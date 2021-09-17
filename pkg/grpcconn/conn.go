package grpcconn

import (
	"context"
	"strings"

	"github.com/alphauslabs/blue-sdk-go/conn"
	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
)

const (
	bluesvc = "blue"

	OrgService     = bluesvc
	IamService     = bluesvc
	CostService    = "cost"
	BillingService = "billing"
	OpsService     = bluesvc
	PrefsService   = bluesvc
	KvStoreService = "kvstore"
)

func GetConnection(ctx context.Context, svcname string) (*conn.GrpcClientConn, error) {
	sess := session.New(
		session.WithLoginUrl(params.AuthUrl),
		session.WithClientId(params.ClientId),
		session.WithClientSecret(params.ClientSecret),
	)

	tgt := conn.BlueEndpoint
	if strings.Contains(params.AuthUrl, "next") {
		tgt = conn.BlueEndpointNext
	}

	return conn.New(ctx,
		conn.WithSession(sess),
		conn.WithTarget(tgt),
		conn.WithTargetService(svcname),
	)
}
