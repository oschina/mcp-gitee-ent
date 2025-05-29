package repository

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
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
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate required parameters
	if checkResult, err := utils.CheckRequired(arguments, "project_id", "ref"); err != nil {
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

	if !utils.IsAllDigits(projectID) {
		arguments["qt"] = "path"
	}
	refArg := arguments["ref"]

	apiUrl := fmt.Sprintf("/%d/projects/%s/repository/tree", enterpriseID, url.QueryEscape(projectID))
	arguments["ref"] = refArg
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.RepoTree{}
	return giteeClient.HandleMCPResult(&data)
}
