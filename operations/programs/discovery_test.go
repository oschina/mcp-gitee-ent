package programs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func withTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	previousAPIBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(server.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() {
		server.Close()
		utils.SetApiBase(previousAPIBase)
		utils.SetGiteeAccessToken(previousToken)
	})
}

func call(t *testing.T, handler utils.OptionHandlerFunc, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	request := mcp.CallToolRequest{}
	request.Params.Arguments = arguments
	result, err := handler(context.Background(), request)
	if result == nil {
		t.Fatalf("handler error = %v", err)
	}
	return result
}

func TestProgramContextPaths(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/programs/9", "/7/programs/9/overview", "/7/programs/9/components":
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":9}`)
	})

	if result := call(t, GetProgramHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9}); result.IsError {
		t.Fatalf("get program error: %#v", result.Content)
	}
	if result := call(t, GetProgramOverviewHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9}); result.IsError {
		t.Fatalf("overview error: %#v", result.Content)
	}
	if result := call(t, GetProgramComponentsHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9}); result.IsError {
		t.Fatalf("components error: %#v", result.Content)
	}
}

func TestProgramRequiresProgramID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without program_id")
	})
	result := call(t, GetProgramHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}
