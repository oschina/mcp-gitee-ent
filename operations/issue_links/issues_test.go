package issue_links

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGetIssueOverviewUsesIdentQuery(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/issues/I123/overview" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.URL.Query().Get("qt") != "ident" {
			t.Errorf("qt = %q, want ident", request.URL.Query().Get("qt"))
		}
		_, _ = io.WriteString(w, `{"children_count":2,"link_pull_request_count":1}`)
	})

	result := call(t, GetIssueOverviewHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": "I123"})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListIssueChildrenPassesFilters(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/12/children" {
			t.Errorf("path = %q", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("issue_type") != "task" || query.Get("page") != "2" || query.Get("per_page") != "50" {
			t.Errorf("query = %v", query)
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListIssueChildrenHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      12,
		"issue_type":    "task",
		"page":          2,
		"per_page":      50,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListIssueAncestorsPassesScope(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/12/ancestors" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("scope") != "all" {
			t.Errorf("scope = %q", request.URL.Query().Get("scope"))
		}
		_, _ = io.WriteString(w, `{"total_count":0}`)
	})

	result := call(t, ListIssueAncestorsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      12,
		"scope":         "all",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestIssueHandlersRequireIssueID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent when issue_id is missing")
	})

	result := call(t, GetIssueOverviewHandleFunc, map[string]any{"enterprise_id": 7})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestGetIssueDevelopmentLinksUsesTypePath(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/issues/I123/data_links/pull_requests" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("qt") != "ident" || query.Get("state") != "open" || query.Get("page") != "2" {
			t.Errorf("query = %v", query)
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, GetIssueDevelopmentLinksHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      "I123",
		"type":          "pull_requests",
		"state":         "open",
		"page":          2,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestGetIssueDevelopmentLinksRejectsUnknownType(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent for an invalid type")
	})

	result := call(t, GetIssueDevelopmentLinksHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      1,
		"type":          "pulls",
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestListIssueOperationLogsPassesCursor(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/12/operate_logs" {
			t.Errorf("path = %q", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("limit") != "20" || query.Get("prev_id") != "5" {
			t.Errorf("query = %v", query)
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListIssueOperationLogsHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      12,
		"limit":         20,
		"prev_id":       5,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListIssueNextStatesUsesIssueStatesPath(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/I9/issue_states" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("qt") != "ident" {
			t.Errorf("qt = %q", request.URL.Query().Get("qt"))
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListIssueNextStatesHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      "I9",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestUpdateIssueNotePutsBody(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut || request.URL.Path != "/7/issues/I123/notes/55" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.URL.Query().Get("qt") != "ident" {
			t.Errorf("qt = %q", request.URL.Query().Get("qt"))
		}
		body := map[string]any{}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["body"] != "updated" {
			t.Errorf("body = %#v", body)
		}
		_, _ = io.WriteString(w, `{"id":55}`)
	})

	result := call(t, UpdateIssueNoteHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      "I123",
		"note_id":       55,
		"body":          "updated",
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestDeleteIssueNoteUsesDeleteMethod(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete || request.URL.Path != "/7/issues/12/notes/55" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	result := call(t, DeleteIssueNoteHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      12,
		"note_id":       55,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestLinkIssuePullRequestUsesPost(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/7/issues/12/link_pull_request/88" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":88}`)
	})

	result := call(t, LinkIssuePullRequestHandleFunc, map[string]any{
		"enterprise_id":   7,
		"issue_id":        12,
		"pull_request_id": 88,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestUnlinkIssuePullRequestUsesDelete(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete || request.URL.Path != "/7/issues/12/link_pull_request/88" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	result := call(t, UnlinkIssuePullRequestHandleFunc, map[string]any{
		"enterprise_id":   7,
		"issue_id":        12,
		"pull_request_id": 88,
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestLinkIssuePullRequestRequiresPullRequestID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		t.Fatal("HTTP request must not be sent without pull_request_id")
	})
	result := call(t, LinkIssuePullRequestHandleFunc, map[string]any{
		"enterprise_id": 7,
		"issue_id":      12,
	})
	if !result.IsError {
		t.Fatalf("result = %#v, want error", result)
	}
}

func TestListIssueAttachments(t *testing.T) {
	var detailCalls, attachmentCalls int
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/issues/I123":
			detailCalls++
			if request.URL.Query().Get("qt") != "ident" {
				t.Errorf("qt = %q, want ident", request.URL.Query().Get("qt"))
			}
			_, _ = io.WriteString(w, `{"id":42,"ident":"I123"}`)
		case "/7/issues/42/paged_attach_files":
			attachmentCalls++
			if _, ok := request.URL.Query()["qt"]; ok {
				t.Errorf("attachment request must not send qt: %v", request.URL.Query())
			}
			if request.URL.Query().Get("page") != "2" {
				t.Errorf("page = %q, want 2", request.URL.Query().Get("page"))
			}
			_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
	})
	result := call(t, ListIssueAttachmentsHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": "I123", "page": 2})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
	if detailCalls != 1 || attachmentCalls != 1 {
		t.Fatalf("detailCalls = %d, attachmentCalls = %d, want 1/1", detailCalls, attachmentCalls)
	}
}

func TestListIssueRelatedIssues(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/12/link_issues" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("direction") != "pre" || request.URL.Query().Get("page") != "2" {
			t.Errorf("query = %v", request.URL.Query())
		}
		_, _ = io.WriteString(w, `{"data":[],"total_count":0}`)
	})
	result := call(t, ListIssueRelatedIssuesHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "direction": "pre", "page": 2})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestListIssueAuthorizations(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/I9/auths" {
			t.Errorf("path = %q", request.URL.Path)
		}
		if request.URL.Query().Get("qt") != "ident" {
			t.Errorf("qt = %q", request.URL.Query().Get("qt"))
		}
		_, _ = io.WriteString(w, `{"read":true}`)
	})
	result := call(t, ListIssueAuthorizationsHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": "I9"})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestIssueCollaborationWrites(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		body := map[string]any{}
		_ = json.NewDecoder(request.Body).Decode(&body)
		switch request.URL.Path {
		case "/7/issues/12/star":
			if request.Method != http.MethodPost && request.Method != http.MethodDelete {
				t.Errorf("star method = %s", request.Method)
			}
		case "/7/issues/12/reaction":
			if request.Method != http.MethodPost || body["text"] != "thumbsup" {
				t.Errorf("reaction = %s %#v", request.Method, body)
			}
		case "/7/reactions/9":
			if request.Method != http.MethodDelete {
				t.Errorf("remove reaction method = %s", request.Method)
			}
		case "/7/issues/12/related_issue":
			if request.Method != http.MethodPost || body["ref_type"] != "normal" {
				t.Errorf("relation add = %s %#v", request.Method, body)
			}
		case "/7/issues/12/related_issue/13":
			if request.Method == http.MethodPut {
				if body["ref_type"] != "finish_to_start" {
					t.Errorf("relation update = %#v", body)
				}
			} else if request.Method != http.MethodDelete {
				t.Errorf("relation remove method = %s", request.Method)
			}
		case "/7/issues/12/link_doc_nodes/5":
			if request.Method != http.MethodPost && request.Method != http.MethodDelete {
				t.Errorf("doc link method = %s", request.Method)
			}
		case "/7/issues/12/link_doc_nodes":
			if request.Method != http.MethodGet || request.URL.Query().Get("linked") != "true" {
				t.Errorf("linked docs = %s %v", request.Method, request.URL.Query())
			}
		case "/7/issues/12/test_plan_cases":
			switch request.Method {
			case http.MethodPost:
				if _, ok := body["test_plan_case_ids"]; !ok {
					t.Errorf("test case link body = %#v", body)
				}
			case http.MethodDelete:
				if request.URL.Query().Get("test_plan_case_id") == "" {
					t.Errorf("unlink query = %v, want test_plan_case_id", request.URL.Query())
				}
				if _, ok := body["test_plan_case_ids"]; ok {
					t.Errorf("unlink must not send the plural body: %#v", body)
				}
			default:
				t.Errorf("test case link method = %s", request.Method)
			}
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"id":12}`)
	})

	cases := []struct {
		handler   utils.OptionHandlerFunc
		arguments map[string]any
	}{
		{SetIssueStarredHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "starred": true}},
		{SetIssueStarredHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "starred": false}},
		{AddIssueReactionHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "text": "thumbsup"}},
		{RemoveIssueReactionHandleFunc, map[string]any{"enterprise_id": 7, "reaction_id": 9}},
		{AddIssueRelationHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "link_issue_id": 13, "ref_type": "normal"}},
		{UpdateIssueRelationHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "link_issue_id": 13, "ref_type": "finish_to_start"}},
		{RemoveIssueRelationHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "link_issue_id": 13}},
		{LinkIssueDocumentHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "doc_node_id": 5}},
		{UnlinkIssueDocumentHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "doc_node_id": 5}},
		{LinkIssueTestCasesHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "test_plan_case_ids": []any{1, 2}}},
		{UnlinkIssueTestCasesHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 12, "test_plan_case_ids": []any{1}}},
	}
	for _, tc := range cases {
		result := call(t, tc.handler, tc.arguments)
		if result.IsError {
			t.Fatalf("handler result is error: %#v", result.Content)
		}
	}
}

