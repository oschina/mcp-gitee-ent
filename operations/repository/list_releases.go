package repository

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"net/url"
)

const (
	ListReleases = "list_enterprise_repo_releases"
)

var ListReleasesTool = mcp.NewTool(ListReleases,
	mcp.WithDescription("List releases for repository in an enterprise"),
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

func ListReleasesHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_id"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectIDArg := request.Params.Arguments["project_id"]
	if !utils.IsAllDigits(projectIDArg.(string)) {
		request.Params.Arguments["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/releases", enterpriseID, url.QueryEscape(projectIDArg.(string)))

	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	data := types.PagedResponse[types.ReleaseDetail]{}
	return giteeClient.HandleMCPResult(&data)
}
