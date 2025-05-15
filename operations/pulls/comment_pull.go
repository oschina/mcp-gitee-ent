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
	CommentEntPull = "comment_enterprise_pull"
)

var CommentPullTool = mcp.NewTool(CommentEntPull,
	mcp.WithDescription("Comment on a pull request in an enterprise"),
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
		mcp.Description("Pull request IID"),
		mcp.Required(),
	),
	mcp.WithString(
		"body",
		mcp.Description("Comment content"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"reply_id",
		mcp.Description("Reply ID"),
	),
)

func CommentPullHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_id", "pull_request_id", "body"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectIDArg := request.Params.Arguments["project_id"]
	projectID, err := utils.SafelyConvertToString(projectIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	pullRequestIDArg := request.Params.Arguments["pull_request_id"]
	pullRequestID, err := utils.SafelyConvertToInt(pullRequestIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/pull_requests/%d/notes", enterpriseID, url.QueryEscape(projectID), pullRequestID)
	request.Params.Arguments["pr_qt"] = "iid"
	if !utils.IsAllDigits(projectID) {
		request.Params.Arguments["qt"] = "path"
	}
	opts = append(opts, utils.WithPayload(request.Params.Arguments))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.PullComment{}
	return giteeClient.HandleMCPResult(&data)
}