func TestListIssueAttachmentsKeepsNumericID(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/7/issues/42/paged_attach_files" {
			t.Errorf("path = %q, want /7/issues/42/paged_attach_files", request.URL.Path)
		}
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	})

	result := call(t, ListIssueAttachmentsHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": 42})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
}

func TestUnlinkIssueDocumentReturnsRemainingLinks(t *testing.T) {
	var deleteCalls, listCalls int
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/7/issues/I123/link_doc_nodes/5":
			deleteCalls++
			if request.Method != http.MethodDelete {
				t.Errorf("method = %s, want DELETE", request.Method)
			}
			if request.URL.Query().Get("qt") != "ident" {
				t.Errorf("qt = %q, want ident", request.URL.Query().Get("qt"))
			}
			w.WriteHeader(http.StatusNoContent)
		case "/7/issues/I123/link_doc_nodes":
			listCalls++
			if request.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", request.Method)
			}
			if request.URL.Query().Get("linked") != "true" || request.URL.Query().Get("qt") != "ident" {
				t.Errorf("query = %v", request.URL.Query())
			}
			_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":9,"name":"keep"}]}`)
		default:
			t.Errorf("unexpected path %s", request.URL.Path)
		}
	})

	result := call(t, UnlinkIssueDocumentHandleFunc, map[string]any{"enterprise_id": 7, "issue_id": "I123", "doc_node_id": 5})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
	var payload struct {
		DocNodeID            int  `json:"doc_node_id"`
		IsLinked             bool `json:"is_linked"`
		LinkedDocumentsTotal int  `json:"linked_documents_total"`
		LinkedDocuments      []struct {
			ID int `json:"id"`
		} `json:"linked_documents"`
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T", result.Content[0])
	}
	if err := json.Unmarshal([]byte(content.Text), &payload); err != nil {
		t.Fatalf("decode %q: %v", content.Text, err)
	}
	if payload.DocNodeID != 5 || payload.IsLinked || payload.LinkedDocumentsTotal != 1 ||
		len(payload.LinkedDocuments) != 1 || payload.LinkedDocuments[0].ID != 9 {
		t.Fatalf("payload = %#v", payload)
	}
	if deleteCalls != 1 || listCalls != 1 {
		t.Fatalf("deleteCalls = %d, listCalls = %d, want 1/1", deleteCalls, listCalls)
	}
}

func TestUnlinkIssueTestCasesUsesSingularParamPerID(t *testing.T) {
	unlinked := map[string]int{}
	withTestServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete || request.URL.Path != "/7/issues/I123/test_plan_cases" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
			return
		}
		if request.URL.Query().Get("qt") != "ident" {
			t.Errorf("qt = %q, want ident", request.URL.Query().Get("qt"))
		}
		unlinked[request.URL.Query().Get("test_plan_case_id")]++
		w.WriteHeader(http.StatusNoContent)
	})

	result := call(t, UnlinkIssueTestCasesHandleFunc, map[string]any{
		"enterprise_id": 7, "issue_id": "I123", "test_plan_case_ids": []any{2001, 2002},
	})
	if result.IsError {
		t.Fatalf("result is error: %#v", result.Content)
	}
	if unlinked["2001"] != 1 || unlinked["2002"] != 1 {
		t.Fatalf("unlink calls = %#v, want one call per id", unlinked)
	}
	content, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T", result.Content[0])
	}
	if !strings.Contains(content.Text, `"linked":false`) {
		t.Fatalf("ack = %s", content.Text)
	}
}
