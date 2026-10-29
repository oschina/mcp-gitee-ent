package test_modules

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListTestCaseModules   = "list_test_case_modules"
	GetTestCaseModuleTree = "get_test_case_module_tree"
	CreateTestCaseModule  = "create_test_case_module"
	UpdateTestCaseModule  = "update_test_case_module"
	MoveTestCaseModule    = "move_test_case_module"
)

var ListTestCaseModulesTool = mcp.NewTool(ListTestCaseModules,
	mcp.WithDescription("List test case modules under a project. Use get_test_case_module_tree when the full hierarchy is needed."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("parent_id", mcp.Description("Optional parent module ID.")),
	mcp.WithString("search", mcp.Description("Optional module name keyword.")),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetTestCaseModuleTreeTool = mcp.NewTool(GetTestCaseModuleTree,
	mcp.WithDescription("Get the test case module tree for a project. Use this to discover module IDs before creating a test case."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("search", mcp.Description("Optional module name keyword.")),
	mcp.WithNumber("test_plan_id", mcp.Description("Optional test plan ID used to exclude cases already in that plan.")),
	mcp.WithReadOnlyHintAnnotation(true),
)

var CreateTestCaseModuleTool = mcp.NewTool(CreateTestCaseModule,
	mcp.WithDescription("Create a test case module. Use get_test_case_module_tree to find the parent module ID."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithString("name", mcp.Description("Module name."), mcp.Required()),
	mcp.WithNumber("parent_id", mcp.Description("Optional parent module ID.")),
)

var UpdateTestCaseModuleTool = mcp.NewTool(UpdateTestCaseModule,
	mcp.WithDescription("Rename an existing test case module."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("case_module_id", mcp.Description("Module ID."), mcp.Required()),
	mcp.WithString("name", mcp.Description("New module name."), mcp.Required()),
)

var MoveTestCaseModuleTool = mcp.NewTool(MoveTestCaseModule,
	mcp.WithDescription("Move a test case module relative to another module."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("case_module_id", mcp.Description("Module ID to move."), mcp.Required()),
	mcp.WithString("direction", mcp.Description("Move operation."), mcp.Required(), mcp.Enum("up", "down", "inside")),
	mcp.WithNumber("target_id", mcp.Description("Target module ID."), mcp.Required()),
)

func ListTestCaseModulesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{"parent_id", "search"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/case_modules", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetTestCaseModuleTreeHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{"search", "test_plan_id"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/case_modules/as_tree", enterpriseID), opts...)
	var data []any
	result, err := client.HandleMCPResultContext(ctx, &data)
	if result != nil && result.IsError {
		return result, err
	}
	return utils.MCPListResult(data)
}

func CreateTestCaseModuleHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := requiredText(arguments, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID, "name": name}
	if value, ok := arguments["parent_id"]; ok && value != nil {
		payload["parent_id"] = value
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/case_modules", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateTestCaseModuleHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	moduleID, err := requiredInt(arguments, "case_module_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := requiredText(arguments, "name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{"program_id": programID, "name": name}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/case_modules/%d", enterpriseID, moduleID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"case_module_id": moduleID, "name": name})
}

func MoveTestCaseModuleHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	moduleID, err := requiredInt(arguments, "case_module_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	targetID, err := requiredInt(arguments, "target_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	direction, err := requiredText(arguments, "direction")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	switch direction {
	case "up", "down", "inside":
	default:
		return mcp.NewToolResultError("direction must be up, down, or inside"), nil
	}
	payload := map[string]any{
		"program_id": programID,
		"direction":  direction,
		"target_id":  targetID,
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/case_modules/%d/move", enterpriseID, moduleID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"case_module_id": moduleID, "direction": direction, "target_id": targetID})
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

const DeleteTestCaseModule = "delete_test_case_module"

var DeleteTestCaseModuleTool = mcp.NewTool(DeleteTestCaseModule,
	mcp.WithDescription("Delete a test case module. WARNING: this cascades and deletes all descendant modules and their test cases."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID required by test management APIs."), mcp.Required()),
	mcp.WithNumber("case_module_id", mcp.Description("Module ID to delete."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteTestCaseModuleHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	moduleID, err := requiredInt(arguments, "case_module_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"program_id": programID}))
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/case_modules/%d", enterpriseID, moduleID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"case_module_id": moduleID, "deleted": true})
}
