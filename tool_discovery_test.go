package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/operations/discovery"
	"gitee.com/oschina/mcp-gitee-ent/operations/documents"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func resetToolsetFlags(t *testing.T) {
	t.Helper()
	previousEnabled := enabledToolsetsFlag
	previousDisabled := disabledToolsetsFlag
	enabledToolsetsFlag = ""
	disabledToolsetsFlag = ""
	t.Setenv("ENABLED_TOOLSETS", "")
	t.Setenv("DISABLED_TOOLSETS", "")
	t.Cleanup(func() {
		enabledToolsetsFlag = previousEnabled
		disabledToolsetsFlag = previousDisabled
	})
}

func initializeInProcessClient(t *testing.T) *client.Client {
	t.Helper()
	mcpServer := newMCPServer()
	addTools(mcpServer)
	inProcessClient, err := client.NewInProcessClient(mcpServer)
	if err != nil {
		t.Fatalf("NewInProcessClient() error = %v", err)
	}
	t.Cleanup(func() { _ = inProcessClient.Close() })
	if err := inProcessClient.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	initializeRequest := mcp.InitializeRequest{}
	initializeRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initializeRequest.Params.ClientInfo = mcp.Implementation{Name: "discovery-test-client", Version: "1.0.0"}
	if _, err := inProcessClient.Initialize(context.Background(), initializeRequest); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	return inProcessClient
}

func callTool(t *testing.T, mcpClient *client.Client, name string, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = arguments
	result, err := mcpClient.CallTool(context.Background(), request)
	if err != nil {
		t.Fatalf("CallTool(%s) error = %v", name, err)
	}
	return result
}

func textContent(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil || len(result.Content) == 0 {
		t.Fatalf("result = %#v, want content", result)
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want mcp.TextContent", result.Content[0])
	}
	return text.Text
}

func TestDiscoveryToolsRegisteredAndSearchable(t *testing.T) {
	resetToolsetFlags(t)
	mcpClient := initializeInProcessClient(t)

	toolsResult, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(toolsResult.Tools) != 43 {
		t.Fatalf("tool count = %d, want 40 existing + 3 discovery tools", len(toolsResult.Tools))
	}

	names := make(map[string]bool)
	for _, tool := range toolsResult.Tools {
		names[tool.Name] = true
	}
	for _, expected := range []string{discovery.SearchTools, discovery.InspectTool, discovery.InvokeTool} {
		if !names[expected] {
			t.Errorf("discovery tool %s was not registered", expected)
		}
	}

	searchResult := callTool(t, mcpClient, discovery.SearchTools, map[string]any{"query": "文档"})
	searchText := textContent(t, searchResult)
	if searchResult.IsError || !strings.Contains(searchText, documents.ListDocuments) {
		t.Fatalf("search result = %q, want %s", searchText, documents.ListDocuments)
	}

	inspectResult := callTool(t, mcpClient, discovery.InspectTool, map[string]any{"name": documents.ListDocuments})
	inspectText := textContent(t, inspectResult)
	if inspectResult.IsError {
		t.Fatalf("inspect result is error: %s", inspectText)
	}
	response := map[string]any{}
	if err := json.Unmarshal([]byte(inspectText), &response); err != nil {
		t.Fatalf("inspect result is not JSON: %v", err)
	}
	if response["name"] != documents.ListDocuments {
		t.Fatalf("inspect name = %v", response["name"])
	}
}

func TestInvokeToolCallsExistingHandlerThroughMCP(t *testing.T) {
	resetToolsetFlags(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/doc_nodes/tile" {
			t.Errorf("request = %s %s, want GET /7/doc_nodes/tile", request.Method, request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":91,"name":"Handbook","file_type":"WikiInfo","file_id":42,"editor_type":"markdown"}]}`)
	}))
	defer apiServer.Close()

	previousAPIBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(apiServer.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() {
		utils.SetApiBase(previousAPIBase)
		utils.SetGiteeAccessToken(previousToken)
	})

	mcpClient := initializeInProcessClient(t)
	result := callTool(t, mcpClient, discovery.InvokeTool, map[string]any{
		"name":      documents.ListDocuments,
		"arguments": map[string]any{"enterprise_id": 7},
	})
	if result.IsError {
		t.Fatalf("invoke result is error: %s", textContent(t, result))
	}
	if !strings.Contains(textContent(t, result), "Handbook") {
		t.Fatalf("invoke result = %q, want document payload", textContent(t, result))
	}
}

