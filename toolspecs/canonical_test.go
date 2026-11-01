package toolspecs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

func buildCanonicalRegistry(t *testing.T) *catalog.Registry {
	t.Helper()
	registry, err := Build(func(string) bool { return true })
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	return registry
}

func schemaProperties(t *testing.T, registry *catalog.Registry, tool string) map[string]any {
	t.Helper()
	spec, ok := registry.Get(tool)
	if !ok {
		t.Fatalf("tool %q is missing", tool)
	}
	return spec.Tool.InputSchema.Properties
}

func TestCanonicalRepositoryParameters(t *testing.T) {
	registry := buildCanonicalRegistry(t)
	for _, tool := range []string{
		"list_project_check_runs",
		"get_check_run_annotations",
		"comment_enterprise_pull",
		"create_enterprise_repo_release",
		"list_enterprise_issues",
	} {
		properties := schemaProperties(t, registry, tool)
		if _, ok := properties["repository_id"]; !ok {
			t.Errorf("%s is missing repository_id", tool)
		}
		if _, ok := properties["project_id"]; ok {
			t.Errorf("%s still exposes project_id", tool)
		}
	}
	properties := schemaProperties(t, registry, "search_members")
	if _, ok := properties["repository_ids"]; !ok {
		t.Error("search_members is missing repository_ids")
	}
	if _, ok := properties["project_ids"]; ok {
		t.Error("search_members still exposes project_ids")
	}
}

func TestCanonicalFilterAndTestManagementParameters(t *testing.T) {
	registry := buildCanonicalRegistry(t)
	tests := []struct {
		tool     string
		present  string
		absent   string
		wantType string
	}{
		{tool: "list_enterprise_issues", present: "program_ids", absent: "program_id"},
		{tool: "list_enterprise_issues", present: "priorities", absent: "priority"},
		{tool: "list_issue_related_issues", present: "priorities", absent: "priority"},
		{tool: "update_test_report", present: "passed", absent: "state"},
		{tool: "update_test_plan_state", present: "state_id", absent: "state"},
		{tool: "list_programs", present: "statuses", absent: "status"},
		{tool: "list_enterprise_pulls", present: "assignee_ids", absent: "assignee_id"},
		{tool: "list_test_plan_cases", present: "linked_issue_id", absent: "issue_id"},
		{tool: "list_test_reports", present: "test_plan_ids", wantType: "array"},
	}
	for _, tt := range tests {
		properties := schemaProperties(t, registry, tt.tool)
		property, ok := properties[tt.present].(map[string]any)
		if !ok {
			t.Errorf("%s is missing %s", tt.tool, tt.present)
			continue
		}
		if _, ok := properties[tt.absent]; ok {
			t.Errorf("%s still exposes %s", tt.tool, tt.absent)
		}
		if tt.wantType != "" && property["type"] != tt.wantType {
			t.Errorf("%s.%s type = %v, want %s", tt.tool, tt.present, property["type"], tt.wantType)
		}
	}
}

func TestCanonicalRequiredAndExamplesUseCanonicalNames(t *testing.T) {
	registry := buildCanonicalRegistry(t)
	spec, ok := registry.Get("list_project_check_runs")
	if !ok {
		t.Fatal("list_project_check_runs is missing")
	}
	found := false
	for _, name := range spec.Tool.InputSchema.Required {
		if name == "repository_id" {
			found = true
		}
		if name == "project_id" {
			t.Errorf("required still contains project_id")
		}
	}
	if !found {
		t.Error("required does not contain repository_id")
	}

	for _, spec := range registry.List() {
		properties := spec.Tool.InputSchema.Properties
		for key := range spec.ExampleArguments {
			if _, ok := properties[key]; !ok {
				t.Errorf("%s example contains unknown parameter %q", spec.Tool.Name, key)
			}
		}
	}
}

func TestCanonicalProgramDescriptions(t *testing.T) {
	registry := buildCanonicalRegistry(t)
	for _, spec := range registry.List() {
		for name, raw := range spec.Tool.InputSchema.Properties {
			if name != "program_id" && name != "program_ids" {
				continue
			}
			property, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			description, _ := property["description"].(string)
			if !strings.Contains(description, "Gitee 项目") {
				t.Errorf("%s.%s description = %q, want Gitee 项目 clarification", spec.Tool.Name, name, description)
			}
			if strings.Contains(description, "Project ID") {
				t.Errorf("%s.%s description still says Project ID: %q", spec.Tool.Name, name, description)
			}
		}
	}
}

