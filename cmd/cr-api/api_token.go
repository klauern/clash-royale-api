package main

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

type apiTokenRequirement struct {
	Reason      string
	OfflineHint string
}

func requireAPIToken(cmd *cli.Command, req apiTokenRequirement) (string, error) {
	return requireAPITokenValue(cmd.String("api-token"), req)
}

func requireAPITokenValue(apiToken string, req apiTokenRequirement) (string, error) {
	if apiToken == "" {
		return "", fmt.Errorf(formatAPITokenRequired(req))
	}
	return apiToken, nil
}

func formatAPITokenRequired(req apiTokenRequirement) string {
	message := "API token is required"
	if req.Reason != "" {
		message = fmt.Sprintf("%s %s", message, req.Reason)
	}
	message += ". Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag"
	if req.OfflineHint != "" {
		message = fmt.Sprintf("%s. %s", message, req.OfflineHint)
	}
	return message
}
