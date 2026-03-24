package dispatch

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewHTTPDispatcher_RejectsPlainHTTP(t *testing.T) {
	_, err := NewHTTPDispatcher("http://evil.example.com")
	if err == nil {
		t.Fatal("expected error for non-HTTPS URL")
	}
}

func TestNewHTTPDispatcher_AllowsLocalhost(t *testing.T) {
	d, err := NewHTTPDispatcher("http://localhost:8080")
	if err != nil {
		t.Fatalf("localhost should be allowed: %v", err)
	}
	if d.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL = %q", d.BaseURL)
	}
}

func TestNewHTTPDispatcher_AllowsHTTPS(t *testing.T) {
	d, err := NewHTTPDispatcher("https://api.openai.com")
	if err != nil {
		t.Fatalf("HTTPS should be allowed: %v", err)
	}
	if d.Model != "gpt-4" {
		t.Errorf("default model = %q, want gpt-4", d.Model)
	}
}

func TestNewHTTPDispatcher_Options(t *testing.T) {
	d, err := NewHTTPDispatcher("https://api.example.com",
		WithModel("claude-3"),
		WithAPIKeyEnv("MY_KEY"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if d.Model != "claude-3" {
		t.Errorf("Model = %q, want claude-3", d.Model)
	}
	if d.apiKeyEnv != "MY_KEY" {
		t.Errorf("apiKeyEnv = %q, want MY_KEY", d.apiKeyEnv)
	}
}

func TestHTTPDispatcher_Dispatch_Integration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key-123" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("unexpected messages: %+v", req.Messages)
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: `{"result": "test_output"}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	t.Setenv("TEST_HTTP_KEY", "test-key-123")

	tmpDir := t.TempDir()
	promptPath := filepath.Join(tmpDir, "prompt.txt")
	artifactPath := filepath.Join(tmpDir, "artifact.json")
	if err := os.WriteFile(promptPath, []byte("Analyze this test case"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := NewHTTPDispatcher(srv.URL,
		WithAPIKeyEnv("TEST_HTTP_KEY"),
		WithModel("test-model"),
		WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := d.Dispatch(context.Background(), Context{
		DispatchID:   1,
		CaseID:       "C01",
		Step:         "F1_TRIAGE",
		PromptPath:   promptPath,
		ArtifactPath: artifactPath,
	})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}

	if string(result) != `{"result": "test_output"}` {
		t.Errorf("result = %q", string(result))
	}

	saved, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("read artifact: %v", err)
	}
	if !bytes.Equal(saved, result) {
		t.Errorf("saved artifact differs from returned result")
	}
}

func TestHTTPDispatcher_Dispatch_MissingAPIKey(t *testing.T) {
	t.Setenv("MISSING_KEY_VAR", "")

	tmpDir := t.TempDir()
	promptPath := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewHTTPDispatcher("http://localhost:9999", WithAPIKeyEnv("MISSING_KEY_VAR"))
	_, err := d.Dispatch(context.Background(), Context{PromptPath: promptPath})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestHTTPDispatcher_Dispatch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("internal server error")); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	t.Setenv("TEST_ERR_KEY", "key")

	tmpDir := t.TempDir()
	promptPath := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewHTTPDispatcher(srv.URL, WithAPIKeyEnv("TEST_ERR_KEY"), WithHTTPClient(srv.Client()))
	_, err := d.Dispatch(context.Background(), Context{
		PromptPath:   promptPath,
		ArtifactPath: filepath.Join(tmpDir, "out.json"),
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestHTTPDispatcher_Dispatch_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(chatResponse{Choices: []chatChoice{}}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	t.Setenv("TEST_EMPTY_KEY", "key")

	tmpDir := t.TempDir()
	promptPath := filepath.Join(tmpDir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	d, _ := NewHTTPDispatcher(srv.URL, WithAPIKeyEnv("TEST_EMPTY_KEY"), WithHTTPClient(srv.Client()))
	_, err := d.Dispatch(context.Background(), Context{
		PromptPath:   promptPath,
		ArtifactPath: filepath.Join(tmpDir, "out.json"),
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}
