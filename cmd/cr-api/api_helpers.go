package main

import (
	"errors"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/urfave/cli/v3"
)

const requiredAPITokenMessage = "API token is required. Set CLASH_ROYALE_API_TOKEN environment variable or use --api-token flag"

type apiClientOptions struct {
	offlineAllowed bool
	offlineHint    string
	missingToken   string
}

func requireAPIClient(cmd *cli.Command, opts apiClientOptions) (*clashroyale.Client, error) {
	return requireAPIClientFromToken(cmd.String("api-token"), opts)
}

func requireAPIClientFromToken(apiToken string, opts apiClientOptions) (*clashroyale.Client, error) {
	token, err := requireAPITokenValue(apiToken, opts)
	if err != nil {
		return nil, err
	}
	return clashroyale.NewClient(token), nil
}

func requireAPIToken(cmd *cli.Command, opts apiClientOptions) (string, error) {
	return requireAPITokenValue(cmd.String("api-token"), opts)
}

func requireAPITokenValue(apiToken string, opts apiClientOptions) (string, error) {
	if apiToken == "" {
		return "", errors.New(buildAPITokenRequiredMessage(opts))
	}
	return apiToken, nil
}

func buildAPITokenRequiredMessage(opts apiClientOptions) string {
	if opts.missingToken != "" {
		return opts.missingToken
	}
	if opts.offlineHint != "" {
		return requiredAPITokenMessage + opts.offlineHint
	}
	if opts.offlineAllowed {
		return requiredAPITokenMessage + ". Use --from-analysis for offline mode"
	}
	return requiredAPITokenMessage
}
