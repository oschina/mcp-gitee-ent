package pulls

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"net/url"
)

const (
	MergeEntPull = "merge_enterprise_pull"
)

var MergePullTool = mcp.NewTool(MergeEntPull,
	mcp.WithDescription("Merge a pull request in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_id",
		mcp.Description("Project ID or path"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"pull_request_id",
		mcp.Description("Pull request ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"merge_method",
		mcp.Description("Merge method"),
		mcp.Enum("merge", "squash", "rebase"),
		mcp.DefaultString("merge"),
	),
	mcp.WithString(
		"title",
		mcp.Description("Merge title"),
	),
	mcp.WithString(
		"description",
		mcp.Description("Merge description"),
	),
)

func MergePullHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_id", "pull_request_id"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectIDArg := request.Params.Arguments["project_id"]
	pullRequestIDArg := request.Params.Arguments["pull_request_id"]
	pullRequestId, err := utils.SafelyConvertToInt(pullRequestIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/pull_requests/%d/merge", enterpriseID, url.QueryEscape(projectIDArg.(string)), pullRequestId)
	request.Params.Arguments["pr_qt"] = "iid"
	giteeClient := utils.NewGiteeClient("POST", apiUrl, utils.WithPayload(request.Params.Arguments))
	return giteeClient.HandleMCPResult(nil)
}
