package execution

import (
	"context"
	"fmt"
	"os"

	anyllmConfig "github.com/mozilla-ai/any-llm-go/config"
	anyllm "github.com/mozilla-ai/any-llm-go/providers"
	anyllmAnthropic "github.com/mozilla-ai/any-llm-go/providers/anthropic"
	anyllmGemini "github.com/mozilla-ai/any-llm-go/providers/gemini"
	anyllmOpenAI "github.com/mozilla-ai/any-llm-go/providers/openai"
)

// Env var names.
const (
	envProvider       = "TROUPE_PROVIDER"
	envUseVertex      = "CLAUDE_CODE_USE_VERTEX"
	envVertexRegion   = "CLOUD_ML_REGION"
	envVertexProject  = "ANTHROPIC_VERTEX_PROJECT_ID"
	envAnthropicKey   = "ANTHROPIC_API_KEY"
	envOpenAIKey      = "OPENAI_API_KEY"
	envGeminiKey      = "GEMINI_API_KEY"
	envOpenRouterKey  = "OPENROUTER_API_KEY"
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

// NewProviderFromEnv creates the LLM provider specified by the given env var.
// Consumers pass their own env var name: "DJINN_PROVIDER", "ORIGAMI_PROVIDER", etc.
// If envName is empty, defaults to TROUPE_PROVIDER.
// Explicit only — no auto-detection, no fallback, no magic.
func NewProviderFromEnv(envName string) (anyllm.Provider, error) {
	if envName == "" {
		envName = envProvider
	}
	name := os.Getenv(envName)
	if name == "" {
		return nil, fmt.Errorf("%s not set (options: vertex-ai, anthropic-api, openai-api, gemini-api, openrouter)", envName)
	}
	return NewProviderByName(name)
}

// NewProviderByName creates a provider by explicit name.
// Fails fast if required credentials are missing.
func NewProviderByName(name string) (anyllm.Provider, error) {
	switch name {
	case "vertex-ai":
		region := os.Getenv(envVertexRegion)
		project := os.Getenv(envVertexProject)
		if region == "" || project == "" {
			return nil, fmt.Errorf("vertex-ai requires CLOUD_ML_REGION and ANTHROPIC_VERTEX_PROJECT_ID")
		}
		return NewVertexProvider(context.Background(), region, project)

	case "anthropic-api", "anthropic", "claude":
		if os.Getenv(envAnthropicKey) == "" {
			return nil, fmt.Errorf("anthropic-api requires ANTHROPIC_API_KEY")
		}
		return anyllmAnthropic.New()

	case "openai-api", "openai", "gpt":
		if os.Getenv(envOpenAIKey) == "" {
			return nil, fmt.Errorf("openai-api requires OPENAI_API_KEY")
		}
		return anyllmOpenAI.New()

	case "gemini-api", "gemini":
		if os.Getenv(envGeminiKey) == "" {
			return nil, fmt.Errorf("gemini-api requires GEMINI_API_KEY")
		}
		return anyllmGemini.New()

	case "openrouter":
		key := os.Getenv(envOpenRouterKey)
		if key == "" {
			return nil, fmt.Errorf("openrouter requires OPENROUTER_API_KEY")
		}
		return anyllmOpenAI.New(
			anyllmConfig.WithAPIKey(key),
			anyllmConfig.WithBaseURL(openRouterBaseURL),
		)

	default:
		return nil, fmt.Errorf("unknown provider %q (options: vertex-ai, anthropic-api, openai-api, gemini-api, openrouter)", name)
	}
}
