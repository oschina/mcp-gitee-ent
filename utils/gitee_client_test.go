package utils

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type testAuthorizer struct{}

func (testAuthorizer) Authorize(req *http.Request) error {
	req.Header.Set("Authorization", "test")
	return nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type trackingReadCloser struct {
	io.Reader
	closed atomic.Bool
}

func (body *trackingReadCloser) Close() error {
	body.closed.Store(true)
	return nil
}

func TestGiteeClientUsesCustomBodyContentTypeAndContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if got := req.Header.Get("Content-Type"); got != "text/plain" {
			t.Errorf("Content-Type = %q, want text/plain", got)
		}
		if got := req.Header.Get("Authorization"); got != "test" {
			t.Errorf("Authorization = %q, want test", got)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(body); got != "request body" {
			t.Errorf("body = %q, want request body", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer server.Close()

	client := NewGiteeClient(
		http.MethodPost,
		"/test",
		WithApiBase(server.URL),
		WithHTTPClient(server.Client()),
		WithAuthorizer(testAuthorizer{}),
		WithBody(strings.NewReader("request body"), "text/plain"),
	)
	result, err := client.DoContext(context.Background())
	if err != nil {
		t.Fatalf("DoContext() error = %v", err)
	}
	defer result.Response.Body.Close()
	if result.Response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.Response.StatusCode)
	}
}

func TestGiteeClientHonorsContextCancellation(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewGiteeClient(
		http.MethodGet,
		"/test",
		WithApiBase("http://example.invalid"),
		WithHTTPClient(httpClient),
		WithAuthorizer(testAuthorizer{}),
	)
	if _, err := client.DoContext(ctx); !IsNetworkError(err) {
		t.Fatalf("DoContext() error = %v, want network error", err)
	}
}

func TestGiteeClientClosesErrorResponseBody(t *testing.T) {
	body := &trackingReadCloser{Reader: strings.NewReader(`{"message":"bad request"}`)}
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       body,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})}

	client := NewGiteeClient(
		http.MethodGet,
		"/test",
		WithApiBase("http://example.invalid"),
		WithHTTPClient(httpClient),
		WithAuthorizer(testAuthorizer{}),
	)
	_, err := client.DoContext(context.Background())
	if !IsAPIError(err) {
		t.Fatalf("DoContext() error = %v, want API error", err)
	}
	if !body.closed.Load() {
		t.Fatal("error response body was not closed")
	}
	var giteeErr *GiteeError
	if !errors.As(err, &giteeErr) || giteeErr.Details != "bad request" {
		t.Fatalf("API error = %#v, want parsed details", err)
	}
}
