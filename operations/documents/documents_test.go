package documents

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

func toolRequest(arguments map[string]interface{}) mcp.CallToolRequest {
	request := mcp.CallToolRequest{}
	request.Params.Arguments = arguments
	return request
}

func toolResultJSON(t *testing.T, result *mcp.CallToolResult, target interface{}) {
	t.Helper()
	if result == nil {
		t.Fatal("tool result is nil")
	}
	if result.IsError {
		t.Fatalf("tool returned an error result: %#v", result.Content)
	}
	if len(result.Content) != 1 {
		t.Fatalf("tool content count = %d, want 1", len(result.Content))
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("tool content type = %T, want mcp.TextContent", result.Content[0])
	}
	if err := json.Unmarshal([]byte(content.Text), target); err != nil {
		t.Fatalf("decode tool result %q: %v", content.Text, err)
	}
}

func documentTestOptions(server *httptest.Server) []utils.Option {
	return []utils.Option{
		utils.WithApiBase(server.URL),
		utils.WithHTTPClient(server.Client()),
		utils.WithAuthorizer(&utils.CookieAuthorizer{Cookie: "session=test"}),
	}
}

func TestListDocumentsNormalizesIDsAndQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/7/doc_nodes/tile" {
			t.Fatalf("request = %s %s, want GET /7/doc_nodes/tile", req.Method, req.URL.Path)
		}
		wantQuery := map[string][]string{
			"direction":  {"asc"},
			"file_type":  {"WikiInfo"},
			"page":       {"2"},
			"per_page":   {"10"},
			"program_id": {"9"},
			"scope":      {"create_by_me"},
			"sort":       {"name"},
		}
		if got := map[string][]string(req.URL.Query()); !reflect.DeepEqual(got, wantQuery) {
			t.Errorf("query = %#v, want %#v", got, wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":91,"parent_id":0,"name":"Handbook","file_path":"Enterprise/Handbook","file_type":"WikiInfo","file_id":42,"editor_type":"markdown","created_at":"created","updated_at":"updated"}]}`)
	}))
	defer server.Close()

	result, err := ListDocumentsHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"program_id":    9,
		"created_by_me": true,
		"sort":          "name",
		"direction":     "asc",
		"page":          2,
		"per_page":      10,
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("ListDocumentsHandleFunc() error = %v", err)
	}

	response := types.DocumentListResponse{}
	toolResultJSON(t, result, &response)
	if response.TotalCount != 1 || len(response.Documents) != 1 {
		t.Fatalf("response = %#v, want one document", response)
	}
	document := response.Documents[0]
	if document.DocumentID != 42 || document.DocumentNodeID != 91 {
		t.Fatalf("document IDs = (%d, %d), want (42, 91)", document.DocumentID, document.DocumentNodeID)
	}
}

