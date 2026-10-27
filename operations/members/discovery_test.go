package members

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

func TestSearchMembersPostsFilters(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/members/search_members" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["search"] != "alice" {
			t.Errorf("search = %v", payload["search"])
		}
		if _, ok := payload["enterprise_id"]; ok {
			t.Error("enterprise_id must not be sent in the body")
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, SearchMembersHandleFunc, map[string]any{
		"enterprise_id": 7,
		"search":        "alice",
		"program_ids":   []any{1, 2},
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestGetMemberUsesUserPath(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/members/alice" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":9,"username":"alice"}`)
	})

	result := call(t, GetMemberHandleFunc, map[string]any{"enterprise_id": 7, "user_id": "alice"})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestGetMemberByEmailPassesEmailQuery(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/members/by_email" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("email") != "alice@example.com" {
			t.Errorf("email = %q", request.URL.Query().Get("email"))
		}
		_, _ = io.WriteString(w, `{"id":9}`)
	})

	result := call(t, GetMemberByEmailHandleFunc, map[string]any{
		"enterprise_id": 7,
		"email":         "alice@example.com",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListMemberMembershipsUsesTypePath(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/members/alice/programs" {
			t.Errorf("path = %q", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("search") != "core" || query.Get("page") != "2" {
			t.Errorf("query = %v", query)
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListMemberMembershipsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"user_id":       "alice",
		"type":          "programs",
		"search":        "core",
		"page":          2,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListMemberMembershipsRejectsUnknownType(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent for an invalid type")
	})

	result := call(t, ListMemberMembershipsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"user_id":       "alice",
		"type":          "teams",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestGetGroupUsesPathQueryForGroupPath(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/groups/team-a" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("qt") != "path" {
			t.Errorf("qt = %q", request.URL.Query().Get("qt"))
		}
		_, _ = io.WriteString(w, `{"id":1,"path":"team-a"}`)
	})

	result := call(t, GetGroupHandleFunc, map[string]any{"enterprise_id": 7, "group_id": "team-a"})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestGetGroupOmitsPathQueryForNumericID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("qt") != "" {
			t.Errorf("qt = %q, want empty", request.URL.Query().Get("qt"))
		}
		_, _ = io.WriteString(w, `{"id":1}`)
	})

	result := call(t, GetGroupHandleFunc, map[string]any{"enterprise_id": 7, "group_id": 1})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}
