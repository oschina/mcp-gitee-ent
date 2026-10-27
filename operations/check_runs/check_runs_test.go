package check_runs

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

func TestCheckRunLists(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/pr_check_runs":
			if request.URL.Query().Get("pull_request_ids") != "1,2" {
				t.Errorf("pr query = %v", request.URL.Query())
			}
		case "/7/commit_check_runs":
			if request.URL.Query().Get("project_id") != "9" || request.URL.Query().Get("commit_ids") != "abc" {
				t.Errorf("commit query = %v", request.URL.Query())
			}
		case "/7/projects/9/check_runs/list":
			if request.URL.Query().Get("head_sha") != "abc" {
				t.Errorf("project query = %v", request.URL.Query())
			}
		case "/7/projects/9/check_runs/annotations":
			if request.Method != http.MethodPost {
				t.Errorf("method = %s", request.Method)
			}
			body := map[string]any{}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["check_run_id"] != float64(5) {
				t.Errorf("annotations body = %#v", body)
			}
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	})

	if result := call(t, ListPRCheckRunsHandleFunc, map[string]any{"enterprise_id": 7, "pull_request_ids": "1,2"}); result.IsError {
		t.Fatalf("pr check error: %#v", result.Content)
	}
	if result := call(t, ListCommitCheckRunsHandleFunc, map[string]any{"enterprise_id": 7, "project_id": 9, "commit_ids": "abc"}); result.IsError {
		t.Fatalf("commit check error: %#v", result.Content)
	}
	if result := call(t, ListProjectCheckRunsHandleFunc, map[string]any{"enterprise_id": 7, "project_id": 9, "head_sha": "abc"}); result.IsError {
		t.Fatalf("project check error: %#v", result.Content)
	}
	if result := call(t, GetCheckRunAnnotationsHandleFunc, map[string]any{"enterprise_id": 7, "project_id": 9, "check_run_id": 5}); result.IsError {
		t.Fatalf("annotations error: %#v", result.Content)
	}
}

func TestCommitCheckRunsRequiresProject(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without project_id")
	})
	result := call(t, ListCommitCheckRunsHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}
