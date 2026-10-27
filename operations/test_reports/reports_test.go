package test_reports

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

func call(t *testing.T, arguments map[string]any) *mcp.CallToolResult {
	return callHandler(t, CreateTestReportHandleFunc, arguments)
}

func callHandler(t *testing.T, handler utils.OptionHandlerFunc, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	request := mcp.CallToolRequest{}
	request.Params.Arguments = arguments
	result, err := handler(context.Background(), request)
	if result == nil {
		t.Fatalf("handler error = %v", err)
	}
	return result
}

func TestCreateTestReportPostsPlanIDsAndComponents(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/test_reports" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["title"] != "Release report" || body["program_id"] != float64(9) || body["state"] != true {
			t.Errorf("body = %#v", body)
		}
		planIDs, ok := body["test_plan_ids"].([]any)
		if !ok || len(planIDs) != 2 {
			t.Errorf("test_plan_ids = %#v", body["test_plan_ids"])
		}
		components, ok := body["components"].([]any)
		if !ok || len(components) != 1 {
			t.Errorf("components = %#v", body["components"])
		}
		_, _ = io.WriteString(w, `{"id":5,"title":"Release report"}`)
	})

	result := call(t, map[string]any{
		"enterprise_id": 7,
		"program_id":    9,
		"title":         "Release report",
		"test_plan_ids": []any{101, 102},
		"state":         true,
		"components": []any{
			map[string]any{"ident": "summary", "title": "Summary"},
		},
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestCreateTestReportRequiresTestPlanIDs(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without test_plan_ids")
	})
	result := call(t, map[string]any{
		"enterprise_id": 7,
		"program_id":    9,
		"title":         "Release report",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestListTestReports(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/test_reports" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.URL.Query().Get("program_id") != "9" || request.URL.Query().Get("search") != "Release" {
			t.Errorf("query = %v", request.URL.Query())
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})

	result := callHandler(t, ListTestReportsHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "search": "Release"})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestGetTestReportUsesPostWithComponents(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/test_reports/5" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["program_id"] != float64(9) {
			t.Errorf("program_id = %v", body["program_id"])
		}
		if _, ok := body["components"]; !ok {
			t.Error("components missing")
		}
		_, _ = io.WriteString(w, `{"id":5}`)
	})

	result := callHandler(t, GetTestReportHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_report_id": 5,
		"components": []any{map[string]any{"id": 1, "ident": "summary", "title": "Summary"}},
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestUpdateAndDeleteTestReport(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodPut && request.URL.Path == "/7/test_reports/5":
			body := map[string]any{}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if body["title"] != "Updated" {
				t.Errorf("update body = %#v", body)
			}
			_, _ = io.WriteString(w, `{"id":5}`)
		case request.Method == http.MethodDelete && request.URL.Path == "/7/test_reports/5":
			if request.URL.Query().Get("program_id") != "9" {
				t.Errorf("delete query = %v", request.URL.Query())
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
	})

	if result := callHandler(t, UpdateTestReportHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "test_report_id": 5, "title": "Updated"}); result.IsError {
		t.Fatalf("update error: %#v", result.Content)
	}
	if result := callHandler(t, DeleteTestReportHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "test_report_id": 5}); result.IsError {
		t.Fatalf("delete error: %#v", result.Content)
	}
}
