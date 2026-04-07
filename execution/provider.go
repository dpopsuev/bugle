package execution

import (
	"context"
	"fmt"
	"os"

	anyllm "github.com/mozilla-ai/any-llm-go/providers"
	anyllmConfig "github.com/mozilla-ai/any-llm-go/config"
	anyllmAnthropic "github.com/mozilla-ai/any-llm-go/providers/anthropic"
	anyllmGemini "github.com/mozilla-ai/any-llm-go/providers/gemini"
	anyllmOpenAI "github.com/mozilla-ai/any-llm-go/providers/openai"
)

// Env var names for provider detection.
const (
	envUseVertex      = "CLAUDE_CODE_USE_VERTEX"
	envVertexRegion   = "CLOUD_ML_REGION"
	envVertexProject  = "ANTHROPIC_VERTEX_PROJECT_ID"
	envAnthropicKey   = "ANTHROPIC_API_KEY"
	envOpenAIKey      = "OPENAI_API_KEY"
	envGeminiKey      = "GEMINI_API_KEY"
	envOpenRouterKey  = "OPENROUTER_API_KEY"
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

// NewProviderFromEnv detects available LLM providers from environment
// variables and returns the best available one.
//
// Priority: Anthropic direct > OpenAI > Gemini.
// Vertex AI support requires upstream any-llm-go changes or a direct
// anthropic-sdk-go integration (see TRP-TSK-35).
func NewProviderFromEnv() (anyllm.Provider, error) {
	if os.Getenv(envUseVertex) == "1" {
		region := os.Getenv(envVertexRegion)
		project := os.Getenv(envVertexProject)
		if region != "" && project != "" {
			return NewVertexProvider(context.Background(), region, project)
		}
	}

	if os.Getenv(envAnthropicKey) != "" {
		return anyllmAnthropic.New()
	}

	if os.Getenv(envOpenAIKey) != "" {
		return anyllmOpenAI.New()
	}

	if os.Getenv(envGeminiKey) != "" {
		return anyllmGemini.New()
	}

	// OpenRouter — universal fallback, 352+ models from all providers.
	if os.Getenv(envOpenRouterKey) != "" {
		return anyllmOpenAI.New(
			anyllmConfig.WithAPIKey(os.Getenv(envOpenRouterKey)),
			anyllmConfig.WithBaseURL(openRouterBaseURL),
		)
	}

	return nil, fmt.Errorf("no LLM provider found: set CLAUDE_CODE_USE_VERTEX, ANTHROPIC_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, or OPENROUTER_API_KEY")
}

// NewProviderByName creates a provider by explicit name.
func NewProviderByName(name string) (anyllm.Provider, error) {
	switch name {
	case "anthropic", "claude":
		if os.Getenv(envUseVertex) == "1" {
			region := os.Getenv(envVertexRegion)
			project := os.Getenv(envVertexProject)
			if region != "" && project != "" {
				return NewVertexProvider(context.Background(), region, project)
			}
		}
		return anyllmAnthropic.New()
	case "openai", "gpt":
		return anyllmOpenAI.New()
	case "gemini":
		return anyllmGemini.New()
	case "openrouter":
		return anyllmOpenAI.New(
			anyllmConfig.WithAPIKey(os.Getenv(envOpenRouterKey)),
			anyllmConfig.WithBaseURL(openRouterBaseURL),
		)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}
