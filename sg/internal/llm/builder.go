package llm

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// SourceBuilder constructs a collection of source readers.
type SourceBuilder struct {
	paths       []string
	contextRoot string
	err         error
}

// SourcesFromPath creates a SourceBuilder with loading sources from the given paths.
func SourcesFromPath(paths []string) *SourceBuilder {
	absolutePaths := make([]string, len(paths))
	for idx, path := range paths {
		p, err := filepath.Abs(path)
		if err != nil {
			return &SourceBuilder{err: err}
		}
		absolutePaths[idx] = p
	}

	return &SourceBuilder{
		paths: absolutePaths,
	}
}

// ContextRoot binds the context root to the SourceBuilder.
// When context root is specified, all paths must be relative to context root,
// and source names will be relative to context root.
func (sb *SourceBuilder) ContextRoot(contextRoot string) *SourceBuilder {
	if sb.err != nil {
		return sb
	}

	p, err := filepath.Abs(contextRoot)
	if err != nil {
		sb.err = err
		return sb
	}

	sb.contextRoot = p
	return sb
}

func (sb *SourceBuilder) Complete() ([]Source, error) {
	if sb.err != nil {
		return nil, sb.err
	}

	var rv []Source

	// load from paths
	{
		sources, err := loadSourceFromPaths(sb.contextRoot, sb.paths)
		if err != nil {
			return nil, err
		}
		rv = append(rv, sources...)
	}

	return rv, nil
}

func AzureOpenAIClientFromEnv() (*openai.Client, error) {
	const (
		envKeyAPIBaseURL = "OPENAI_API_BASE_URL"
		envKeyAPIVersion = "OPENAI_API_VERSION"
	)

	baseURL := os.Getenv(envKeyAPIBaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("environment variable %q is not set", envKeyAPIBaseURL)
	}
	apiVersion := os.Getenv(envKeyAPIVersion)
	if apiVersion == "" {
		apiVersion = "2024-02-15-preview"
	}

	identity, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Identity: %w", err)
	}

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithQuery("api-version", apiVersion),
		option.WithMiddleware(func(r *http.Request, next option.MiddlewareNext) (*http.Response, error) {
			// ref: https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/managed-identity#chat-completions
			token, err := identity.GetToken(r.Context(), policy.TokenRequestOptions{
				Scopes: []string{"https://cognitiveservices.azure.com/.default"},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to get Azure Identity token: %w", err)
			}

			r.Header.Set("Authorization", "Bearer "+token.Token)

			return next(r)
		}),
	)

	return client, nil
}
