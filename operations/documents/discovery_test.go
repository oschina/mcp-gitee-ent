package documents

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func withDiscoveryTestServer(t *testing.T, handler http.HandlerFunc) {
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

func callDiscovery(t *testing.T, handler utils.OptionHandlerFunc, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	request := mcp.CallToolRequest{}
	request.Params.Arguments = arguments
	result, err := handler(context.Background(), request)
	if result == nil {
		t.Fatalf("handler error = %v", err)
	}
	return result
}

func TestDocumentDiscoveryPaths(t *testing.T) {
	withDiscoveryTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/doc_nodes/search/v2":
			if request.URL.Query().Get("search") != "onboarding" {
				t.Errorf("search query = %v", request.URL.Query())
			}
		case "/7/doc_nodes/recent_doc":
		case "/7/doc_nodes/shared":
		case "/7/doc_nodes/5":
		case "/7/doc_nodes/5/file_versions":
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})

	if result := callDiscovery(t, SearchDocumentsHandleFunc, map[string]any{"enterprise_id": 7, "search": "onboarding"}); result.IsError {
		t.Fatalf("search error: %#v", result.Content)
	}
	if result := callDiscovery(t, ListRecentDocumentsHandleFunc, map[string]any{"enterprise_id": 7}); result.IsError {
		t.Fatalf("recent error: %#v", result.Content)
	}
	if result := callDiscovery(t, ListSharedDocumentsHandleFunc, map[string]any{"enterprise_id": 7}); result.IsError {
		t.Fatalf("shared error: %#v", result.Content)
	}
	if result := callDiscovery(t, GetDocumentNodeHandleFunc, map[string]any{"enterprise_id": 7, "doc_node_id": 5}); result.IsError {
		t.Fatalf("node error: %#v", result.Content)
	}
	if result := callDiscovery(t, ListDocumentVersionsHandleFunc, map[string]any{"enterprise_id": 7, "doc_node_id": 5}); result.IsError {
		t.Fatalf("versions error: %#v", result.Content)
	}
}

func TestSearchDocumentsRequiresSearch(t *testing.T) {
	withDiscoveryTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without search")
	})
	result := callDiscovery(t, SearchDocumentsHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestGetDocumentNodeRenamesLinkedIssues(t *testing.T) {
	withDiscoveryTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/doc_nodes/5" {
			t.Errorf("path = %q", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":5,"file_type":"WikiInfo","linked_issues_count":[{"id":31},{"id":32}]}`)
	})

	result := callDiscovery(t, GetDocumentNodeHandleFunc, map[string]any{"enterprise_id": 7, "doc_node_id": 5})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
	var decoded map[string]any
	toolResultJSON(t, result, &decoded)
	if _, ok := decoded["linked_issues_count"]; ok {
		t.Fatalf("linked_issues_count must be renamed: %v", decoded)
	}
	issues, ok := decoded["linked_issues"].([]any)
	if !ok || len(issues) != 2 {
		t.Fatalf("linked_issues = %#v", decoded["linked_issues"])
	}
	if total, ok := decoded["linked_issues_total"].(float64); !ok || int(total) != 2 {
		t.Fatalf("linked_issues_total = %#v", decoded["linked_issues_total"])
	}
}

func TestGetDocumentNodeWithoutLinkedIssueField(t *testing.T) {
	withDiscoveryTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(w, `{"id":5,"file_type":"AttachFile"}`)
	})

	result := callDiscovery(t, GetDocumentNodeHandleFunc, map[string]any{"enterprise_id": 7, "doc_node_id": 5})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
	var decoded map[string]any
	toolResultJSON(t, result, &decoded)
	if _, ok := decoded["linked_issues"]; ok {
		t.Fatalf("linked_issues must not be invented: %v", decoded)
	}
	if fileType, _ := decoded["file_type"].(string); fileType != "AttachFile" {
		t.Fatalf("file_type = %#v", decoded["file_type"])
	}
}
