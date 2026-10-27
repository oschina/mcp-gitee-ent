package webhooks

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestWebhookLists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/hooks":
			if request.URL.Query().Get("scope") != "all" {
				t.Errorf("query = %v", request.URL.Query())
			}
		case "/7/hooks/5/logs":
			if request.Method != http.MethodPost {
				t.Errorf("method = %s", request.Method)
			}
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	}))
	defer server.Close()
	previousBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() { utils.SetApiBase(previousBase); utils.SetGiteeAccessToken(previousToken) })

	for _, tc := range []struct {
		handler   utils.OptionHandlerFunc
		arguments map[string]any
	}{
		{ListWebhooksHandleFunc, map[string]any{"enterprise_id": 7, "scope": "all"}},
		{ListWebhookLogsHandleFunc, map[string]any{"enterprise_id": 7, "webhook_id": 5}},
	} {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = tc.arguments
		result, err := tc.handler(context.Background(), request)
		if err != nil || result.IsError {
			t.Fatalf("handler error = %v result = %#v", err, result)
		}
	}
}

func TestListWebhooksRedactsSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(w, `{"data":[{"password":"secret-signing-key","url":"https://example.com/hook?access_token=secret-token"}],"total_count":1}`)
	}))
	defer server.Close()
	previousBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() { utils.SetApiBase(previousBase); utils.SetGiteeAccessToken(previousToken) })

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"enterprise_id": 7, "scope": "all"}
	result, err := ListWebhooksHandleFunc(context.Background(), request)
	if err != nil || result.IsError {
		t.Fatalf("handler error = %v result = %#v", err, result)
	}
	text := result.Content[0].(mcp.TextContent).Text
	for _, secret := range []string{"secret-signing-key", "secret-token"} {
		if strings.Contains(text, secret) {
			t.Fatalf("response leaked %q: %s", secret, text)
		}
	}
	if !strings.Contains(text, "[REDACTED]") {
		t.Fatalf("response was not redacted: %s", text)
	}
}
