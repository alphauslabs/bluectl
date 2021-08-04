package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alphauslabs/blue-sdk-go/api"
	"github.com/alphauslabs/blue-sdk-go/operations/v1"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func OpsListCmd() *cobra.Command {
	var (
		rawInput string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List long-running operations",
		Long:  `List long-running operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.OpsService)
			if err != nil {
				logger.Error(err)
				return
			}

			client, err := operations.NewClient(ctx, &operations.ClientOptions{Conn: mycon})
			if err != nil {
				logger.Error(err)
				return
			}

			defer client.Close()
			var stream operations.Operations_ListOperationsClient

			switch {
			case rawInput != "":
				var in operations.ListOperationsRequest
				err := json.Unmarshal([]byte(rawInput), &in)
				if err != nil {
					fnerr(err)
					return
				}

				stream, err = client.ListOperations(ctx, &in)
				if err != nil {
					fnerr(err)
					return
				}
			default:
				stream, err = client.ListOperations(ctx, &operations.ListOperationsRequest{})
				if err != nil {
					fnerr(err)
					return
				}
			}

			for {
				v, err := stream.Recv()
				if err == io.EOF {
					break
				}

				if err != nil {
					fnerr(err)
					return
				}

				b, _ := json.Marshal(v)
				logger.Info(string(b))
			}
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVar(&rawInput, "raw-input", rawInput, "raw JSON input; see https://alphauslabs.github.io/blueapidocs/#/Operations/Operations_ListOperations")
	return cmd
}

func OpsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <name>",
		Short: "Query a long-running operation",
		Long:  `Query a long-running operation.`,
		Run: func(cmd *cobra.Command, args []string) {
			var ret int
			defer func(r *int) {
				if *r != 0 {
					os.Exit(*r)
				}
			}(&ret)

			fnerr := func(e error) {
				logger.Error(e)
				ret = 1
			}

			if len(args) == 0 {
				fnerr(fmt.Errorf("<name> cannot be empty"))
				return
			}

			ctx := context.Background()
			mycon, err := grpcconn.GetConnection(ctx, "blue")
			if err != nil {
				logger.Error(err)
				return
			}

			client, err := operations.NewClient(ctx, &operations.ClientOptions{Conn: mycon})
			if err != nil {
				logger.Error(err)
				return
			}

			defer client.Close()
			resp, err := client.GetOperation(ctx, &operations.GetOperationRequest{
				Name: args[0],
			})

			if err != nil {
				logger.Error(err)
				return
			}

			logger.Info("name:", resp.Name)
			var meta api.OperationAwsCalculateCostsMetadataV1
			anypb.UnmarshalTo(resp.Metadata, &meta, proto.UnmarshalOptions{})
			sm, _ := json.Marshal(meta)
			logger.Info("metadata:", string(sm))
			logger.Info("done:", resp.Done)
			if v, ok := resp.Result.(*api.Operation_Response); ok {
				var r api.KeyValue
				anypb.UnmarshalTo(v.Response, &r, proto.UnmarshalOptions{})
				sr, _ := json.Marshal(r)
				logger.Info("response:", string(sr))
			}

			if v, ok := resp.Result.(*api.Operation_Error); ok {
				logger.Info("error:", v)
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func OpsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ops",
		Short: "Subcommand for long-running operations",
		Long:  `Subcommand for long-running operations.`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("see -h for more information")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.AddCommand(
		OpsListCmd(),
		OpsGetCmd(),
	)

	return cmd
}
