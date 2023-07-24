package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/alphauslabs/blue-sdk-go/operations/v1"
	"github.com/alphauslabs/blue-sdk-go/protos"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"google.golang.org/protobuf/types/known/durationpb"
)

type WaitForOperationInput struct {
	Name   string
	Client *operations.GrpcClient // optional
}

func WaitForOperation(ctx context.Context, in WaitForOperationInput) (*protos.Operation, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("in.Name cannot be empty")
	}

	var local bool
	client := in.Client
	if client == nil {
		ctx := context.Background()
		mycon, err := grpcconn.GetConnection(ctx, grpcconn.OpsService)
		if err != nil {
			return nil, err
		}

		client, err = operations.NewClient(ctx, &operations.ClientOptions{Conn: mycon})
		if err != nil {
			return nil, err
		}

		local = true
	}

	if local {
		defer client.Close()
	}

	type data struct {
		op  *protos.Operation
		err error
	}

	done := make(chan *data, 1)

	go func() {
		for {
			resp, err := client.WaitOperation(ctx, &operations.WaitOperationRequest{
				Name:    in.Name,
				Timeout: durationpb.New(time.Minute * 4),
			})

			if err != nil {
				done <- &data{err: err}
				return
			}

			if resp.Done {
				done <- &data{op: resp}
				return
			}
		}
	}()

	select {
	case d := <-done:
		if d.err != nil {
			return nil, d.err
		} else {
			return d.op, nil
		}
	case <-ctx.Done():
		return nil, nil
	}
}
