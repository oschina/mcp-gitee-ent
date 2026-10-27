package milestones

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

func TestListMilestonesPassesFilters(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/milestones" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("search") != "Q4" || query.Get("state") != "active" || query.Get("page") != "2" {
			t.Errorf("query = %v", query)
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListMilestonesHandleFunc, map[string]any{
		"enterprise_id": 7,
		"search":        "Q4",
		"state":         "active",
		"page":          2,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestGetMilestoneUsesMilestonePath(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/milestones/9" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":9,"title":"Q4"}`)
	})

	result := call(t, GetMilestoneHandleFunc, map[string]any{"enterprise_id": 7, "milestone_id": 9})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestMilestoneHandlersRequireMilestoneID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent when milestone_id is missing")
	})

	result := call(t, GetMilestoneHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestCreateMilestonePostsPayload(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/milestones" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["title"] != "Q4" || body["due_date"] != "2026-12-31" || body["description"] != "release" {
			t.Errorf("body = %#v", body)
		}
		_, _ = io.WriteString(w, `{"id":9,"title":"Q4"}`)
	})

	result := call(t, CreateMilestoneHandleFunc, map[string]any{
		"enterprise_id": 7,
		"title":         "Q4",
		"due_date":      "2026-12-31",
		"description":   "release",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestCreateMilestoneRejectsMissingDueDate(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent when due_date is missing")
	})
	result := call(t, CreateMilestoneHandleFunc, map[string]any{"enterprise_id": 7, "title": "Q4"})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestUpdateMilestonePutsFields(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/7/milestones/9" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["due_date"] != "2026-12-31" || body["top"] != float64(1) {
			t.Errorf("body = %#v", body)
		}
		_, _ = io.WriteString(w, `{"id":9}`)
	})

	result := call(t, UpdateMilestoneHandleFunc, map[string]any{
		"enterprise_id": 7,
		"milestone_id":  9,
		"due_date":      "2026-12-31",
		"top":           1,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestUpdateMilestoneRequiresMutableField(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without mutable fields")
	})
	result := call(t, UpdateMilestoneHandleFunc, map[string]any{"enterprise_id": 7, "milestone_id": 9})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}
