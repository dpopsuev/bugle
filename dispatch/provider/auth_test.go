package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticTokenAuth_InjectsBearer(t *testing.T) {
	t.Setenv("TEST_AUTH_TOKEN", "my-secret-token")

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	auth := NewStaticTokenAuth("TEST_AUTH_TOKEN", nil)
	client := &http.Client{Transport: auth}

	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	resp.Body.Close()

	if gotAuth != "Bearer my-secret-token" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer my-secret-token")
	}
}

func TestStaticTokenAuth_MissingEnvVar(t *testing.T) {
	t.Setenv("EMPTY_TOKEN_VAR", "")

	auth := NewStaticTokenAuth("EMPTY_TOKEN_VAR", nil)
	client := &http.Client{Transport: auth}

	resp, err := client.Get("http://localhost:1") //nolint:bodyclose // resp is nil on error path
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected error for missing token")
	}
}

func TestStaticTokenAuth_NilInner_UsesDefault(t *testing.T) {
	auth := NewStaticTokenAuth("SOME_VAR", nil)
	if auth.Inner == nil {
		t.Fatal("Inner should default to http.DefaultTransport")
	}
}

func TestStaticTokenAuth_PreservesOriginalRequest(t *testing.T) {
	t.Setenv("TEST_PRESERVE_TOKEN", "tok")

	var gotUserAgent string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	auth := NewStaticTokenAuth("TEST_PRESERVE_TOKEN", nil)
	client := &http.Client{Transport: auth}

	req, _ := http.NewRequest("GET", srv.URL, http.NoBody)
	req.Header.Set("User-Agent", "bugle-test/1.0")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if gotUserAgent != "bugle-test/1.0" {
		t.Errorf("User-Agent = %q, want bugle-test/1.0", gotUserAgent)
	}
}