func TestListDocumentsSearchMapsPageHits(t *testing.T) {
	long := strings.Repeat("内", 400)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/7/doc_nodes/search" {
			t.Fatalf("request = %s %s, want GET /7/doc_nodes/search", req.Method, req.URL.Path)
		}
		wantQuery := map[string][]string{"page": {"2"}, "search": {"部署"}, "sort": {"name"}}
		if got := map[string][]string(req.URL.Query()); !reflect.DeepEqual(got, wantQuery) {
			t.Errorf("query = %#v, want %#v", got, wantQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":3,"data":[`+
			`{"id":11,"name":"<em class=\"gitee-highlight-keyword\">部署</em>流程","content":"a\n b\t<em>c</em>","info_id":42,"info_name":"发布手册","doc_path":"Release/发布手册","doc_id":91,"creator":{"id":5,"username":"alice","name":"Alice"},"last_updated_at":"2026-08-30 10:00"},`+
			`{"id":12,"name":null,"content":null,"info_id":42,"info_name":"发布手册","doc_path":"Release/发布手册","doc_id":null,"creator":null,"last_updated_at":"2026-08-31 09:00"},`+
			`{"id":13,"name":"ops","content":"`+long+`","info_id":43,"info_name":"Runbook","doc_path":"Ops/Runbook","doc_id":92,"creator":null,"last_updated_at":"2026-08-31 11:00"}]}`)
	}))
	defer server.Close()

	result, err := ListDocumentsHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"search":        "  部署  ",
		"sort":          "name",
		"page":          2,
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("ListDocumentsHandleFunc() error = %v", err)
	}

	response := types.DocumentSearchResponse{}
	toolResultJSON(t, result, &response)
	if response.Search != "部署" || response.TotalCount != 3 || len(response.Hits) != 3 {
		t.Fatalf("response = %#v", response)
	}
	first := response.Hits[0]
	if first.DocumentID != 42 || first.DocumentNodeID != 91 || first.PageID != 11 {
		t.Fatalf("first hit IDs = %#v", first)
	}
	if first.DocumentName != "发布手册" || first.PageName != "部署流程" || first.ContentPreview != "a b c" {
		t.Fatalf("first hit fields = %#v, want highlights stripped", first)
	}
	if first.Creator == nil || first.Creator.Username != "alice" {
		t.Fatalf("first hit creator = %#v", first.Creator)
	}

	second := response.Hits[1]
	if second.PageName != "" || second.ContentPreview != "" || second.DocumentNodeID != 0 || second.Creator != nil {
		t.Fatalf("second hit = %#v, want null-tolerant normalization", second)
	}

	preview := []rune(response.Hits[2].ContentPreview)
	if len(preview) != 301 || string(preview[300]) != "…" {
		t.Fatalf("content preview runes = %d, want 301 including trailing ellipsis", len(preview))
	}
}

func TestListDocumentsSearchRejectsUnsupportedParameters(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCount.Add(1)
	}))
	defer server.Close()

	cases := []struct {
		name string
		key  string
	}{
		{name: "program_id", key: "program_id"},
		{name: "created_by_me", key: "created_by_me"},
		{name: "direction", key: "direction"},
		{name: "per_page", key: "per_page"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			arguments := map[string]interface{}{"enterprise_id": 7, "search": "kw"}
			switch test.key {
			case "program_id":
				arguments["program_id"] = 9
			case "created_by_me":
				arguments["created_by_me"] = true
			case "direction":
				arguments["direction"] = "desc"
			case "per_page":
				arguments["per_page"] = 20
			}
			result, err := ListDocumentsHandleFunc(context.Background(), toolRequest(arguments), documentTestOptions(server)...)
			if !utils.IsParamError(err) {
				t.Fatalf("error = %v, want parameter error", err)
			}
			if result == nil || !result.IsError {
				t.Fatalf("result = %#v, want MCP error result", result)
			}
			if got := requestCount.Load(); got != 0 {
				t.Fatalf("request count = %d, want 0", got)
			}
		})
	}
}

func TestListDocumentsBlankSearchFallsBackToTile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/7/doc_nodes/tile" {
			t.Fatalf("path = %s, want /7/doc_nodes/tile", req.URL.Path)
		}
		if req.URL.Query().Has("search") {
			t.Errorf("query = %s, want no search parameter", req.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	}))
	defer server.Close()

	result, err := ListDocumentsHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"search":        "   ",
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("ListDocumentsHandleFunc() error = %v", err)
	}
	response := types.DocumentListResponse{}
	toolResultJSON(t, result, &response)
}

