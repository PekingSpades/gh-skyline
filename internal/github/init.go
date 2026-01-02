// Package github provides a function to initialize the GitHub client.
package github

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ClientInitializer is a function type for initializing GitHub clients
type ClientInitializer func() (*Client, error)

// InitializeGitHubClient is the default client initializer
var InitializeGitHubClient ClientInitializer = func() (*Client, error) {
	apiClient, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}
	return NewClient(apiClient), nil
}

// InitializeGitHubClientWithToken creates a GitHub client using the provided token.
// If token is empty, falls back to the default GitHub CLI authentication.
func InitializeGitHubClientWithToken(token string) (*Client, error) {
	if token == "" {
		return InitializeGitHubClient()
	}

	opts := api.ClientOptions{
		AuthToken: token,
	}
	apiClient, err := api.NewGraphQLClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client with token: %w", err)
	}
	return NewClient(apiClient), nil
}
