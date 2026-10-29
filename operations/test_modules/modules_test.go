package test_modules

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

func TestListAndTreeModules(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/7/case_modules" {
			if request.URL.Query().Get("program_id") != "9" || request.URL.Query().Get("search") != "login" {
				t.Errorf("list query = %v", request.URL.Query())
			}
		} else if request.URL.Path == "/7/case_modules/as_tree" {
			if request.URL.Query().Get("program_id") != "9" {
				t.Errorf("tree query = %v", request.URL.Query())
			}
			_, _ = io.WriteString(w, `[]`)
			return
		} else {
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})

	if result := call(t, ListTestCaseModulesHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "search": "login"}); result.IsError {
		t.Fatalf("list error: %#v", result.Content)
	}
	if result := call(t, GetTestCaseModuleTreeHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9}); result.IsError {
		t.Fatalf("tree error: %#v", result.Content)
	}
}

func TestCreateUpdateMoveModule(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		body := map[string]any{}
		_ = json.NewDecoder(request.Body).Decode(&body)
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/7/case_modules":
			if body["name"] != "Login" || body["program_id"] != float64(9) {
				t.Errorf("create body = %#v", body)
			}
			_, _ = io.WriteString(w, `{"id":5}`)
			return
		case request.Method == http.MethodPut && request.URL.Path == "/7/case_modules/5":
			if body["name"] != "Auth" {
				t.Errorf("update body = %#v", body)
			}
		case request.Method == http.MethodPut && request.URL.Path == "/7/case_modules/5/move":
			if body["direction"] != "inside" || body["target_id"] != float64(6) {
				t.Errorf("move body = %#v", body)
			}
		default:
			t.Errorf("unexpected request %s %s", request.Method, request.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if result := call(t, CreateTestCaseModuleHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "name": "Login"}); result.IsError {
		t.Fatalf("create error: %#v", result.Content)
	}
	if result := call(t, UpdateTestCaseModuleHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "case_module_id": 5, "name": "Auth"}); result.IsError {
		t.Fatalf("update error: %#v", result.Content)
	}
	if result := call(t, MoveTestCaseModuleHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "case_module_id": 5, "direction": "inside", "target_id": 6}); result.IsError {
		t.Fatalf("move error: %#v", result.Content)
	}
}

func TestModuleRequiresProgramID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without program_id")
	})
	result := call(t, ListTestCaseModulesHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestDeleteTestCaseModule(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete || request.URL.Path != "/7/case_modules/5" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.URL.Query().Get("program_id") != "9" {
			t.Errorf("query = %v", request.URL.Query())
		}
		w.WriteHeader(http.StatusNoContent)
	})
	result := call(t, DeleteTestCaseModuleHandleFunc, map[string]any{"enterprise_id": 7, "program_id": 9, "case_module_id": 5})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}
