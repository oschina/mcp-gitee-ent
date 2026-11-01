package discovery

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func okHandler(_ context.Context, _ mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("ok"), nil
}

func spec(name, summary, domain string, invokable bool, aliases ...string) catalog.Spec {
	return catalog.Spec{
		Tool:             mcp.NewTool(name, mcp.WithDescription(summary)),
		Handler:          okHandler,
		Domain:           domain,
		Tags:             []string{domain},
		Aliases:          aliases,
		Summary:          summary,
		Risk:             catalog.RiskRead,
		ReadOnly:         true,
		Verification:     catalog.VerificationNone,
		Invokable:        invokable,
		Visible:          true,
		ExampleArguments: map[string]any{"enterprise_id": 1},
	}
}

func testRegistry(t *testing.T) *catalog.Registry {
	t.Helper()
	registry := catalog.NewRegistry()
	err := registry.Register(
		spec("list_milestones", "List or search milestones.", "milestones", true, "里程碑"),
		spec("merge_enterprise_pull", "Merge a pull request.", "pulls", true, "pr", "mr", "合并请求"),
		spec("list_enterprise_issues", "List enterprise work items.", "issues", true, "任务", "需求", "bug"),
		spec("admin_only_tool", "Administrative operation.", "admin", false, "管理"),
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return registry
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		t.Fatalf("result = %#v, want text content", result)
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want mcp.TextContent", result.Content[0])
	}
	return text.Text
}

func TestLexicalSearchRanksExactNameFirst(t *testing.T) {
	searcher := NewLexicalSearcher(testRegistry(t))
	hits, err := searcher.Search(context.Background(), "merge_enterprise_pull", 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) == 0 || hits[0].Name != "merge_enterprise_pull" {
		t.Fatalf("hits = %#v, want merge_enterprise_pull first", hits)
	}
}

func TestLexicalSearchFindsChineseAlias(t *testing.T) {
	searcher := NewLexicalSearcher(testRegistry(t))
	hits, err := searcher.Search(context.Background(), "里程碑", 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) == 0 || hits[0].Name != "list_milestones" {
		t.Fatalf("hits = %#v, want list_milestones first", hits)
	}
}

func TestLexicalSearchSkipsNonInvokableTools(t *testing.T) {
	searcher := NewLexicalSearcher(testRegistry(t))
	hits, err := searcher.Search(context.Background(), "admin", 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	for _, hit := range hits {
		if hit.Name == "admin_only_tool" {
			t.Fatal("non-invokable tool appeared in search results")
		}
	}
}

func TestLexicalSearchHonorsLimitAndEmptyQuery(t *testing.T) {
	searcher := NewLexicalSearcher(testRegistry(t))
	hits, err := searcher.Search(context.Background(), "enterprise", 1)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) != 1 {
		t.Fatalf("len(hits) = %d, want 1", len(hits))
	}

	empty, err := searcher.Search(context.Background(), "   ", 5)
	if err != nil {
		t.Fatalf("Search(empty) error = %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("Search(empty) = %#v, want no hits", empty)
	}
}

func TestSearchToolsHandlerReturnsNameDescriptionAndRequired(t *testing.T) {
	registry := catalog.NewRegistry()
	spec := spec("list_milestones", "List or search milestones.", "milestones", true, "里程碑")
	spec.Tool = mcp.NewTool("list_milestones",
		mcp.WithDescription("List or search milestones."),
		mcp.WithNumber("enterprise_id", mcp.Required()),
		mcp.WithString("keyword"),
	)
	if err := registry.Register(spec); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	handler := NewSearchToolsHandler(registry, nil)
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"query": "里程碑"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	text := resultText(t, result)
	for _, expected := range []string{"list_milestones", "List or search milestones", "required: enterprise_id"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("search output %q does not contain %q", text, expected)
		}
	}
}

func TestSearchToolsHandlerRejectsMissingQuery(t *testing.T) {
	handler := NewSearchToolsHandler(testRegistry(t), nil)
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !result.IsError || !strings.Contains(resultText(t, result), "query") {
		t.Fatalf("result = %#v, want missing query error", result)
	}
}

