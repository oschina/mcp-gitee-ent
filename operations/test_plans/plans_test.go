package test_plans

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

func TestListAndGetTestPlan(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %s", request.Method)
		}
		switch request.URL.Path {
		case "/7/test_plans":
			if request.URL.Query().Get("program_id") != "9" || request.URL.Query().Get("states") != "0,1" {
				t.Errorf("list query = %v", request.URL.Query())
			}
		case "/7/test_plans/5":
			if request.URL.Query().Get("program_id") != "9" {
				t.Errorf("detail query = %v", request.URL.Query())
			}
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})

	if result := call(t, ListTestPlansHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "states": "0,1"}); result.IsError {
		t.Fatalf("list error: %#v", result.Content)
	}
	if result := call(t, GetTestPlanHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "test_plan_id": 5}); result.IsError {
		t.Fatalf("get error: %#v", result.Content)
	}
}

func TestCreateUpdateAndStateTestPlan(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		body := map[string]any{}
		_ = json.NewDecoder(request.Body).Decode(&body)
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/7/test_plans":
			if body["title"] != "Release" || body["assignee_id"] != float64(1001) {
				t.Errorf("create body = %#v", body)
			}
		case request.Method == http.MethodPut && request.URL.Path == "/7/test_plans/5":
			if body["title"] != "Release v2" || body["assignee_id"] != float64(1001) {
				t.Errorf("update body = %#v", body)
			}
		case request.Method == http.MethodPut && request.URL.Path == "/7/test_plans/5/state":
			if body["state"] != float64(1) {
				t.Errorf("state body = %#v", body)
			}
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":5}`)
	})

	if result := call(t, CreateTestPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "title": "Release", "assignee_id": 1001,
	}); result.IsError {
		t.Fatalf("create error: %#v", result.Content)
	}
	if result := call(t, UpdateTestPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "title": "Release v2", "assignee_id": 1001,
	}); result.IsError {
		t.Fatalf("update error: %#v", result.Content)
	}
	if result := call(t, UpdateTestPlanStateHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "state": 1,
	}); result.IsError {
		t.Fatalf("state error: %#v", result.Content)
	}
}

func TestUpdateTestPlanRequiresTitleAndAssignee(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without title and assignee_id")
	})
	result := call(t, UpdateTestPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5,
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestDeleteTestPlan(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete || request.URL.Path != "/7/test_plans/5" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.URL.Query().Get("program_id") != "9" {
			t.Errorf("query = %v", request.URL.Query())
		}
		w.WriteHeader(http.StatusNoContent)
	})
	result := call(t, DeleteTestPlanHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "test_plan_id": 5})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}
