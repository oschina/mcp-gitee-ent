package test_plans

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListAndGetTestPlanCase(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/test_plans/5/test_cases":
			if request.URL.Query().Get("program_id") != "9" || request.URL.Query().Get("results") != "3" {
				t.Errorf("list query = %v", request.URL.Query())
			}
		case "/7/test_plans/5/test_cases/101":
			if request.URL.Query().Get("program_id") != "9" {
				t.Errorf("detail query = %v", request.URL.Query())
			}
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})

	if result := call(t, ListTestPlanCasesHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "results": "3",
	}); result.IsError {
		t.Fatalf("list error: %#v", result.Content)
	}
	if result := call(t, GetTestPlanCaseHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "test_case_id": 101,
	}); result.IsError {
		t.Fatalf("get error: %#v", result.Content)
	}
}

func TestAddTestCasesToPlanResolvesLatestApprovedVersion(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/test_cases/101/test_case_versions":
			_, _ = io.WriteString(w, `{"data":[{"id":11,"state":0},{"id":12,"state":2}]}`)
		case "/7/test_cases/102/test_case_versions":
			_, _ = io.WriteString(w, `{"data":[{"id":21,"state":2}]}`)
		case "/7/test_plans/5/test_cases/add_test_cases":
			body := map[string]any{}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["version_ids"] != "101-12,102-21" {
				t.Errorf("version_ids = %v", body["version_ids"])
			}
			_, _ = io.WriteString(w, `{}`)
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
	})

	result := call(t, AddTestCasesToPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5,
		"case_ids": []any{101, 102},
	})
	if result.IsError {
		t.Fatalf("add error: %#v", result.Content)
	}
}

func TestAddTestCasesToPlanHandlesNoContent(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/7/test_cases/101/test_case_versions" {
			_, _ = io.WriteString(w, `{"data":[{"id":12,"state":2}]}`)
			return
		}
		if request.URL.Path != "/7/test_plans/5/test_cases/add_test_cases" {
			t.Errorf("unexpected path %s", request.URL.Path)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	result := call(t, AddTestCasesToPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "case_ids": []any{101},
	})
	if result.IsError {
		t.Fatalf("add error: %#v", result.Content)
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T", result.Content[0])
	}
	if !strings.Contains(content.Text, `"queued":true`) {
		t.Fatalf("ack = %s", content.Text)
	}
}

func TestAddTestCasesRequiresCaseIDsOrVersionIDs(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without case_ids or version_ids")
	})
	result := call(t, AddTestCasesToPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5,
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestCreateTestCaseInPlanRequiresSteps(t *testing.T) {
	for _, steps := range []any{nil, []any{}, []any{map[string]any{"description": ""}}} {
		withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
			t.Fatal("HTTP request must not be sent without valid case_steps")
		})
		arguments := map[string]any{"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "title": "Plan case", "module_id": 3}
		if steps != nil {
			arguments["case_steps"] = steps
		}
		if result := call(t, CreateTestCaseInPlanHandleFunc, arguments); !result.IsError {
			t.Fatalf("result = %#v, want error for steps %#v", result, steps)
		}
	}
}

func TestCreateAndUpdateTestPlanCase(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		body := map[string]any{}
		_ = json.NewDecoder(request.Body).Decode(&body)
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/7/test_plans/5/test_cases":
			if body["title"] != "Plan case" || body["module_id"] != float64(3) {
				t.Errorf("create body = %#v", body)
			}
		case request.Method == http.MethodPut && request.URL.Path == "/7/test_plans/5/test_cases/101":
			if body["assignee_id"] != float64(1001) {
				t.Errorf("update body = %#v", body)
			}
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":101}`)
	})

	if result := call(t, CreateTestCaseInPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "title": "Plan case", "module_id": 3,
		"case_steps": []any{map[string]any{"description": "Open page"}},
	}); result.IsError {
		t.Fatalf("create error: %#v", result.Content)
	}
	if result := call(t, UpdateTestPlanCaseHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "test_case_id": 101, "assignee_id": 1001,
	}); result.IsError {
		t.Fatalf("update error: %#v", result.Content)
	}
}

func TestLinkUnlinkAndListPlanCaseIssues(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/7/test_plans/5/test_cases/101/link_issues":
			body := map[string]any{}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if len(body["issue_ids"].([]any)) != 2 {
				t.Errorf("link body = %#v", body)
			}
		case request.Method == http.MethodDelete && request.URL.Path == "/7/test_plans/5/test_cases/101/link_issues":
			body := map[string]any{}
			_ = json.NewDecoder(request.Body).Decode(&body)
			if len(body["issue_ids"].([]any)) != 1 {
				t.Errorf("unlink body = %#v", body)
			}
		case request.Method == http.MethodGet && request.URL.Path == "/7/test_plans/5/test_cases/101/issues":
			if request.URL.Query().Get("linked") != "true" {
				t.Errorf("list issues query = %v", request.URL.Query())
			}
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		_, _ = io.WriteString(w, `{}`)
	})

	if result := call(t, LinkTestPlanCaseIssuesHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "test_case_id": 101, "issue_ids": []any{1, 2},
	}); result.IsError {
		t.Fatalf("link error: %#v", result.Content)
	}
	if result := call(t, UnlinkTestPlanCaseIssuesHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "test_case_id": 101, "issue_ids": []any{1},
	}); result.IsError {
		t.Fatalf("unlink error: %#v", result.Content)
	}
	if result := call(t, ListTestPlanCaseIssuesHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "test_case_id": 101, "linked": true,
	}); result.IsError {
		t.Fatalf("list issues error: %#v", result.Content)
	}
}

func TestExecuteTestPlanCases(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/test_plans/5/batch_execute" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		_ = json.NewDecoder(request.Body).Decode(&body)
		if body["execute_state"] != float64(1) {
			t.Errorf("execute_state = %#v", body["execute_state"])
		}
		ids, ok := body["test_plan_case_ids"].([]any)
		if !ok || len(ids) != 2 {
			t.Errorf("test_plan_case_ids = %#v", body["test_plan_case_ids"])
		}
		_, _ = io.WriteString(w, `{"message":"queued"}`)
	})

	result := call(t, ExecuteTestPlanCasesHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5,
		"test_plan_case_ids": []any{2001, 2002}, "execute_state": 1,
	})
	if result.IsError {
		t.Fatalf("execute error: %#v", result.Content)
	}
}

func TestExecuteTestPlanCasesRequiresStateOrAssignee(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without execute_state or assignee_id")
	})
	result := call(t, ExecuteTestPlanCasesHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5, "test_plan_case_ids": []any{2001},
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestAddTestCasesRejectsUnapprovedVersion(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/7/test_cases/103/test_case_versions" {
			_, _ = io.WriteString(w, `{"data":[{"id":31,"state":0}]}`)
			return
		}
		t.Fatalf("unexpected request %s %s; add_test_cases must not be called", request.Method, request.URL.Path)
	})

	result := call(t, AddTestCasesToPlanHandleFunc, map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_id": 5,
		"case_ids": []any{103},
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want unapproved-version error", result)
	}
}