func TestInspectToolHandlerReturnsSchemaAndExample(t *testing.T) {
	handler := NewInspectToolHandler(testRegistry(t))
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"name": "list_milestones"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	response := map[string]any{}
	if err := json.Unmarshal([]byte(resultText(t, result)), &response); err != nil {
		t.Fatalf("inspect output is not JSON: %v", err)
	}
	if response["name"] != "list_milestones" {
		t.Fatalf("name = %v", response["name"])
	}
	if _, ok := response["input_schema"]; !ok {
		t.Fatal("input_schema missing")
	}
	if _, ok := response["example_arguments"]; !ok {
		t.Fatal("example_arguments missing")
	}
}

func TestInvokeToolHandlerDecodesJSONStringArguments(t *testing.T) {
	registry := catalog.NewRegistry()
	var received map[string]any
	spec := spec("capture_tool", "Capture arguments.", "issues", true)
	spec.Handler = func(_ context.Context, request mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
		received, _ = utils.ConvertArgumentsToMap(request.Params.Arguments)
		return mcp.NewToolResultText("captured"), nil
	}
	if err := registry.Register(spec); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	handler := NewInvokeToolHandler(registry)
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"name":      "capture_tool",
		"arguments": `{"enterprise_id":7,"issue_id":9}`,
	}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if resultText(t, result) != "captured" {
		t.Fatalf("result text = %q", resultText(t, result))
	}
	if received["enterprise_id"] != float64(7) || received["issue_id"] != float64(9) {
		t.Fatalf("received = %#v", received)
	}
}

func TestInvokeToolHandlerRejectsRecursion(t *testing.T) {
	handler := NewInvokeToolHandler(testRegistry(t))
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"name": InvokeTool}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !result.IsError || !strings.Contains(resultText(t, result), "cannot call itself") {
		t.Fatalf("result = %#v, want recursion error", result)
	}
}

func TestInvokeToolHandlerRejectsDisabledTool(t *testing.T) {
	handler := NewInvokeToolHandler(testRegistry(t))
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"name": "admin_only_tool"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !result.IsError || !strings.Contains(resultText(t, result), "not enabled") {
		t.Fatalf("result = %#v, want not-enabled error", result)
	}
}

func TestInvokeToolHandlerSuggestsUnknownTool(t *testing.T) {
	handler := NewInvokeToolHandler(testRegistry(t))
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"name": "list_milestone"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "Unknown tool") || !strings.Contains(text, "search_tools") {
		t.Fatalf("result = %q, want unknown-tool guidance", text)
	}
}

type fixedSearcher struct {
	hits []Hit
}

func (s fixedSearcher) Search(context.Context, string, int) ([]Hit, error) {
	return s.hits, nil
}

func TestSearchToolsHandlerBackfillsWhenSearcherCapsResults(t *testing.T) {
	registry := catalog.NewRegistry()
	err := registry.Register(
		spec("list_test_reports", "List test reports.", "test_reports", true, "测试报告"),
		spec("create_test_report", "Create a test report.", "test_reports", true, "测试报告"),
		spec("list_enterprise_issues", "List enterprise work items.", "issues", true, "任务"),
	)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// The injected searcher ignores the requested limit and only returns a
	// candidate from another domain. Domain filtering would starve the result
	// set without a lexical backfill over the registry.
	handler := NewSearchToolsHandler(registry, fixedSearcher{hits: []Hit{{Name: "list_enterprise_issues"}}})
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"query": "测试报告", "domain": "test_reports", "limit": 5}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	text := resultText(t, result)
	for _, expected := range []string{"list_test_reports", "create_test_report"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("search output %q does not contain %q", text, expected)
		}
	}
	if strings.Contains(text, "No tools matched") {
		t.Fatalf("search output %q reported no match despite eligible backfill candidates", text)
	}
}

func TestSearchToolsHandlerDedupesSearcherAndBackfill(t *testing.T) {
	registry := catalog.NewRegistry()
	if err := registry.Register(spec("list_test_reports", "List test reports.", "test_reports", true, "测试报告")); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	handler := NewSearchToolsHandler(registry, fixedSearcher{hits: []Hit{
		{Name: "list_test_reports"},
		{Name: "list_test_reports"},
	}})
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"query": "测试报告"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if count := strings.Count(resultText(t, result), "list_test_reports"); count != 1 {
		t.Fatalf("list_test_reports appears %d times, want 1 in %q", count, resultText(t, result))
	}
}

