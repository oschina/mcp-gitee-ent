package repository

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	UpdateRepositoryFile = "update_repository_file"
)

var UpdateRepositoryFileTool = mcp.NewTool(
	UpdateRepositoryFile,
	mcp.WithDescription("Update one existing file on a branch and commit the change. ref is the branch name; "+
		"previous_path is the current path; file_name is the resulting path/name (may contain /); "+
		"content replaces the whole file; commit_title is the commit message."),
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
		"previous_path",
		mcp.Description("File path before the update"),
		mcp.Required(),
	),
	mcp.WithString(
		"file_name",
		mcp.Description("File name after the update; may contain /"),
		mcp.Required(),
	),
	mcp.WithString(
		"content",
		mcp.Description("File content"),
		mcp.Required(),
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
	mcp.WithString(
		"branch_name",
		mcp.Description("Branch name the commit targets"),
	),
)

func UpdateRepositoryFileHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "repository_id", "ref", "previous_path", "file_name", "content", "commit_title"); err != nil {
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
	previousPath, err := utils.SafelyConvertToString(arguments["previous_path"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	fileName, err := utils.SafelyConvertToString(arguments["file_name"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	content, err := utils.SafelyConvertToString(arguments["content"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	commitTitle, err := utils.SafelyConvertToString(arguments["commit_title"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	payload := map[string]any{
		"previous_path": previousPath,
		"file_name":     fileName,
		"content":       content,
		"commit_title":  commitTitle,
	}
	for _, key := range []string{"commit_body", "branch_name"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}

	query := map[string]any{}
	if !utils.IsAllDigits(projectID) {
		query["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/blobs/%s/update_file", enterpriseID, url.QueryEscape(projectID), url.QueryEscape(ref))
	opts = append(opts, utils.WithPayload(payload), utils.WithQuery(query))
	giteeClient := utils.NewGiteeClient("PUT", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