func TestDiscoveryToolCannotBypassDisabledToolsets(t *testing.T) {
	resetToolsetFlags(t)
	t.Setenv("DISABLED_TOOLSETS", documents.ListDocuments)

	mcpClient := initializeInProcessClient(t)
	result := callTool(t, mcpClient, discovery.InvokeTool, map[string]any{
		"name":      documents.ListDocuments,
		"arguments": map[string]any{"enterprise_id": 7},
	})
	if !result.IsError || !strings.Contains(textContent(t, result), "not enabled") {
		t.Fatalf("invoke result = %#v, want not-enabled error", result)
	}

	toolsResult, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	for _, tool := range toolsResult.Tools {
		if tool.Name == documents.ListDocuments {
			t.Fatal("disabled tool is still listed in tools/list")
		}
	}
}

func TestPilotToolsAreDiscoverableButNotDirectlyListed(t *testing.T) {
	resetToolsetFlags(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/7/milestones" {
			t.Errorf("request = %s %s, want GET /7/milestones", request.Method, request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":9,"title":"Q4"}]}`)
	}))
	defer apiServer.Close()

	previousAPIBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(apiServer.URL)
	utils.SetGiteeAccessToken("test-token")
	t.Cleanup(func() {
		utils.SetApiBase(previousAPIBase)
		utils.SetGiteeAccessToken(previousToken)
	})

	mcpClient := initializeInProcessClient(t)
	toolsResult, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	for _, tool := range toolsResult.Tools {
		if tool.Name == "list_milestones" {
			t.Fatal("progressive-only tool must not be listed directly")
		}
	}

	searchResult := callTool(t, mcpClient, discovery.SearchTools, map[string]any{"query": "里程碑"})
	if searchResult.IsError || !strings.Contains(textContent(t, searchResult), "list_milestones") {
		t.Fatalf("search result = %q, want list_milestones", textContent(t, searchResult))
	}

	invokeResult := callTool(t, mcpClient, discovery.InvokeTool, map[string]any{
		"name":      "list_milestones",
		"arguments": map[string]any{"enterprise_id": 7, "search": "Q4"},
	})
	if invokeResult.IsError || !strings.Contains(textContent(t, invokeResult), "Q4") {
		t.Fatalf("invoke result = %q", textContent(t, invokeResult))
	}
}

func TestReadPilotToolsAreDiscoverableByBusinessLanguage(t *testing.T) {
	resetToolsetFlags(t)
	mcpClient := initializeInProcessClient(t)

	cases := map[string]string{
		"工时":   "list_workloads",
		"成员":   "search_members",
		"操作日志": "list_issue_operation_logs",
		"状态流转": "list_issue_next_states",
	}
	for query, expected := range cases {
		result := callTool(t, mcpClient, discovery.SearchTools, map[string]any{"query": query})
		if result.IsError {
			t.Fatalf("search %q returned error: %s", query, textContent(t, result))
		}
		if !strings.Contains(textContent(t, result), expected) {
			t.Errorf("search %q = %q, want %s", query, textContent(t, result), expected)
		}
	}
}

func TestCatalogDomainsAreExposedInSearchDomainEnum(t *testing.T) {
	resetToolsetFlags(t)
	registry, err := buildToolCatalog()
	if err != nil {
		t.Fatalf("buildToolCatalog() error = %v", err)
	}
	allowed := make(map[string]bool, len(discovery.SearchDomains))
	for _, domain := range discovery.SearchDomains {
		allowed[domain] = true
	}
	for _, spec := range registry.List() {
		if !allowed[spec.Domain] {
			t.Errorf("tool %s uses domain %q which is missing from search_tools enum", spec.Tool.Name, spec.Domain)
		}
	}

	domainSchema, ok := discovery.SearchToolsTool.InputSchema.Properties["domain"].(map[string]any)
	if !ok {
		t.Fatal("search_tools domain schema missing")
	}
	enumValues, ok := domainSchema["enum"].([]string)
	if !ok || len(enumValues) != len(discovery.SearchDomains) {
		t.Fatalf("domain enum = %#v, want %d values", domainSchema["enum"], len(discovery.SearchDomains))
	}
}
