package labels

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

func TestCreateLabelPostsPayload(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/labels" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["name"] != "backend" || body["color"] != "1E90FF" {
			t.Errorf("body = %#v", body)
		}
		_, _ = io.WriteString(w, `{"id":5,"name":"backend"}`)
	})

	result := call(t, CreateLabelHandleFunc, map[string]any{
		"enterprise_id": 7,
		"name":          "backend",
		"color":         "#1E90FF",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestCreateLabelRejectsInvalidColor(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent for an invalid color")
	})
	result := call(t, CreateLabelHandleFunc, map[string]any{
		"enterprise_id": 7,
		"name":          "backend",
		"color":         "blue",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestUpdateLabelPutsNameAndColor(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/7/labels/5" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["name"] != "frontend" || body["color"] != "FF0000" {
			t.Errorf("body = %#v", body)
		}
		_, _ = io.WriteString(w, `{"id":5}`)
	})

	result := call(t, UpdateLabelHandleFunc, map[string]any{
		"enterprise_id": 7,
		"label_id":      5,
		"name":          "frontend",
		"color":         "#FF0000",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestUpdateLabelRequiresColor(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent when color is missing")
	})
	result := call(t, UpdateLabelHandleFunc, map[string]any{
		"enterprise_id": 7,
		"label_id":      5,
		"name":          "frontend",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}
