package groups

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

func TestGroupContextPaths(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/groups/team-a/members":
			if request.URL.Query().Get("qt") != "path" || request.URL.Query().Get("search") != "alice" {
				t.Errorf("members query = %v", request.URL.Query())
			}
		case "/7/groups/9/projects":
			if request.URL.Query().Get("search") != "core" {
				t.Errorf("projects query = %v", request.URL.Query())
			}
		case "/7/groups/9/group_relate_programs":
		case "/7/groups/9/project_statistics":
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})

	if result := call(t, ListGroupMembersHandleFunc, map[string]any{"enterprise_id": 7, "group_id": "team-a", "search": "alice"}); result.IsError {
		t.Fatalf("members error: %#v", result.Content)
	}
	if result := call(t, ListGroupProjectsHandleFunc, map[string]any{"enterprise_id": 7, "group_id": 9, "search": "core"}); result.IsError {
		t.Fatalf("projects error: %#v", result.Content)
	}
	if result := call(t, ListGroupProgramsHandleFunc, map[string]any{"enterprise_id": 7, "group_id": 9}); result.IsError {
		t.Fatalf("programs error: %#v", result.Content)
	}
	if result := call(t, GetGroupStatisticsHandleFunc, map[string]any{"enterprise_id": 7, "group_id": 9}); result.IsError {
		t.Fatalf("statistics error: %#v", result.Content)
	}
}

func TestGroupRequiresGroupID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without group_id")
	})
	result := call(t, ListGroupMembersHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}
