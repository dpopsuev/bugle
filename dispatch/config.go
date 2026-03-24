package dispatch

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// ProviderConfig is the YAML-loadable specification for a dispatch topology:
// named providers with their type and configuration, plus fallback chains.
type ProviderConfig struct {
	Providers []ProviderDef       `yaml:"providers"`
	Fallbacks map[string][]string `yaml:"fallbacks,omitempty"`
}

// ProviderDef describes a single named provider.
// Type selects the dispatcher factory; Config carries type-specific parameters.
type ProviderDef struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	Config map[string]any `yaml:"config,omitempty"`
}

// LoadProviderConfig parses a YAML file into a ProviderConfig.
func LoadProviderConfig(path string) (*ProviderConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("dispatch/config: read %s: %w", path, err)
	}
	return ParseProviderConfig(data)
}

// ParseProviderConfig parses raw YAML bytes into a ProviderConfig.
func ParseProviderConfig(data []byte) (*ProviderConfig, error) {
	var cfg ProviderConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("dispatch/config: parse YAML: %w", err)
	}
	if len(cfg.Providers) == 0 {
		return nil, fmt.Errorf("dispatch/config: no providers defined")
	}
	for i, p := range cfg.Providers {
		if p.Name == "" {
			return nil, fmt.Errorf("dispatch/config: provider[%d] missing name", i)
		}
		if p.Type == "" {
			return nil, fmt.Errorf("dispatch/config: provider %q missing type", p.Name)
		}
	}
	return &cfg, nil
}

// DispatcherFactory creates a Dispatcher from a ProviderDef's config map.
type DispatcherFactory func(config map[string]any) (Dispatcher, error)

// BuildRouter constructs a ProviderRouter from a ProviderConfig.
//
// Built-in types ("http", "cli", "stdin", "static") are resolved automatically.
// For types that require runtime wiring (e.g. "mux", "file"), register a factory
// via the extraFactories parameter or replace the route entry after construction.
//
// The first provider in the list becomes the default dispatcher.
func BuildRouter(cfg *ProviderConfig, extraFactories map[string]DispatcherFactory) (*ProviderRouter, error) {
	factories := builtinFactories()
	for k, v := range extraFactories {
		factories[k] = v
	}

	routes := make(map[string]Dispatcher, len(cfg.Providers))
	var defaultDisp Dispatcher

	for i, pdef := range cfg.Providers {
		factory, ok := factories[pdef.Type]
		if !ok {
			return nil, fmt.Errorf("dispatch/config: provider %q: unknown type %q", pdef.Name, pdef.Type)
		}
		d, err := factory(pdef.Config)
		if err != nil {
			return nil, fmt.Errorf("dispatch/config: provider %q: %w", pdef.Name, err)
		}
		routes[pdef.Name] = d
		if i == 0 {
			defaultDisp = d
		}
	}

	router := NewProviderRouter(defaultDisp, routes, WithFallbacks(cfg.Fallbacks))
	return router, nil
}

func builtinFactories() map[string]DispatcherFactory {
	return map[string]DispatcherFactory{
		"http":   httpFactory,
		"cli":    cliFactory,
		"stdin":  stdinFactory,
		"static": staticFactory,
	}
}

func staticFactory(config map[string]any) (Dispatcher, error) {
	dir, _ := config["dir"].(string)
	return NewStaticDispatcher(dir), nil
}

func httpFactory(config map[string]any) (Dispatcher, error) {
	baseURL, _ := config["base_url"].(string)
	if baseURL == "" {
		return nil, fmt.Errorf("http provider requires config.base_url")
	}

	var opts []HTTPOption
	if model, ok := config["model"].(string); ok && model != "" {
		opts = append(opts, WithModel(model))
	}
	if keyEnv, ok := config["api_key_env"].(string); ok && keyEnv != "" {
		opts = append(opts, WithAPIKeyEnv(keyEnv))
	}

	return NewHTTPDispatcher(baseURL, opts...)
}

func cliFactory(config map[string]any) (Dispatcher, error) {
	command, _ := config["command"].(string)
	if command == "" {
		return nil, fmt.Errorf("cli provider requires config.command")
	}

	var opts []CLIOption
	if args, ok := config["args"].([]any); ok {
		strArgs := make([]string, 0, len(args))
		for _, a := range args {
			strArgs = append(strArgs, fmt.Sprintf("%v", a))
		}
		opts = append(opts, WithCLIArgs(strArgs...))
	}
	if timeoutStr, ok := config["timeout"].(string); ok {
		dur, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("cli provider: invalid timeout %q: %w", timeoutStr, err)
		}
		opts = append(opts, WithCLITimeout(dur))
	}

	return NewCLIDispatcher(command, opts...)
}

func stdinFactory(_ map[string]any) (Dispatcher, error) {
	return NewStdinDispatcher(), nil
}
