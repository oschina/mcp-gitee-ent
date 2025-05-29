package issues

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListIssueComments = "list_enterprise_issue_comments"
)

var ListIssueCommentsTool = mcp.NewTool(ListIssueComments,
	mcp.WithDescription("List comments of an issue"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"issue_id",
		mcp.Description("Issue Ident"),
		mcp.Required(),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (name or created_at)"),
		mcp.Enum("name", "created_at"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc or desc)"),
		mcp.Enum("asc", "desc"),
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

func ListIssueCommentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate required parameters
	if checkResult, err := utils.CheckRequired(arguments, "issue_id"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	issueIDArg := arguments["issue_id"]
	arguments["qt"] = "ident"

	apiUrl := fmt.Sprintf("/%d/issues/%s/notes", enterpriseID, issueIDArg)
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.IssueComment]{}
	return giteeClient.HandleMCPResult(&data)
}
