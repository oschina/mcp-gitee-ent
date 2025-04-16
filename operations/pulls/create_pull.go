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
	CreateEntPull = "create_enterprise_repo_pull"
)

var CreateEntPullTool = mcp.NewTool(CreateEntPull,
	mcp.WithDescription("Create a pull request for repository in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_id",
		mcp.Description("Repository ID or repository's PathWithNamespace"),
		mcp.Required(),
	),
	mcp.WithString(
		"source_repo",
		mcp.Description("Source repository ID"),
	),
	mcp.WithString(
		"source_branch",
		mcp.Description("Source branch name"),
		mcp.Required(),
	),
	mcp.WithString(
		"target_branch",
		mcp.Description("Target branch name"),
		mcp.Required(),
	),
	mcp.WithString(
		"title",
		mcp.Description("Title"),
		mcp.Required(),
	),
	mcp.WithString(
		"body",
		mcp.Description("Content"),
	),
	mcp.WithString(
		"assignee_id",
		mcp.Description("Assignee IDs (comma separated)"),
	),
	mcp.WithString(
		"tester_id",
		mcp.Description("Tester IDs (comma separated)"),
	),
	mcp.WithString(
		"draft",
		mcp.Description("Whether to specify as draft (true/false)"),
	),
)

func CreateEntPullHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_id", "source_branch", "target_branch", "title"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectIDArg := request.Params.Arguments["project_id"]

	apiUrl := fmt.Sprintf("/%d/projects/%s/pull_requests", enterpriseID, url.QueryEscape(projectIDArg.(string)))
	if !utils.IsAllDigits(projectIDArg.(string)) {
		request.Params.Arguments["qt"] = "path"
	}
	giteeClient := utils.NewGiteeClient("POST", apiUrl, utils.WithPayload(request.Params.Arguments))

	data := types.PullDetail{}
	return giteeClient.HandleMCPResult(&data)
}
