package conn

import (
	"context"
	"crypto/tls"

	"github.com/alphauslabs/blue-sdk-go/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	BlueEndpoint = "blue.alphaus.cloud:443" // global LB
	// BlueEndpoint     = "bluerpc.alphaus.cloud:8443" // network LB (JP)
	BlueEndpointNext = "bluenext.alphaus.cloud:8443"
)

type clientOptions struct {
	target string // gRPC server address
	sess   *session.Session
	conn   *grpc.ClientConn
	svc    string // our target service
}

type ClientOption interface {
	apply(*clientOptions)
}

// fnClientOption wraps a function that modifies clientOptions into an
// implementation of the ClientOption interface.
type fnClientOption struct {
	f func(*clientOptions)
}

func (o *fnClientOption) apply(do *clientOptions) { o.f(do) }

func newFnClientOption(f func(*clientOptions)) *fnClientOption {
	return &fnClientOption{f: f}
}

func WithTarget(s string) ClientOption {
	return newFnClientOption(func(o *clientOptions) {
		o.target = s
	})
}

func WithTargetService(s string) ClientOption {
	return newFnClientOption(func(o *clientOptions) {
		o.svc = s
	})
}

func WithSession(v *session.Session) ClientOption {
	return newFnClientOption(func(o *clientOptions) {
		o.sess = v
	})
}

func WithGrpcConnection(v *grpc.ClientConn) ClientOption {
	return newFnClientOption(func(o *clientOptions) {
		o.conn = v
	})
}

type GrpcClientConn struct {
	*grpc.ClientConn
	opts clientOptions
}

// Close closes the underlying connection.
func (c *GrpcClientConn) Close() {
	if c.opts.conn != nil {
		c.opts.conn.Close()
	}
}

// New returns a grpc connection to a Blue API target service.
func New(ctx context.Context, opts ...ClientOption) (*GrpcClientConn, error) {
	sess := session.New()
	co := clientOptions{
		target: BlueEndpoint,
		sess:   sess,
	}

	for _, opt := range opts {
		opt.apply(&co)
	}

	if co.conn == nil {
		var err error
		var gopts []grpc.DialOption
		creds := credentials.NewTLS(&tls.Config{})
		gopts = append(gopts, grpc.WithTransportCredentials(creds))
		gopts = append(gopts, grpc.WithPerRPCCredentials(
			session.NewRpcCredentials(session.RpcCredentialsInput{
				LoginUrl:     co.sess.LoginUrl(),
				ClientId:     co.sess.ClientId(),
				ClientSecret: co.sess.ClientSecret(),
			}),
		))

		if co.svc != "" {
			gopts = append(gopts, grpc.WithUnaryInterceptor(func(ctx context.Context,
				method string, req interface{}, reply interface{}, cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
			) error {
				ctx = metadata.AppendToOutgoingContext(ctx, "service-name", co.svc)
				ctx = metadata.AppendToOutgoingContext(ctx, "x-agent", "blue-sdk-go")
				return invoker(ctx, method, req, reply, cc, opts...)
			}))

			gopts = append(gopts, grpc.WithStreamInterceptor(func(ctx context.Context,
				desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
				streamer grpc.Streamer, opts ...grpc.CallOption,
			) (grpc.ClientStream, error) {
				ctx = metadata.AppendToOutgoingContext(ctx, "service-name", co.svc)
				ctx = metadata.AppendToOutgoingContext(ctx, "x-agent", "blue-sdk-go")
				return streamer(ctx, desc, cc, method, opts...)
			}))
		}

		co.conn, err = grpc.NewClient(co.target, gopts...)
		if err != nil {
			return nil, err
		}
	}

	return &GrpcClientConn{co.conn, co}, nil
}