func TestSearchToolsHandlerPreservesNoMatchWhenOnlyIneligibleToolsMatch(t *testing.T) {
	handler := NewSearchToolsHandler(testRegistry(t), fixedSearcher{hits: []Hit{{Name: "admin_only_tool"}}})
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"query": "admin"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	text := resultText(t, result)
	if !strings.Contains(text, "No tools matched") || strings.Contains(text, "admin_only_tool") {
		t.Fatalf("result = %q, want no-match without ineligible tool", text)
	}
}

func TestInspectToolHandlerRejectsFilteredTool(t *testing.T) {
	registry := testRegistry(t)
	filter := ToolFilterFunc(func(_ context.Context, spec catalog.Spec) bool {
		return spec.Tool.Name != "list_milestones"
	})
	handler := NewInspectToolHandler(registry).WithFilter(filter)
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"name": "list_milestones"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !result.IsError || !strings.Contains(resultText(t, result), "not allowed") {
		t.Fatalf("result = %#v, want not-allowed error", result)
	}
}

func TestInspectToolHandlerRejectsNonInvokableTool(t *testing.T) {
	handler := NewInspectToolHandler(testRegistry(t))
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"name": "admin_only_tool"}
	result, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !result.IsError || !strings.Contains(resultText(t, result), "not enabled") {
		t.Fatalf("result = %#v, want not-enabled error", result)
	}
}

func TestUnknownToolSuggestionsRespectRequestFilter(t *testing.T) {
	registry := testRegistry(t)
	filter := ToolFilterFunc(func(_ context.Context, spec catalog.Spec) bool {
		return spec.Tool.Name != "list_milestones"
	})

	t.Run("inspect", func(t *testing.T) {
		handler := NewInspectToolHandler(registry).WithFilter(filter)
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]any{"name": "list_milestone"}
		result, err := handler.Handle(context.Background(), request)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if strings.Contains(resultText(t, result), "list_milestones") {
			t.Fatalf("filtered tool appeared in inspect suggestion: %s", resultText(t, result))
		}
	})

	t.Run("invoke", func(t *testing.T) {
		handler := NewInvokeToolHandler(registry).WithFilter(filter)
		request := mcp.CallToolRequest{}
		request.Params.Arguments = map[string]any{"name": "list_milestone"}
		result, err := handler.Handle(context.Background(), request)
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if strings.Contains(resultText(t, result), "list_milestones") {
			t.Fatalf("filtered tool appeared in invoke suggestion: %s", resultText(t, result))
		}
	})
}

func TestSearchAndInvokeRespectRequestFilter(t *testing.T) {
	registry := testRegistry(t)
	filter := ToolFilterFunc(func(_ context.Context, spec catalog.Spec) bool {
		return spec.Tool.Name != "list_milestones"
	})

	searchHandler := NewSearchToolsHandler(registry, nil).WithFilter(filter)
	searchRequest := mcp.CallToolRequest{}
	searchRequest.Params.Arguments = map[string]any{"query": "里程碑"}
	searchResult, err := searchHandler.Handle(context.Background(), searchRequest)
	if err != nil {
		t.Fatalf("search Handle() error = %v", err)
	}
	if strings.Contains(resultText(t, searchResult), "list_milestones") {
		t.Fatalf("filtered tool appeared in search: %s", resultText(t, searchResult))
	}

	invokeHandler := NewInvokeToolHandler(registry).WithFilter(filter)
	invokeRequest := mcp.CallToolRequest{}
	invokeRequest.Params.Arguments = map[string]any{"name": "list_milestones", "arguments": map[string]any{}}
	invokeResult, err := invokeHandler.Handle(context.Background(), invokeRequest)
	if err != nil {
		t.Fatalf("invoke Handle() error = %v", err)
	}
	if !invokeResult.IsError || !strings.Contains(resultText(t, invokeResult), "not allowed") {
		t.Fatalf("invoke result = %#v, want not-allowed error", invokeResult)
	}
}