func TestCanonicalHandlerTranslatesAndRejectsLegacy(t *testing.T) {
	var received map[string]any
	handler := canonicalHandler(func(_ context.Context, request mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
		received, _ = request.Params.Arguments.(map[string]any)
		return mcp.NewToolResultText("ok"), nil
	}, []paramRename{{Legacy: "project_id", Canonical: "repository_id"}}, []string{"repository_id"})

	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"repository_id": 9}
	result, err := handler(context.Background(), request)
	if err != nil || result.IsError {
		t.Fatalf("canonical call failed: %v %#v", err, result)
	}
	if received["project_id"] != 9 {
		t.Errorf("handler received project_id = %v, want 9", received["project_id"])
	}
	if _, ok := received["repository_id"]; ok {
		t.Error("canonical alias leaked into handler arguments")
	}

	legacyRequest := mcp.CallToolRequest{}
	legacyRequest.Params.Arguments = map[string]any{"project_id": 9}
	result, err = handler(context.Background(), legacyRequest)
	if err != nil {
		t.Fatalf("legacy rejection returned error: %v", err)
	}
	if !result.IsError || !strings.Contains(result.Content[0].(mcp.TextContent).Text, "renamed to") {
		t.Errorf("legacy parameter was not rejected: %#v", result)
	}
}

func TestCanonicalHandlerValidatesRequiredCanonicalName(t *testing.T) {
	handler := canonicalHandler(func(context.Context, mcp.CallToolRequest, ...utils.Option) (*mcp.CallToolResult, error) {
		t.Fatal("handler must not run when a canonical required parameter is missing")
		return nil, nil
	}, []paramRename{{Legacy: "project_id", Canonical: "repository_id"}}, []string{"repository_id"})
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{}
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("required validation returned error: %v", err)
	}
	if !result.IsError || !strings.Contains(result.Content[0].(mcp.TextContent).Text, "repository_id") {
		t.Errorf("required error = %#v, want canonical name", result)
	}
}

func TestCanonicalTestPlanIDsArrayEncodesLegacyString(t *testing.T) {
	var received map[string]any
	handler := canonicalHandler(func(_ context.Context, request mcp.CallToolRequest, _ ...utils.Option) (*mcp.CallToolResult, error) {
		received, _ = request.Params.Arguments.(map[string]any)
		return mcp.NewToolResultText("ok"), nil
	}, canonicalParamRules["list_test_reports"], nil)
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"test_plan_ids": []any{1, 2}}
	result, err := handler(context.Background(), request)
	if err != nil || result.IsError {
		t.Fatalf("array call failed: %v %#v", err, result)
	}
	if received["test_plan_ids"] != "1,2" {
		t.Errorf("legacy test_plan_ids = %#v, want \"1,2\"", received["test_plan_ids"])
	}
}

type canonicalNoopAuthorizer struct{}

func (canonicalNoopAuthorizer) Authorize(*http.Request) error { return nil }

func TestCanonicalRepositoryParameterReachesLegacyAPI(t *testing.T) {
	var path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = io.WriteString(w, `{"data":[]}`)
	}))
	defer server.Close()

	registry := buildCanonicalRegistry(t)
	spec, ok := registry.Get("list_project_check_runs")
	if !ok {
		t.Fatal("list_project_check_runs is missing")
	}
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{"enterprise_id": 7, "repository_id": 9}
	result, err := spec.Handler(context.Background(), request,
		utils.WithApiBase(server.URL), utils.WithAuthorizer(canonicalNoopAuthorizer{}))
	if err != nil || result.IsError {
		t.Fatalf("canonical handler failed: %v %#v", err, result)
	}
	if path != "/7/projects/9/check_runs/list" {
		t.Errorf("upstream path = %q, want repository path", path)
	}
}

func TestCanonicalTestPlanIDsReachLegacyQuery(t *testing.T) {
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query().Get("test_plan_ids")
		_, _ = io.WriteString(w, `{"total_count":0,"data":[]}`)
	}))
	defer server.Close()

	registry := buildCanonicalRegistry(t)
	spec, ok := registry.Get("list_test_reports")
	if !ok {
		t.Fatal("list_test_reports is missing")
	}
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]any{
		"enterprise_id": 7, "program_id": 9, "test_plan_ids": []any{1, 2},
	}
	result, err := spec.Handler(context.Background(), request,
		utils.WithApiBase(server.URL), utils.WithAuthorizer(canonicalNoopAuthorizer{}))
	if err != nil || result.IsError {
		t.Fatalf("canonical handler failed: %v %#v", err, result)
	}
	if query != "1,2" {
		t.Errorf("upstream test_plan_ids = %q, want 1,2", query)
	}
}
