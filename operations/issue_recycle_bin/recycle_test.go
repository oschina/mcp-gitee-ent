package issue_recycle_bin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRestoreIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/7/issue_recycle_bin/recover_issue" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["issue_ids"] != "1,2" || body["recover_children"] != true {
			t.Errorf("body = %#v", body)
		}
		_, _ = io.WriteString(w, `{"status":true}`)
	}))
	defer server.Close()
	previousBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() { utils.SetApiBase(previousBase); utils.SetGiteeAccessToken(previousToken) })

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"enterprise_id":    7,
		"issue_ids":        []any{1, 2},
		"recover_children": true,
	}
	result, err := RestoreIssuesHandleFunc(context.Background(), request)
	if err != nil || result.IsError {
		t.Fatalf("handler error = %v result = %#v", err, result)
	}
}