func TestListDocumentPagesFlattensTree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/7/docs/42/files" {
			t.Fatalf("request = %s %s, want GET /7/docs/42/files", req.Method, req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":10,"name":"Guides","parent_id":0,"file_type":2,"children":[{"id":11,"name":"Setup","parent_id":10,"file_type":1,"version":"v1","content_type":"markdown"}]}]}`)
	}))
	defer server.Close()

	result, err := ListDocumentPagesHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"document_id":   42,
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("ListDocumentPagesHandleFunc() error = %v", err)
	}

	response := types.DocumentPagesResponse{}
	toolResultJSON(t, result, &response)
	if response.TotalCount != 2 || len(response.Pages) != 2 {
		t.Fatalf("response = %#v, want two flattened entries", response)
	}
	if response.Pages[0].Path != "Guides" || response.Pages[0].Type != "folder" {
		t.Fatalf("root page = %#v", response.Pages[0])
	}
	if response.Pages[1].PageID != 11 || response.Pages[1].Path != "Guides/Setup" || response.Pages[1].Type != "page" {
		t.Fatalf("child page = %#v", response.Pages[1])
	}
}

func TestGetDocumentPageNormalizesPageIDAndControlsHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/7/docs/42/files/11" {
			t.Fatalf("path = %s, want /7/docs/42/files/11", req.URL.Path)
		}
		if got := req.URL.Query().Get("version_id"); got != "version-1" {
			t.Errorf("version_id = %q, want version-1", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":11,"name":"Setup","parent_id":10,"file_type":1,"version":"version-1","content_type":"markdown","content":"# Setup","content_html":"<h1>Setup</h1>"}`)
	}))
	defer server.Close()

	for _, includeHTML := range []bool{false, true} {
		t.Run(map[bool]string{false: "without_html", true: "with_html"}[includeHTML], func(t *testing.T) {
			result, err := GetDocumentPageHandleFunc(context.Background(), toolRequest(map[string]interface{}{
				"enterprise_id":         7,
				"document_id":           42,
				"page_id":               11,
				"version":               "version-1",
				"include_rendered_html": includeHTML,
			}), documentTestOptions(server)...)
			if err != nil {
				t.Fatalf("GetDocumentPageHandleFunc() error = %v", err)
			}

			var response map[string]interface{}
			toolResultJSON(t, result, &response)
			page := response["page"].(map[string]interface{})
			if page["page_id"] != float64(11) {
				t.Fatalf("page_id = %#v, want 11", page["page_id"])
			}
			_, hasHTML := page["content_html"]
			if hasHTML != includeHTML {
				t.Fatalf("content_html presence = %v, want %v", hasHTML, includeHTML)
			}
		})
	}
}

func TestUpdateDocumentPageRejectsStaleVersion(t *testing.T) {
	var putCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPut {
			putCount.Add(1)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":11,"name":"Setup","file_type":1,"version":"current"}`)
	}))
	defer server.Close()

	result, err := UpdateDocumentPageHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id":    7,
		"document_id":      42,
		"page_id":          11,
		"content":          "new content",
		"expected_version": "stale",
		"commit_message":   "Update setup",
	}), documentTestOptions(server)...)
	if !utils.IsParamError(err) {
		t.Fatalf("UpdateDocumentPageHandleFunc() error = %v, want parameter error", err)
	}
	if result == nil || !result.IsError {
		t.Fatalf("result = %#v, want MCP error result", result)
	}
	if got := putCount.Load(); got != 0 {
		t.Fatalf("PUT count = %d, want 0", got)
	}
}

func TestUpdateDocumentPageMapsPayload(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case http.MethodGet:
			_, _ = io.WriteString(w, `{"id":11,"name":"Setup","file_type":1,"version":"current"}`)
		case http.MethodPut:
			var payload map[string]interface{}
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			want := map[string]interface{}{
				"content":            "new content",
				"message":            "Update setup",
				"mentioned_user_ids": "5,7",
				"name":               "Installation",
			}
			if !reflect.DeepEqual(payload, want) {
				t.Errorf("payload = %#v, want %#v", payload, want)
			}
			_, _ = io.WriteString(w, `{"id":11,"name":"Installation","file_type":1,"version":"updated","content":"new content","content_html":"<p>new content</p>"}`)
		default:
			t.Fatalf("method = %s, want GET or PUT", req.Method)
		}
	}))
	defer server.Close()

	result, err := UpdateDocumentPageHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id":      7,
		"document_id":        42,
		"page_id":            11,
		"content":            "new content",
		"expected_version":   "current",
		"commit_message":     "Update setup",
		"page_name":          "Installation",
		"mentioned_user_ids": []interface{}{float64(5), float64(7)},
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("UpdateDocumentPageHandleFunc() error = %v", err)
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}

	response := types.DocumentPageResponse{}
	toolResultJSON(t, result, &response)
	if response.Page.PageID != 11 || response.Page.Version != "updated" || response.Page.ContentHTML != "" {
		t.Fatalf("response page = %#v", response.Page)
	}
}

