package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/utils/urlbuilder"
	"github.com/mark3labs/mcp-go/mcp"
)

func withEnterpriseResolver(t *testing.T, entPath string) {
	t.Helper()

	urlbuilder.SetWebBase("https://e.gitee.com")
	ResetEnterprisePathCache()
	SetEnterprisePath(1, entPath)

	t.Cleanup(func() {
		urlbuilder.SetWebBase("")
		ResetEnterprisePathCache()
	})
}

func TestEnrichResultAddsIssueURL(t *testing.T) {
	withEnterpriseResolver(t, "oschina")

	result := EnrichResult("/1/issues/I5DFIB", map[string]any{"ident": "I5DFIB"})
	item, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	want := "https://e.gitee.com/oschina/issues/list?issue=I5DFIB"
	if item["html_url"] != want {
		t.Fatalf("html_url = %v, want %q", item["html_url"], want)
	}
}

func TestEnrichResultAddsURLsToListItems(t *testing.T) {
	withEnterpriseResolver(t, "oschina")

	result := EnrichResult("/1/pull_requests", map[string]any{
		"total_count": 2,
		"data": []any{
			map[string]any{"iid": float64(24300), "project": map[string]any{"path_with_namespace": "oschina/gitee"}},
			map[string]any{"iid": float64(7), "project": map[string]any{"path_with_namespace": "oschina/sub/gitee"}},
		},
	})
	root, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	items, _ := root["data"].([]any)
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	first := items[0].(map[string]any)
	if first["html_url"] != "https://e.gitee.com/oschina/repos/oschina/gitee/pulls/24300" {
		t.Fatalf("first html_url = %v", first["html_url"])
	}
	second := items[1].(map[string]any)
	if second["html_url"] != "https://e.gitee.com/oschina/repos/oschina/sub/gitee/-/pulls/7" {
		t.Fatalf("second html_url = %v", second["html_url"])
	}
}

func TestEnrichResultConvertsRepoTreePaths(t *testing.T) {
	withEnterpriseResolver(t, "hightest")

	result := EnrichResult("/1/projects/st-diff/tree/master", map[string]any{
		"tree": map[string]any{
			"folders": []any{map[string]any{"path": "/hightest/st-diff/tree/master/docs"}},
			"files":   []any{map[string]any{"path": "/hightest/st-diff/blob/master/README.md", "raw_path": "README.md"}},
		},
	})
	root := result.(map[string]any)
	tree := root["tree"].(map[string]any)

	folder := tree["folders"].([]any)[0].(map[string]any)
	if folder["html_url"] != "https://e.gitee.com/hightest/repos/hightest/st-diff/tree/master/docs" {
		t.Fatalf("folder html_url = %v", folder["html_url"])
	}
	if folder["path"] != "/hightest/st-diff/tree/master/docs" {
		t.Fatalf("path must stay untouched, got %v", folder["path"])
	}

	file := tree["files"].([]any)[0].(map[string]any)
	if file["html_url"] != "https://e.gitee.com/hightest/repos/hightest/st-diff/blob/master/README.md" {
		t.Fatalf("file html_url = %v", file["html_url"])
	}
	if file["raw_path"] != "README.md" {
		t.Fatalf("raw_path must stay untouched, got %v", file["raw_path"])
	}
}

func TestEnrichResultLeavesUnmatchedPathsUntouched(t *testing.T) {
	withEnterpriseResolver(t, "oschina")

	input := map[string]any{"id": float64(1), "name": "bug"}
	result := EnrichResult("/1/labels", input)
	item, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", result)
	}
	if _, exists := item["html_url"]; exists {
		t.Fatalf("html_url must not be added for unmatched paths: %#v", item)
	}
	if item["name"] != "bug" {
		t.Fatalf("payload changed: %#v", item)
	}
}

func TestEnrichResultSkipsWhenEnterprisePathUnknown(t *testing.T) {
	withEnterpriseResolver(t, "")

	result := EnrichResult("/1/issues", map[string]any{"data": []any{map[string]any{"ident": "I1"}}})
	root := result.(map[string]any)
	item := root["data"].([]any)[0].(map[string]any)
	if _, exists := item["html_url"]; exists {
		t.Fatalf("html_url must not be added without enterprise path: %#v", item)
	}
}

func TestWarmEnterprisePathsLoadsCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"data":[{"id":1,"name":"oschina","path":"oschina"}]}`))
	}))
	defer server.Close()

	previousBase := GetApiBase()
	previousToken := GetGiteeAccessToken()
	SetApiBase(server.URL)
	SetGiteeAccessToken("test-token")
	ResetEnterprisePathCache()
	t.Cleanup(func() {
		SetApiBase(previousBase)
		SetGiteeAccessToken(previousToken)
		ResetEnterprisePathCache()
	})

	WarmEnterprisePaths()
	if got := EnterprisePath(1); got != "oschina" {
		t.Fatalf("EnterprisePath() = %q, want oschina", got)
	}
	if got := EnterprisePath(2); got != "" {
		t.Fatalf("EnterprisePath(2) = %q, want empty", got)
	}
}

func TestHandleMCPResultEnrichesIssueList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"data":[{"id":1,"ident":"I5DFIB","title":"demo"}]}`))
	}))
	defer server.Close()

	previousBase := GetApiBase()
	previousToken := GetGiteeAccessToken()
	SetApiBase(server.URL)
	SetGiteeAccessToken("test-token")
	urlbuilder.SetWebBase("https://e.gitee.com")
	ResetEnterprisePathCache()
	SetEnterprisePath(7, "oschina")
	t.Cleanup(func() {
		SetApiBase(previousBase)
		SetGiteeAccessToken(previousToken)
		urlbuilder.SetWebBase("")
		ResetEnterprisePathCache()
	})

	client := NewGiteeClient("GET", "/7/issues", WithQuery(map[string]interface{}{}))
	var payload map[string]any
	result, err := client.HandleMCPResult(&payload)
	if err != nil {
		t.Fatalf("HandleMCPResult() error = %v", err)
	}
	if result.IsError {
		t.Fatalf("result = %#v", result)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"html_url":"https://e.gitee.com/oschina/issues/list?issue=I5DFIB"`) {
		t.Fatalf("unexpected payload: %s", text)
	}
}

func TestEnrichResultAddsDocumentURLs(t *testing.T) {
	withEnterpriseResolver(t, "hightest")

	result := EnrichResult("/1/doc_nodes/tile", map[string]any{
		"total_count": 2,
		"documents": []any{
			map[string]any{"document_id": float64(8512977), "document_node_id": float64(4168359)},
			map[string]any{"document_id": float64(20), "document_node_id": float64(10), "program": map[string]any{"id": float64(776601)}},
		},
	})
	root := result.(map[string]any)
	items := root["documents"].([]any)

	if got := items[0].(map[string]any)["html_url"]; got != "https://e.gitee.com/hightest/docs/4168359/file/8512977" {
		t.Fatalf("enterprise doc url = %v", got)
	}
	if got := items[1].(map[string]any)["html_url"]; got != "https://e.gitee.com/hightest/projects/776601/docs/10/file/20" {
		t.Fatalf("program doc url = %v", got)
	}
}

func TestEnterprisePathResolvesOnDemand(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":1,"path":"oschina"}`))
	}))
	defer server.Close()

	previousBase := GetApiBase()
	previousToken := GetGiteeAccessToken()
	SetApiBase(server.URL)
	SetGiteeAccessToken("test-token")
	ResetEnterprisePathCache()
	EnableEnterprisePathAutoResolve(true)
	t.Cleanup(func() {
		EnableEnterprisePathAutoResolve(false)
		SetApiBase(previousBase)
		SetGiteeAccessToken(previousToken)
		ResetEnterprisePathCache()
	})

	for i := 0; i < 3; i++ {
		if got := EnterprisePath(1); got != "oschina" {
			t.Fatalf("EnterprisePath() = %q, want oschina", got)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("resolver calls = %d, want 1", got)
	}
}

func TestEnrichResultCoversGroupAndProgramSubPages(t *testing.T) {
	withEnterpriseResolver(t, "hightest")

	group := EnrichResult("/1/groups/loris09", map[string]any{"id": float64(1), "path": "loris09"})
	if got := group.(map[string]any)["html_url"]; got != "https://e.gitee.com/hightest/members/team/loris09" {
		t.Fatalf("group url = %v", got)
	}

	overview := EnrichResult("/1/programs/776601/overview", map[string]any{"bugs_count": float64(0)})
	if got := overview.(map[string]any)["html_url"]; got != "https://e.gitee.com/hightest/projects/776601/overview" {
		t.Fatalf("program overview url = %v", got)
	}

	projects := EnrichResult("/1/groups/loris09/projects", map[string]any{
		"data": []any{map[string]any{"path_with_namespace": "hightest/sdff"}},
	})
	item := projects.(map[string]any)["data"].([]any)[0].(map[string]any)
	if got := item["html_url"]; got != "https://e.gitee.com/hightest/repos/hightest/sdff" {
		t.Fatalf("group project url = %v", got)
	}

	planCases := EnrichResult("/1/test_plans/21354/test_cases", map[string]any{
		"data": []any{map[string]any{"id": float64(9), "ident": "TLKSX", "program_id": float64(776601)}},
	})
	planCase := planCases.(map[string]any)["data"].([]any)[0].(map[string]any)
	if got := planCase["html_url"]; got != "https://e.gitee.com/hightest/projects/776601/tests/test_plans/21354?test_case=TLKSX" {
		t.Fatalf("test plan case url = %v", got)
	}
}

func TestEnrichResultAddsIssueOverviewURLFromPath(t *testing.T) {
	withEnterpriseResolver(t, "hightest")

	result := EnrichResult("/1/issues/IKT2F8/overview", map[string]any{"children_count": float64(1)})
	if got := result.(map[string]any)["html_url"]; got != "https://e.gitee.com/hightest/issues/list?issue=IKT2F8" {
		t.Fatalf("issue overview url = %v", got)
	}

	numeric := EnrichResult("/1/issues/12345/overview", map[string]any{"children_count": float64(1)})
	if _, exists := numeric.(map[string]any)["html_url"]; exists {
		t.Fatalf("numeric issue id must not produce a url: %#v", numeric)
	}
}
