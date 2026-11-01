package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/operations/documents"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestInfrastructureToolAllowedPrecedence(t *testing.T) {
	testCases := []struct {
		name     string
		enabled  string
		disabled string
		want     bool
	}{
		{name: "explicit enabled wins over disabled", enabled: "search_tools", disabled: "search_tools", want: true},
		{name: "allowlist alone does not hide meta tools", enabled: "list_enterprise_issues", want: true},
		{name: "disabled hides meta tools", disabled: "search_tools", want: false},
		{name: "default allows meta tools", want: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resetToolsetFlags(t)
			enabledToolsetsFlag = testCase.enabled
			disabledToolsetsFlag = testCase.disabled
			if got := infrastructureToolAllowed("search_tools"); got != testCase.want {
				t.Fatalf("infrastructureToolAllowed(search_tools) = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestDocumentToolsThroughMCPServer(t *testing.T) {
	previousEnabledToolsetsFlag := enabledToolsetsFlag
	previousDisabledToolsetsFlag := disabledToolsetsFlag
	enabledToolsetsFlag = ""
	disabledToolsetsFlag = ""
	t.Setenv("ENABLED_TOOLSETS", "")
	t.Setenv("DISABLED_TOOLSETS", "")
	defer func() {
		enabledToolsetsFlag = previousEnabledToolsetsFlag
		disabledToolsetsFlag = previousDisabledToolsetsFlag
	}()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/7/doc_nodes/tile" {
			t.Fatalf("request = %s %s, want GET /7/doc_nodes/tile", req.Method, req.URL.Path)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization = %q, want Bearer test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"total_count":1,"data":[{"id":91,"name":"Handbook","file_type":"WikiInfo","file_id":42,"editor_type":"markdown"}]}`)
	}))
	defer apiServer.Close()

	previousAPIBase := utils.GetApiBase()
	previousToken := utils.GetGiteeAccessToken()
	utils.SetApiBase(apiServer.URL)
	utils.SetGiteeAccessToken("test-token")
	defer func() {
		utils.SetApiBase(previousAPIBase)
		utils.SetGiteeAccessToken(previousToken)
	}()

	mcpServer := newMCPServer()
	addTools(mcpServer)
	mcpClient, err := client.NewInProcessClient(mcpServer)
	if err != nil {
		t.Fatalf("NewInProcessClient() error = %v", err)
	}
	defer mcpClient.Close()
	if err := mcpClient.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	initializeRequest := mcp.InitializeRequest{}
	initializeRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initializeRequest.Params.ClientInfo = mcp.Implementation{Name: "document-test-client", Version: "1.0.0"}
	initializeResult, err := mcpClient.Initialize(context.Background(), initializeRequest)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if initializeResult.ServerInfo.Name != "mcp-gitee-ent" {
		t.Fatalf("server name = %q, want mcp-gitee-ent", initializeResult.ServerInfo.Name)
	}

	toolsResult, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	documentToolNames := map[string]bool{
		documents.ListDocuments:      false,
		documents.ListDocumentPages:  false,
		documents.GetDocumentPage:    false,
		documents.UploadDocument:     false,
		documents.DownloadDocument:   false,
		documents.UpdateDocumentPage: false,
	}
	for _, tool := range toolsResult.Tools {
		if _, ok := documentToolNames[tool.Name]; !ok {
			continue
		}
		documentToolNames[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %s has no description", tool.Name)
		}
		if _, ok := tool.InputSchema.Properties["enterprise_id"]; !ok {
			t.Errorf("tool %s has no enterprise_id property", tool.Name)
		}
	}
	for name, found := range documentToolNames {
		if !found {
			t.Errorf("document tool %s was not registered", name)
		}
	}

	callRequest := mcp.CallToolRequest{}
	callRequest.Params.Name = documents.ListDocuments
	callRequest.Params.Arguments = map[string]interface{}{"enterprise_id": 7}
	callResult, err := mcpClient.CallTool(context.Background(), callRequest)
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if callResult.IsError || len(callResult.Content) != 1 {
		t.Fatalf("CallTool() result = %#v", callResult)
	}
	content, ok := callResult.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("CallTool() content type = %T, want mcp.TextContent", callResult.Content[0])
	}
	response := types.DocumentListResponse{}
	if err := json.Unmarshal([]byte(content.Text), &response); err != nil {
		t.Fatalf("decode CallTool() content: %v", err)
	}
	if len(response.Documents) != 1 || response.Documents[0].DocumentID != 42 || response.Documents[0].DocumentNodeID != 91 {
		t.Fatalf("CallTool() response = %#v", response)
	}
}
