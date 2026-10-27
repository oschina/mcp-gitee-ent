package common_views

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListCommonViews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/common_views" || request.URL.Query().Get("type") != "program_view" {
			t.Errorf("request = %s %s?%s", request.Method, request.URL.Path, request.URL.RawQuery)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	}))
	defer server.Close()
	previousBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() { utils.SetApiBase(previousBase); utils.SetGiteeAccessToken(previousToken) })

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"enterprise_id": 7, "type": "program_view"}
	result, err := ListCommonViewsHandleFunc(context.Background(), request)
	if err != nil || result.IsError {
		t.Fatalf("handler error = %v result = %#v", err, result)
	}
}
