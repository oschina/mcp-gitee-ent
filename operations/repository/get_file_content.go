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
	GetFileContent = "get_enterprise_repository_file_content"
)

var GetFileContentTool = mcp.NewTool(
	GetFileContent,
	mcp.WithDescription("Get the content of the specified file in the repository"),
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
		mcp.Description("Format: branch + file path (eg: /master/readme.md)"),
		mcp.Required(),
	),
)

func GetFileContentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	projectID, err := utils.SafelyConvertToString(projectIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	if !utils.IsAllDigits(projectID) {
		request.Params.Arguments["qt"] = "path"
	}
	ref, ok := request.Params.Arguments["ref"].(string)
	if !ok {
		return mcp.NewToolResultError("ref is invalid"), nil
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/blob/%s", enterpriseID, url.QueryEscape(projectID), url.QueryEscape(ref))

	opts = append(opts, utils.WithQuery(request.Params.Arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.FileContent{}
	return giteeClient.HandleMCPResult(&data)
}
