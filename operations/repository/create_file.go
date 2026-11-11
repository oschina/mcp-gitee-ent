package repository

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateRepositoryFile = "create_repository_file"
)

var CreateRepositoryFileTool = mcp.NewTool(
	CreateRepositoryFile,
	mcp.WithDescription("Create a file (or folder) on a branch and commit it. ref is the branch name; "+
		"file_path is the path inside the repository; commit_title is the commit message. "+
		"Set is_new_folder=1 to create a folder instead of a file."),
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
		mcp.Description("Branch name (eg: master)"),
		mcp.Required(),
	),
	mcp.WithString(
		"file_path",
		mcp.Description("New file path"),
	),
	mcp.WithString(
		"content",
		mcp.Description("File content"),
	),
	mcp.WithString(
		"commit_title",
		mcp.Description("Commit title"),
		mcp.Required(),
	),
	mcp.WithString(
		"commit_body",
		mcp.Description("Commit body"),
	),
	mcp.WithNumber(
		"is_new_folder",
		mcp.Description("Set to 1 to create the parent folder when it does not exist"),
	),
	mcp.WithBoolean(
		"overlay",
		mcp.Description("Overwrite the file when it already exists"),
	),
	mcp.WithString(
		"branch_name",
		mcp.Description("Branch name the commit targets"),
	),
)

func CreateRepositoryFileHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "repository_id", "ref", "commit_title"); err != nil {
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
	commitTitle, err := utils.SafelyConvertToString(arguments["commit_title"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	payload := map[string]any{
		"commit_title": commitTitle,
	}
	for _, key := range []string{"file_path", "content", "commit_body", "is_new_folder", "overlay", "branch_name"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}

	query := map[string]any{}
	if !utils.IsAllDigits(projectID) {
		query["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/blobs/%s", enterpriseID, url.QueryEscape(projectID), url.QueryEscape(ref))
	opts = append(opts, utils.WithPayload(payload), utils.WithQuery(query))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