func TestUploadDocumentSendsMultipartFilesAndRefreshesPages(t *testing.T) {
	root := t.TempDir()
	t.Setenv(utils.AllowedFileRootsEnv, root)
	markdownPath := filepath.Join(root, "guide.MD")
	textPath := filepath.Join(root, "notes.txt")
	if err := os.WriteFile(markdownPath, []byte("# Guide"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(textPath, []byte("Notes"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case req.Method == http.MethodPost && req.URL.Path == "/7/docs/42/upload":
			if err := req.ParseMultipartForm(20 << 20); err != nil {
				t.Fatal(err)
			}
			files := req.MultipartForm.File["files[]"]
			if len(files) != 2 {
				t.Fatalf("multipart files = %d, want 2", len(files))
			}
			want := map[string]string{"guide.md": "# Guide", "notes.txt": "Notes"}
			for _, header := range files {
				file, err := header.Open()
				if err != nil {
					t.Fatal(err)
				}
				body, readErr := io.ReadAll(file)
				_ = file.Close()
				if readErr != nil {
					t.Fatal(readErr)
				}
				if got := string(body); got != want[header.Filename] {
					t.Errorf("file %q content = %q, want %q", header.Filename, got, want[header.Filename])
				}
			}
			w.WriteHeader(http.StatusNoContent)
		case req.Method == http.MethodGet && req.URL.Path == "/7/docs/42/files":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":11,"name":"Guide","file_type":1,"version":"v1"}]}`)
		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
	}))
	defer server.Close()

	result, err := UploadDocumentHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"document_id":   42,
		"file_paths":    []interface{}{markdownPath, textPath},
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("UploadDocumentHandleFunc() error = %v", err)
	}

	response := types.DocumentUploadResponse{}
	toolResultJSON(t, result, &response)
	if !response.Imported || len(response.Pages) != 1 || response.Pages[0].PageID != 11 {
		t.Fatalf("response = %#v", response)
	}
}

