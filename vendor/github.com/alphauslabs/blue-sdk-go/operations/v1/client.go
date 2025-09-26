package operations

import (
	"context"

	"github.com/alphauslabs/blue-sdk-go/conn"
)

type GrpcClient struct {
	OperationsClient
	conn *conn.GrpcClientConn
}

func (c *GrpcClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// ClientOptions represents the optional options to NewClient.
type ClientOptions struct {
	Conn *conn.GrpcClientConn
}

// NewClient returns a client connection to the 'operations' service.
func NewClient(ctx context.Context, opts ...*ClientOptions) (*GrpcClient, error) {
	var fconn *conn.GrpcClientConn
	var err error
	switch {
	case len(opts) > 0:
		if opts[0].Conn != nil {
			fconn = opts[0].Conn
		}
	default:
		fconn, err = conn.New(ctx, conn.WithTargetService("blue"))
		if err != nil {
			return nil, err
		}
	}

	cc := NewOperationsClient(fconn)
	return &GrpcClient{cc, fconn}, nil
}
