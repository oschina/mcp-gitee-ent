package programs

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListPrograms = "list_programs"
)

var ListProgramsTool = mcp.NewTool(ListPrograms,
	mcp.WithDescription("List programs of an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"type",
		mcp.Description("Filter type (joined: joined programs, assigned: assigned programs, created: created programs, only_star: starred programs)"),
		mcp.Enum("joined", "assigned", "created", "only_star"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at: creation time, updated_at: update time, users_count: member count, projects_count: repository count, issues_count: task count, accessed_at: access time, name: program name)"),
		mcp.Enum("created_at", "updated_at", "users_count", "projects_count", "issues_count", "accessed_at", "name"),
	),
	mcp.WithBoolean(
		"priority_topped",
		mcp.Description("Whether to sort by priority first"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending, desc: descending)"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithString(
		"status",
		mcp.Description("Program status (0: start, 1: pause, 2: close), comma separated, e.g. 0,1"),
	),
	mcp.WithString(
		"category",
		mcp.Description("Program category (all: all, common: common, kanban: kanban), comma separated, e.g. common,kanban"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search parameter"),
	),
	mcp.WithString(
		"assignee_ids",
		mcp.Description("Program assignee IDs, comma separated"),
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

func ListProgramsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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

	apiUrl := fmt.Sprintf("/%d/programs", enterpriseID)
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.Program]{}
	return giteeClient.HandleMCPResult(&data)
}
