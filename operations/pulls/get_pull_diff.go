package pulls

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"net/url"
)

const (
	GetEntPullDiff = "get_enterprise_pull_diff"
)

var GetPullDiffTool = mcp.NewTool(GetEntPullDiff,
	mcp.WithDescription("Get the diff of a pull request"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_id",
		mcp.Description("Project PathWithNamespace"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"pull_request_id",
		mcp.Description("Pull request IID"),
		mcp.Required(),
	),
	mcp.WithString(
		"qt",
		mcp.Description("Query type (path/id)"),
		mcp.Enum("path", "id"),
		mcp.DefaultString("path"),
	),
)

func GetPullDiffHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	pullRequestID, err := utils.SafelyConvertToInt(pullRequestIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	request.Params.Arguments["qt"] = "path"
	request.Params.Arguments["pr_qt"] = "iid"

	apiUrl := fmt.Sprintf("/%d/projects/%s/pull_requests/%d/diff", enterpriseID, url.QueryEscape(projectIDArg.(string)), pullRequestID)
	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	data := types.PullDiff{}
	return giteeClient.HandleMCPResult(&data)
}
