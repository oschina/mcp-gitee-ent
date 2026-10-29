package kanbans

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListKanbans       = "list_kanbans"
	ListKanbanColumns = "list_kanban_columns"
)

var ListKanbansTool = mcp.NewTool(ListKanbans,
	mcp.WithDescription("List kanban boards in an enterprise, optionally scoped to a project."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Optional project ID.")),
	mcp.WithString("search", mcp.Description("Board title keyword.")),
	mcp.WithBoolean("include_columns", mcp.Description("Include column details in each board.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListKanbanColumnsTool = mcp.NewTool(ListKanbanColumns,
	mcp.WithDescription("List columns of one kanban board."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithNumber("kanban_id", mcp.Description("Kanban board ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListKanbansHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"program_id", "search", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	if value, ok := arguments["include_columns"].(bool); ok && value {
		query["with_columns"] = 1
	}
	if _, ok := query["page"]; ok {
		query["use_page"] = 1
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/kanbans", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListKanbanColumnsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	programID, err := requiredInt(arguments, "program_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	kanbanID, err := requiredInt(arguments, "kanban_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/programs/%d/kanbans/%d/columns", enterpriseID, programID, kanbanID),
		opts...,
	)
	var data []any
	result, err := client.HandleMCPResultContext(ctx, &data)
	if result != nil && result.IsError {
		return result, err
	}
	return utils.MCPListResult(data)
}

func parseEnterprise(request mcp.CallToolRequest) (map[string]any, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, err
	}
	return arguments, enterpriseID, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
