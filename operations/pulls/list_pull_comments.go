package pulls

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEntPullComments = "list_enterprise_pull_comments"
)

var ListPullCommentsTool = mcp.NewTool(ListEntPullComments,
	mcp.WithDescription("List comments of a pull request"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_id",
		mcp.Description("Project ID or PathWithNamespace"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"pull_request_id",
		mcp.Description("Pull request ID"),
		mcp.Required(),
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

func ListPullCommentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate required parameters
	if checkResult, err := utils.CheckRequired(arguments, "project_id", "pull_request_id"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectIDArg := arguments["project_id"]
	projectID, err := utils.SafelyConvertToString(projectIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	pullRequestIDArg := arguments["pull_request_id"]
	pullRequestID, err := utils.SafelyConvertToInt(pullRequestIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if !utils.IsAllDigits(projectID) {
		arguments["qt"] = "path"
	}
	arguments["pr_qt"] = "iid"

	apiUrl := fmt.Sprintf("/%d/projects/%s/pull_requests/%d/comments", enterpriseID, url.QueryEscape(projectID), pullRequestID)
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.PullComment]{}
	return giteeClient.HandleMCPResult(&data)
}