func TestDownloadDocumentSavesAndProtectsExistingFile(t *testing.T) {
	root := t.TempDir()
	t.Setenv(utils.AllowedFileRootsEnv, root)
	var exportCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/7/docs/42/export" {
			t.Fatalf("request = %s %s, want GET /7/docs/42/export", req.Method, req.URL.Path)
		}
		if got := req.URL.Query().Get("export_type"); got != "markdown" {
			t.Errorf("export_type = %q, want markdown", got)
		}
		count := exportCount.Add(1)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="handbook.zip"`)
		_, _ = io.WriteString(w, "archive-"+string(rune('0'+count)))
	}))
	defer server.Close()

	arguments := map[string]interface{}{
		"enterprise_id":    7,
		"document_id":      42,
		"format":           "markdown",
		"output_directory": root,
	}
	result, err := DownloadDocumentHandleFunc(context.Background(), toolRequest(arguments), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("DownloadDocumentHandleFunc() error = %v", err)
	}
	response := types.DocumentDownloadResponse{}
	toolResultJSON(t, result, &response)
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if response.Filename != "handbook.zip" || response.SavedPath != filepath.Join(resolvedRoot, "handbook.zip") || response.Size != 9 {
		t.Fatalf("response = %#v", response)
	}
	body, err := os.ReadFile(response.SavedPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "archive-1" {
		t.Fatalf("saved content = %q, want archive-1", got)
	}

	result, err = DownloadDocumentHandleFunc(context.Background(), toolRequest(arguments), documentTestOptions(server)...)
	if !utils.IsParamError(err) || result == nil || !result.IsError {
		t.Fatalf("second download = (%#v, %v), want overwrite parameter error", result, err)
	}
	body, err = os.ReadFile(response.SavedPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "archive-1" {
		t.Fatalf("existing content = %q after rejected overwrite, want archive-1", got)
	}

	arguments["overwrite"] = true
	result, err = DownloadDocumentHandleFunc(context.Background(), toolRequest(arguments), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("overwrite download error = %v", err)
	}
	toolResultJSON(t, result, &response)
	body, err = os.ReadFile(response.SavedPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body); got != "archive-3" {
		t.Fatalf("overwritten content = %q, want archive-3", got)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".mcp-gitee-ent-download-") {
			t.Fatalf("temporary download file was not cleaned up: %s", entry.Name())
		}
	}
}

func TestCreateDocumentPostsPayloadAndMapsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/7/doc_nodes/create" {
			t.Fatalf("request = %s %s, want POST /7/doc_nodes/create", req.Method, req.URL.Path)
		}
		if got := req.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content type = %q, want application/json", got)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]interface{}{
			"name":            "Handbook",
			"parent_id":       float64(12),
			"program_id":      float64(9),
			"auth_type":       "read_write",
			"scrum_sprint_id": float64(4),
		}
		got := map[string]interface{}{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request body %q: %v", body, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("payload = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":91,"parent_id":12,"name":"Handbook","file_path":"企业知识库/Handbook","file_type":"WikiInfo","file_id":42,"editor_type":"markdown","created_at":"created","updated_at":"updated"}`)
	}))
	defer server.Close()

	result, err := CreateDocumentHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id":   7,
		"name":            "  Handbook  ",
		"parent_id":       12,
		"program_id":      9,
		"auth_type":       "read_write",
		"scrum_sprint_id": 4,
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("CreateDocumentHandleFunc() error = %v", err)
	}

	document := types.Document{}
	toolResultJSON(t, result, &document)
	if document.DocumentID != 42 || document.DocumentNodeID != 91 || document.Name != "Handbook" {
		t.Fatalf("document = %#v, want IDs (42, 91) and name Handbook", document)
	}
}

func TestCreateDocumentSendsOnlyNameAndDefaultAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/7/doc_nodes/create" {
			t.Fatalf("request = %s %s, want POST /7/doc_nodes/create", req.Method, req.URL.Path)
		}
		body, _ := io.ReadAll(req.Body)
		want := map[string]interface{}{"name": "Runbook", "auth_type": "inherit"}
		got := map[string]interface{}{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request body %q: %v", body, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("payload = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":92,"parent_id":0,"name":"Runbook","file_type":"WikiInfo","file_id":43}`)
	}))
	defer server.Close()

	result, err := CreateDocumentHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"name":          "Runbook",
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("CreateDocumentHandleFunc() error = %v", err)
	}
	document := types.Document{}
	toolResultJSON(t, result, &document)
	if document.DocumentID != 43 || document.DocumentNodeID != 92 {
		t.Fatalf("document IDs = (%d, %d), want (43, 92)", document.DocumentID, document.DocumentNodeID)
	}
}

