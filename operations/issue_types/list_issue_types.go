package issue_types

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListIssueTypes = "list_issue_types"
)

var ListIssueTypesTool = mcp.NewTool(ListIssueTypes,
	mcp.WithDescription("List issue types of an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search field"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at, updated_at, serial)"),
		mcp.Enum("created_at", "updated_at", "serial"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending desc: descending)"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithNumber(
		"program_id",
		mcp.Description("Program ID"),
	),
	mcp.WithString(
		"scope",
		mcp.Description("Query scope (all, customize)"),
		mcp.Enum("all", "customize"),
	),
	mcp.WithString(
		"category",
		mcp.Description("Category of issue type"),
		mcp.Enum("task", "bug", "requirement", "feature"),
	),
	mcp.WithNumber(
		"page",
		mcp.Description("Current page number"),
		mcp.DefaultNumber(1),
	),
	mcp.WithNumber(
		"per_page",
		mcp.Description("Number of items per page, default 20"),
		mcp.DefaultNumber(20),
	),
)

func ListIssueTypesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate required parameters
	if checkResult, err := utils.CheckRequired(arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/issue_types/enterprise_issue_types", enterpriseID)
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.IssueType]{}
	return giteeClient.HandleMCPResult(&data)
}
