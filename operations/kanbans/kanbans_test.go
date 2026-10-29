package kanbans

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestKanbanPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/kanbans":
			if request.URL.Query().Get("with_columns") != "1" {
				t.Errorf("query = %v", request.URL.Query())
			}
		case "/7/programs/9/kanbans/5/columns":
			_, _ = io.WriteString(w, `[]`)
			return
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
		{ListKanbansHandleFunc, map[string]any{"enterprise_id": 7, "include_columns": true}},
		{ListKanbanColumnsHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "kanban_id": 5}},
	} {
		request := mcp.CallToolRequest{}
		request.Params.Arguments = tc.arguments
		result, err := tc.handler(context.Background(), request)
		if err != nil || result.IsError {
			t.Fatalf("handler error = %v result = %#v", err, result)
		}
	}
}