func TestCreateDocumentValidatesParameters(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCount.Add(1)
	}))
	defer server.Close()

	cases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{name: "missing enterprise id", arguments: map[string]interface{}{"name": "Doc"}},
		{name: "missing name", arguments: map[string]interface{}{"enterprise_id": 7}},
		{name: "blank name", arguments: map[string]interface{}{"enterprise_id": 7, "name": "   "}},
		{name: "invalid auth type", arguments: map[string]interface{}{"enterprise_id": 7, "name": "Doc", "auth_type": "shared"}},
		{name: "sprint without program", arguments: map[string]interface{}{"enterprise_id": 7, "name": "Doc", "scrum_sprint_id": 4}},
		{name: "zero parent id", arguments: map[string]interface{}{"enterprise_id": 7, "name": "Doc", "parent_id": 0}},
		{name: "zero program id", arguments: map[string]interface{}{"enterprise_id": 7, "name": "Doc", "program_id": 0}},
		{name: "zero sprint id", arguments: map[string]interface{}{"enterprise_id": 7, "name": "Doc", "program_id": 9, "scrum_sprint_id": 0}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result, err := CreateDocumentHandleFunc(context.Background(), toolRequest(test.arguments), documentTestOptions(server)...)
			if !utils.IsParamError(err) {
				t.Fatalf("error = %v, want parameter error", err)
			}
			if result == nil || !result.IsError {
				t.Fatalf("result = %#v, want MCP error result", result)
			}
		})
	}
	if got := requestCount.Load(); got != 0 {
		t.Fatalf("request count = %d, want 0", got)
	}
}

func TestCreateDocumentNodeCreatesPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/7/docs/42/create_file" {
			t.Fatalf("request = %s %s, want POST /7/docs/42/create_file", req.Method, req.URL.Path)
		}
		body, _ := io.ReadAll(req.Body)
		want := map[string]interface{}{"name": "Setup", "parent_id": float64(3)}
		got := map[string]interface{}{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request body %q: %v", body, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("payload = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":11,"name":"Setup","parent_id":3,"file_type":1,"version":"v1","content_type":"markdown","content":"","content_html":"<p></p>"}`)
	}))
	defer server.Close()

	result, err := CreateDocumentNodeHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"document_id":   42,
		"type":          "page",
		"name":          "  Setup  ",
		"parent_id":     3,
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("CreateDocumentNodeHandleFunc() error = %v", err)
	}

	response := types.DocumentPageResponse{}
	toolResultJSON(t, result, &response)
	if response.DocumentID != 42 {
		t.Fatalf("document_id = %d, want 42", response.DocumentID)
	}
	page := response.Page
	if page.PageID != 11 || page.Name != "Setup" || page.ParentPageID != 3 || page.Type != "page" || page.Version != "v1" {
		t.Fatalf("page = %#v", page)
	}
	if page.ContentHTML != "" {
		t.Fatalf("content_html = %q, want stripped", page.ContentHTML)
	}
}

func TestCreateDocumentNodeCreatesFolder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.URL.Path != "/7/docs/42/create_folder" {
			t.Fatalf("request = %s %s, want POST /7/docs/42/create_folder", req.Method, req.URL.Path)
		}
		body, _ := io.ReadAll(req.Body)
		want := map[string]interface{}{"name": "Guides"}
		got := map[string]interface{}{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request body %q: %v", body, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("payload = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":13,"name":"Guides","parent_id":0,"file_type":2,"content_type":"markdown"}`)
	}))
	defer server.Close()

	result, err := CreateDocumentNodeHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"document_id":   42,
		"type":          "folder",
		"name":          "Guides",
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("CreateDocumentNodeHandleFunc() error = %v", err)
	}

	response := types.DocumentPageResponse{}
	toolResultJSON(t, result, &response)
	if response.Page.PageID != 13 || response.Page.Type != "folder" {
		t.Fatalf("page = %#v, want folder 13", response.Page)
	}
}

func TestCreateDocumentNodeValidatesParameters(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCount.Add(1)
	}))
	defer server.Close()

	cases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{name: "missing enterprise", arguments: map[string]interface{}{"document_id": 42, "type": "page", "name": "Doc"}},
		{name: "missing document", arguments: map[string]interface{}{"enterprise_id": 7, "type": "page", "name": "Doc"}},
		{name: "missing type", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "name": "Doc"}},
		{name: "invalid type", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "type": "link", "name": "Doc"}},
		{name: "missing name", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "type": "folder"}},
		{name: "blank name", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "type": "folder", "name": "  "}},
		{name: "zero parent", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "type": "page", "name": "Doc", "parent_id": 0}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result, err := CreateDocumentNodeHandleFunc(context.Background(), toolRequest(test.arguments), documentTestOptions(server)...)
			if !utils.IsParamError(err) {
				t.Fatalf("error = %v, want parameter error", err)
			}
			if result == nil || !result.IsError {
				t.Fatalf("result = %#v, want MCP error result", result)
			}
		})
	}
	if got := requestCount.Load(); got != 0 {
		t.Fatalf("request count = %d, want 0", got)
	}
}

