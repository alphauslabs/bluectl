package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alphauslabs/blue-sdk-go/operations/v1"
	"github.com/alphauslabs/blue-sdk-go/session"
	"github.com/alphauslabs/bluectl/params"
	"github.com/alphauslabs/bluectl/pkg/grpcconn"
	"github.com/alphauslabs/bluectl/pkg/logger"
	"github.com/alphauslabs/bluectl/pkg/ops"
	"github.com/spf13/cobra"
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
				fnerr(err)
				return
			}

			client, err := operations.NewClient(ctx, &operations.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
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

			switch {
			case params.OutFmt == "json":
				// TODO: Support for accessing beta (next) environment.
				s := session.New(
					session.WithClientId(params.ClientId),
					session.WithClientSecret(params.ClientSecret),
				)

				t, err := s.AccessToken()
				if err != nil {
					fnerr(err)
					return
				}

				// Simpler to get the raw JSON this way.
				hc := &http.Client{Timeout: 60 * time.Second}
				u := fmt.Sprintf("https://api.alphaus.cloud/m/blue/ops/v1/%v", args[0])
				r, err := http.NewRequest(http.MethodGet, u, nil)
				if err != nil {
					fnerr(err)
					return
				}

				r.Header.Add("Authorization", "Bearer "+t)
				resp, err := hc.Do(r)
				if err != nil {
					fnerr(err)
					return
				}

				if (resp.StatusCode / 100) != 2 {
					fnerr(fmt.Errorf(resp.Status))
					return
				}

				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fnerr(err)
					return
				}

				logger.Info(string(body))
			default:
				ctx := context.Background()
				mycon, err := grpcconn.GetConnection(ctx, grpcconn.OpsService)
				if err != nil {
					fnerr(err)
					return
				}

				client, err := operations.NewClient(ctx, &operations.ClientOptions{Conn: mycon})
				if err != nil {
					fnerr(err)
					return
				}

				defer client.Close()
				resp, err := client.GetOperation(ctx, &operations.GetOperationRequest{
					Name: args[0],
				})

				if err != nil {
					fnerr(err)
					return
				}

				logger.Info(resp)
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func OpsWaitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait <name>",
		Short: "Wait for a long-running operation to finish",
		Long:  `Wait for a long-running operation to finish.`,
		Run: func(cmd *cobra.Command, args []string) {
			defer func(begin time.Time) {
				logger.Info("duration:", time.Since(begin))
			}(time.Now())

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

			quit, cancel := context.WithCancel(context.Background())
			done := make(chan struct{}, 1)

			// Interrupt handler.
			go func() {
				sigch := make(chan os.Signal)
				signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
				<-sigch
				cancel()
			}()

			go func() {
				for {
					q := context.WithValue(quit, struct{}{}, nil)
					op, err := ops.WaitForOperation(q, ops.WaitForOperationInput{
						Name: args[0],
					})

					if err != nil {
						logger.Error(err)
						done <- struct{}{}
						return
					}

					if op != nil {
						if op.Done {
							logger.Infof("[%v] done", args[0])
							done <- struct{}{}
							return
						}
					}
				}
			}()

			logger.Infof("wait for [%v], this could take some time...", args[0])

			select {
			case <-done:
			case <-quit.Done():
				logger.Info("interrupted")
			}
		},
	}

	cmd.Flags().SortFlags = false
	return cmd
}

func OpsDelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Delete a long-running operation",
		Long:  `Delete a long-running operation.`,
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
			mycon, err := grpcconn.GetConnection(ctx, grpcconn.OpsService)
			if err != nil {
				fnerr(err)
				return
			}

			client, err := operations.NewClient(ctx, &operations.ClientOptions{Conn: mycon})
			if err != nil {
				fnerr(err)
				return
			}

			defer client.Close()
			_, err = client.DeleteOperation(ctx, &operations.DeleteOperationRequest{
				Name: args[0],
			})

			if err != nil {
				fnerr(err)
				return
			}

			logger.Infof("deleted: %v", args[0])
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
		OpsWaitCmd(),
		OpsDelCmd(),
	)

	return cmd
}
