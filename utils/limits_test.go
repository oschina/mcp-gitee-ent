package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestWithQueryClampsPerPage(t *testing.T) {
	client := NewGiteeClient("GET", "/test", WithQuery(map[string]interface{}{"per_page": 500}))
	if got := client.Query["per_page"]; got != "100" {
		t.Fatalf("per_page = %q, want 100", got)
	}
}

func TestGETRetriesTransientServerError(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	previousToken := GetGiteeAccessToken()
	SetGiteeAccessToken("test-token")
	t.Cleanup(func() { SetGiteeAccessToken(previousToken) })

	client := NewGiteeClient("GET", "/test", WithApiBase(server.URL))
	if _, err := client.DoContext(context.Background()); err != nil {
		t.Fatalf("DoContext() error = %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestMCPJSONResultRejectsOversizedResponse(t *testing.T) {
	t.Setenv("GITEE_ENT_MAX_RESPONSE_BYTES", "8")
	result, err := MCPJSONResult(strings.Repeat("x", 32))
	if err != nil {
		t.Fatalf("MCPJSONResult() error = %v", err)
	}
	if !result.IsError {
		t.Fatalf("result = %#v, want size error", result)
	}
}
