package pulls

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEntPulls = "list_enterprise_pulls"
)

var ListEntPullsTool = mcp.NewTool(ListEntPulls,
	mcp.WithDescription("List enterprise pull requests"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"state",
		mcp.Description("PR state"),
		mcp.Enum("opened", "closed", "merged"),
		mcp.DefaultString("opened"),
	),
	mcp.WithString(
		"scope",
		mcp.Description("Scope filter"),
		mcp.Enum("assigned_or_test", "related_to_me", "participate_in", "draft", "create", "assign", "test"),
	),
	mcp.WithString(
		"author_id",
		mcp.Description("Filter PR creator, Multiple creators are separated by commas."),
	),
	mcp.WithString(
		"assignee_id",
		mcp.Description("Filter PR reviewer, Multiple reviewers are separated by commas."),
	),
	mcp.WithString(
		"tester_id",
		mcp.Description("Filter PR testers, Multiple testers are separated by commas."),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search parameter"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at, closed_at, priority, updated_at)"),
		mcp.Enum("created_at", "closed_at", "priority", "updated_at"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending, desc: descending)"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithNumber(
		"group_id",
		mcp.Description("Team/Group ID"),
	),
	mcp.WithString(
		"labels",
		mcp.Description("Label names. Multiple labels are separated by commas."),
	),
	mcp.WithNumber(
		"can_be_merged",
		mcp.Description("Whether the PR can be merged, 0: no, 1: yes"),
	),
	mcp.WithString(
		"created_at",
		mcp.Description("Creation time, e.g.: \"2023.12.18-2023.12.31\""),
	),
	mcp.WithString(
		"updated_at",
		mcp.Description("Update time, e.g.: \"2023.12.18-2023.12.31\""),
	),
	mcp.WithString(
		"merged_at",
		mcp.Description("Merge time, e.g.: \"2023.12.18-2023.12.31\""),
	),
	mcp.WithString(
		"target_branch",
		mcp.Description("Target branch name"),
	),
	mcp.WithString(
		"source_branch",
		mcp.Description("Source branch name"),
	),
	mcp.WithNumber(
		"project_id",
		mcp.Description("Repository ID"),
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

func ListEntPullsHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/pull_requests", enterpriseID)
	request.Params.Arguments["pr_qt"] = "iid"
	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	// Handle response
	data := types.PagedResponse[types.BasicPull]{}
	return giteeClient.HandleMCPResult(&data)
}