func TestRenameDocumentFolderPutsPayloadAndMapsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut || req.URL.Path != "/7/docs/42/files/13/rename" {
			t.Fatalf("request = %s %s, want PUT /7/docs/42/files/13/rename", req.Method, req.URL.Path)
		}
		body, _ := io.ReadAll(req.Body)
		want := map[string]interface{}{"name": "Guides v2"}
		got := map[string]interface{}{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode request body %q: %v", body, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("payload = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":13,"name":"Guides v2","parent_id":0,"file_type":2,"content_type":"markdown"}`)
	}))
	defer server.Close()

	result, err := RenameDocumentFolderHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"document_id":   42,
		"page_id":       13,
		"name":          "  Guides v2  ",
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("RenameDocumentFolderHandleFunc() error = %v", err)
	}

	response := types.DocumentPageResponse{}
	toolResultJSON(t, result, &response)
	if response.DocumentID != 42 || response.Page.PageID != 13 || response.Page.Type != "folder" || response.Page.Name != "Guides v2" {
		t.Fatalf("response = %#v, want folder 13 renamed", response)
	}
}

func TestRenameDocumentFolderValidatesParameters(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCount.Add(1)
	}))
	defer server.Close()

	cases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{name: "missing enterprise", arguments: map[string]interface{}{"document_id": 42, "page_id": 13, "name": "Guides"}},
		{name: "missing document", arguments: map[string]interface{}{"enterprise_id": 7, "page_id": 13, "name": "Guides"}},
		{name: "missing page", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "name": "Guides"}},
		{name: "missing name", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "page_id": 13}},
		{name: "blank name", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "page_id": 13, "name": "   "}},
		{name: "name with slash", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "page_id": 13, "name": "a/b"}},
		{name: "name too long", arguments: map[string]interface{}{"enterprise_id": 7, "document_id": 42, "page_id": 13, "name": strings.Repeat("名", 65)}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result, err := RenameDocumentFolderHandleFunc(context.Background(), toolRequest(test.arguments), documentTestOptions(server)...)
			if !utils.IsParamError(err) {
				t.Fatalf("error = %v, want parameter error", err)
			}
			if result == nil || !result.IsError {
				t.Fatalf("result = %#v, want MCP error result", result)
			}
		})
	}

	if got := requestCount.Load(); got != 0 {
		t.Fatalf("request count = %d, want 0", got)
	}
}

func TestRenameDocumentFolderAcceptsMaxLengthName(t *testing.T) {
	name := strings.Repeat("名", 64)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut || req.URL.Path != "/7/docs/42/files/13/rename" {
			t.Fatalf("request = %s %s, want PUT /7/docs/42/files/13/rename", req.Method, req.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":13,"name":"`+name+`","parent_id":0,"file_type":2}`)
	}))
	defer server.Close()

	result, err := RenameDocumentFolderHandleFunc(context.Background(), toolRequest(map[string]interface{}{
		"enterprise_id": 7,
		"document_id":   42,
		"page_id":       13,
		"name":          name,
	}), documentTestOptions(server)...)
	if err != nil {
		t.Fatalf("RenameDocumentFolderHandleFunc() error = %v", err)
	}
	response := types.DocumentPageResponse{}
	toolResultJSON(t, result, &response)
	if response.Page.Name != name {
		t.Fatalf("name = %q, want 64-rune name", response.Page.Name)
	}
}
