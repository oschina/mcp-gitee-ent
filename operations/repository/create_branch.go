package repository

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateBranch = "create_project_branch"
)

var CreateBranchTool = mcp.NewTool(
	CreateBranch,
	mcp.WithDescription("Create a branch in a repository."),
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
		"branch_name",
		mcp.Description("New branch name"),
		mcp.Required(),
	),
	mcp.WithString(
		"refs",
		mcp.Description("Start point branch or commit"),
		mcp.Required(),
	),
	mcp.WithString(
		"action",
		mcp.Description("Branch action"),
		mcp.Enum("new", "recovery"),
	),
	mcp.WithNumber(
		"type",
		mcp.Description("Branch type (0: normal, 1: protected, 2: read-only)"),
	),
	mcp.WithString(
		"remark",
		mcp.Description("Branch remark"),
	),
	mcp.WithString(
		"new_name",
		mcp.Description("New name when renaming the branch"),
	),
	mcp.WithNumber(
		"ref_project_id",
		mcp.Description("Source project ID the start point belongs to"),
	),
)

func CreateBranchHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "repository_id", "branch_name", "refs"); err != nil {
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
	branchName, err := utils.SafelyConvertToString(arguments["branch_name"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	refs, err := utils.SafelyConvertToString(arguments["refs"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	payload := map[string]any{
		"branch_name": branchName,
		"refs":        refs,
	}
	for _, key := range []string{"action", "type", "remark", "new_name", "ref_project_id"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}

	query := map[string]any{}
	if !utils.IsAllDigits(projectID) {
		query["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/branches", enterpriseID, url.QueryEscape(projectID))
	opts = append(opts, utils.WithPayload(payload), utils.WithQuery(query))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
