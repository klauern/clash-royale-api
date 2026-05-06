package main

import (
	"errors"
	"os"

	"github.com/klauer/clash-royale-api/go/pkg/clashroyale"
	"github.com/urfave/cli/v3"
)

// apiTokenEnvVar names the environment variable that holds the Clash Royale
// API token. Subcommand flag declarations should also wire `cli.EnvVars` to
// this name so urfave/cli auto-populates the flag — but resolveAPIToken
// always retries os.Getenv as a defensive fallback in case a particular
// subcommand was registered without `Sources`.
const apiTokenEnvVar = "CLASH_ROYALE_API_TOKEN"

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

// resolveAPIToken returns a non-empty API token sourced from the explicit
// argument first, then the CLASH_ROYALE_API_TOKEN env var. Returns "" when
// neither is set; callers should pair this with requireAPITokenValue when an
// empty token is an error.
func resolveAPIToken(apiToken string) string {
	if apiToken != "" {
		return apiToken
	}
	return os.Getenv(apiTokenEnvVar)
}

func requireAPITokenValue(apiToken string, opts apiClientOptions) (string, error) {
	resolved := resolveAPIToken(apiToken)
	if resolved == "" {
		return "", errors.New(buildAPITokenRequiredMessage(opts))
	}
	return resolved, nil
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
