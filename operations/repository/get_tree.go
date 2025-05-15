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
	GetRepoTree = "get_enterprise_repo_tree"
)

var GetRepoTreeTool = mcp.NewTool(GetRepoTree,
	mcp.WithDescription("Get the tree of a repository"),
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
	mcp.WithString(
		"ref",
		mcp.Description("Format: branch + directory (eg: /master or /master/directory)"),
		mcp.Required(),
	),
)

func GetRepoTreeHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_id", "ref"); err != nil {
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
	ref, ok := request.Params.Arguments["ref"].(string)
	if !ok {
		return mcp.NewToolResultError("ref is invalid"), nil
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/tree/%s", enterpriseID, url.QueryEscape(projectIDArg.(string)), url.QueryEscape(ref))

	opts = append(opts, utils.WithQuery(request.Params.Arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.RepoTree{}
	return giteeClient.HandleMCPResult(&data)
}
