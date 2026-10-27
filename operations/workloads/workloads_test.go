package workloads

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

func TestListWorkloadsPassesRangeAndFilters(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/workloads/v2" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("registered_dates") != "2026-10-01,2026-10-31" {
			t.Errorf("registered_dates = %q", query.Get("registered_dates"))
		}
		if query.Get("user_ids") != "1,2" {
			t.Errorf("user_ids = %q", query.Get("user_ids"))
		}
		if query.Get("program_ids") != "3" {
			t.Errorf("program_ids = %q", query.Get("program_ids"))
		}
		if query.Get("per_page") != "50" {
			t.Errorf("per_page = %q", query.Get("per_page"))
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListWorkloadsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"start_date":    "2026-10-01",
		"end_date":      "2026-10-31",
		"user_ids":      []any{1, 2},
		"program_ids":   []any{3},
		"per_page":      50,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListWorkloadsRejectsInvalidDate(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent for an invalid date")
	})

	result := call(t, ListWorkloadsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"start_date":    "2026/10/01",
		"end_date":      "2026-10-31",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestListWorkloadsRejectsInvertedRange(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent for an inverted range")
	})

	result := call(t, ListWorkloadsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"start_date":    "2026-11-01",
		"end_date":      "2026-10-01",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}
