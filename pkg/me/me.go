package me

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alphauslabs/blue-sdk-go/iam/v1"
)

func WhoAmI(loginUrl, clientId, clientSecret string) (string, error) {
	ctx := context.Background()
	client, err := iam.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("NewClient failed: %w", err)
	}

	defer client.Close()
	resp, err := client.WhoAmI(ctx, &iam.WhoAmIRequest{})
	if err != nil {
		return "", fmt.Errorf("WhoAmI failed: %w", err)
	}

	b, _ := json.Marshal(resp)
	return string(b), nil
}
