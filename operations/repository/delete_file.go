package repository

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DeleteRepositoryFile = "delete_repository_file"
)

var DeleteRepositoryFileTool = mcp.NewTool(
	DeleteRepositoryFile,
	mcp.WithDescription("Delete a file from a repository and commit the removal. "+
		"ref must be the full path including the branch, e.g. \"master/docs/a.md\"."),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"repository_id",
		mcp.Description("Repository ID or PathWithNamespace"),
		mcp.Required(),
	),
	mcp.WithString(
		"ref",
		mcp.Description("Format: branch/directory/file name (eg: master/docs/a.md)"),
		mcp.Required(),
	),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteRepositoryFileHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "repository_id", "ref"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectID, err := utils.SafelyConvertToString(arguments["repository_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	ref, err := utils.SafelyConvertToString(arguments["ref"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	query := map[string]any{"ref": ref}
	if !utils.IsAllDigits(projectID) {
		query["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/blobs/file", enterpriseID, url.QueryEscape(projectID))
	opts = append(opts, utils.WithQuery(query))
	giteeClient := utils.NewGiteeClient("DELETE", apiUrl, opts...)

	if result, err := giteeClient.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"repository_id": projectID, "ref": ref, "deleted": true})
}
