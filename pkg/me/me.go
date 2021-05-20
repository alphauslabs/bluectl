package me

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"

	"github.com/alphauslabs/blue-sdk-go/blue/v1"
	"github.com/alphauslabs/blue-sdk-go/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Me(loginUrl, clientId, clientSecret string) (string, error) {
	var opts []grpc.DialOption
	creds := credentials.NewTLS(&tls.Config{})
	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithPerRPCCredentials(
		session.NewRpcCredentials(session.RpcCredentialsInput{
			LoginUrl:     loginUrl,
			ClientId:     clientId,
			ClientSecret: clientSecret,
		}),
	))

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, session.BlueEndpoint, opts...)
	if err != nil {
		return "", fmt.Errorf("DialContext failed: %w", err)
	}

	defer conn.Close()
	client := blue.NewBlueClient(conn)
	resp, err := client.Me(ctx, &blue.MeRequest{})
	if err != nil {
		return "", fmt.Errorf("Me failed: %w", err)
	}

	b, _ := json.Marshal(resp)
	return string(b), nil
}
