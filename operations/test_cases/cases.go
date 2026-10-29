package test_cases

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListTestCases   = "list_test_cases"
	GetTestCase     = "get_test_case"
	CreateTestCase  = "create_test_case"
	UpdateTestCase  = "update_test_case"
	ArchiveTestCase = "archive_test_case"
	RestoreTestCase = "restore_test_case"
)

var ListTestCasesTool = mcp.NewTool(ListTestCases,
	mcp.WithDescription("List test cases in a project. Use get_test_case_module_tree to discover module IDs."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("module_id", mcp.Description("Optional module ID.")),
	mcp.WithString("case_types", mcp.Description("Comma-separated case type IDs: 1 other, 2 functional, 3 performance, 4 configuration, 5 deployment, 6 API, 7 security, 8 compatibility, 9 UI.")),
	mcp.WithString("maintainer_ids", mcp.Description("Comma-separated maintainer user IDs.")),
	mcp.WithString("priorities", mcp.Description("Comma-separated priorities: 0 P0, 1 P1, 2 P2, 3 P3, 4 P4.")),
	mcp.WithString("states", mcp.Description("Comma-separated states: 0 draft, 1 reviewing, 2 passed, 3 failed, 4 deprecated.")),
	mcp.WithString("creator_ids", mcp.Description("Comma-separated creator user IDs.")),
	mcp.WithString("search", mcp.Description("Search keyword.")),
	mcp.WithBoolean("show_submodule", mcp.Description("Include cases in submodules.")),
	mcp.WithString("sort", mcp.Description("Sort field: created_at, updated_at, state, priority, ident.")),
	mcp.WithString("direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetTestCaseTool = mcp.NewTool(GetTestCase,
	mcp.WithDescription("Get one test case by ID, including steps and expected results."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var CreateTestCaseTool = mcp.NewTool(CreateTestCase,
	mcp.WithDescription("Create a test case. Requires title, module_id, and case_steps; use get_test_case_module_tree to discover module IDs. "+
		"Every step needs a non-empty description; expected_result is optional. Step order follows the array order."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Test case title."), mcp.Required()),
	mcp.WithNumber("module_id", mcp.Description("Module ID."), mcp.Required()),
	mcp.WithString("precondition", mcp.Description("Precondition.")),
	mcp.WithNumber("case_type", mcp.Description("Case type: 1 other, 2 functional, 3 performance, 4 configuration, 5 deployment, 6 API, 7 security, 8 compatibility, 9 UI."), mcp.Min(1), mcp.Max(9)),
	mcp.WithNumber("priority", mcp.Description("Priority: 0 P0, 1 P1, 2 P2, 3 P3, 4 P4."), mcp.Min(0), mcp.Max(4)),
	mcp.WithNumber("maintainer_id", mcp.Description("Maintainer user ID.")),
	mcp.WithString("remark", mcp.Description("Remark.")),
	mcp.WithString("attachment_ids", mcp.Description("Comma-separated attachment IDs.")),
	mcp.WithArray("case_steps",
		mcp.Description("Required case steps, at least one. Each item needs description; expected_result is optional."),
		mcp.Required(),
		mcp.MinItems(1),
		mcp.Items(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description":     map[string]any{"type": "string"},
				"expected_result": map[string]any{"type": "string"},
			},
			"required": []string{"description"},
		}),
	),
)

var UpdateTestCaseTool = mcp.NewTool(UpdateTestCase,
	mcp.WithDescription("Update mutable fields of one test case. Omitted fields remain unchanged; at least one mutable field is required."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithString("title", mcp.Description("Test case title.")),
	mcp.WithString("precondition", mcp.Description("Precondition.")),
	mcp.WithNumber("state", mcp.Description("Case state: 0 draft, 1 reviewing, 2 passed, 3 failed, 4 deprecated."), mcp.Min(0), mcp.Max(4)),
	mcp.WithNumber("module_id", mcp.Description("Module ID.")),
	mcp.WithNumber("case_type", mcp.Description("Case type: 1-9."), mcp.Min(1), mcp.Max(9)),
	mcp.WithNumber("priority", mcp.Description("Priority: 0-4."), mcp.Min(0), mcp.Max(4)),
	mcp.WithNumber("maintainer_id", mcp.Description("Maintainer user ID.")),
	mcp.WithString("remark", mcp.Description("Remark.")),
	mcp.WithArray("case_steps",
		mcp.Description("Case steps. Each item supports description, expected_result, sort, and id."),
		mcp.Items(map[string]any{"type": "object"}),
	),
)

var ArchiveTestCaseTool = mcp.NewTool(ArchiveTestCase,
	mcp.WithDescription("Archive one test case."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
)

var RestoreTestCaseTool = mcp.NewTool(RestoreTestCase,
	mcp.WithDescription("Restore one archived test case."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
)

var mutableCaseFields = []string{
	"title", "precondition", "state", "module_id",
	"case_type", "priority", "maintainer_id", "remark", "case_steps",
}

func ListTestCasesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{
		"module_id", "case_types", "maintainer_ids", "priorities", "states",
		"creator_ids", "search", "show_submodule", "sort", "direction", "page", "per_page",
	} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/test_cases", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetTestCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/test_cases/%d", enterpriseID, testCaseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func CreateTestCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredText(arguments, "title"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if _, err := requiredInt(arguments, "module_id"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	caseSteps, err := utils.NormalizeCaseSteps(arguments["case_steps"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID, "case_steps": caseSteps}
	for _, key := range []string{
		"title", "precondition", "case_type", "priority", "maintainer_id",
		"remark", "attachment_ids", "module_id",
	} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/test_cases", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID}
	for _, key := range mutableCaseFields {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if len(payload) == 1 {
		return mcp.NewToolResultError("At least one mutable field is required: title, precondition, state, module_id, case_type, priority, maintainer_id, remark, case_steps."), nil
	}
	if title, ok := payload["title"]; ok {
		text, ok := title.(string)
		if !ok || text == "" {
			return mcp.NewToolResultError("Parameter title must be a non-empty string"), nil
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/test_cases/%d", enterpriseID, testCaseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ArchiveTestCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changeArchiveState(ctx, request, opts, "archive")
}

func RestoreTestCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changeArchiveState(ctx, request, opts, "restore")
}

func changeArchiveState(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, action string) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient(
		"PUT",
		fmt.Sprintf("/%d/test_cases/%d/%s", enterpriseID, testCaseID, action),
		opts...,
	)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"test_case_id": testCaseID, "archived": action == "archive"})
}

func parseScope(request mcp.CallToolRequest) (map[string]any, int, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, 0, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, 0, err
	}
	programID, err := requiredInt(arguments, "program_id")
	if err != nil {
		return nil, 0, 0, err
	}
	return arguments, enterpriseID, programID, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}

func requiredText(arguments map[string]any, key string) (string, error) {
	value, ok := arguments[key].(string)
	if !ok || value == "" {
		return "", fmt.Errorf("Missing required parameter: %s", key)
	}
	return value, nil
}

const DeleteTestCase = "delete_test_case"

var DeleteTestCaseTool = mcp.NewTool(DeleteTestCase,
	mcp.WithDescription("Delete one test case permanently."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("test_case_id", mcp.Description("Test case ID."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteTestCaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	testCaseID, err := requiredInt(arguments, "test_case_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/test_cases/%d", enterpriseID, testCaseID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"test_case_id": testCaseID, "deleted": true})
}
