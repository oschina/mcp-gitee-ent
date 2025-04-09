package repository

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEnterpriseProjects = "list_enterprise_repositories"
)

var ListRepositoriesTool = mcp.NewTool(ListEnterpriseProjects,
	mcp.WithDescription("List repositories in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"scope",
		mcp.Description("Scope filter"),
		mcp.Enum("private", "public", "internal-open", "not-belong-any-program", "outsources", "all"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search parameter"),
	),
	mcp.WithString(
		"type",
		mcp.Description("Relationship type: created - created by me, joined - joined by me, star - starred by me, template - template repositories, personal_namespace - repositories under personal namespace in enterprise"),
		mcp.Enum("joined", "created", "star", "template", "personal_namespace"),
	),
	mcp.WithNumber(
		"status",
		mcp.Description("Status: 0 - open, 1 - paused, 2 - closed"),
	),
	mcp.WithNumber(
		"creator_id",
		mcp.Description("Creator ID"),
	),
	mcp.WithNumber(
		"parent_id",
		mcp.Description("Parent repository ID for forked repositories"),
	),
	mcp.WithString(
		"fork_filter",
		mcp.Description("Fork filter: all, not_fork, only_fork, my_fork"),
		mcp.Enum("all", "not_fork", "only_fork", "my_fork"),
	),
	mcp.WithNumber(
		"outsourced",
		mcp.Description("Outsourced status: 0 - internal, 1 - outsourced"),
	),
	mcp.WithNumber(
		"group_id",
		mcp.Description("Team/Group ID"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field: created_at, last_push_at"),
		mcp.Enum("created_at", "last_push_at"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction: asc - ascending, desc - descending"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithString(
		"namespace_scope",
		mcp.Description("Namespace scope, used with group_id: belongs_to - includes sub-levels, only_this - includes only current level"),
		mcp.Enum("belongs_to", "only_this"),
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

func ListRepositoriesHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if checkResult, err := utils.CheckRequired(request.Params.Arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/projects", enterpriseID)

	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	data := types.PagedResponse[types.Repository]{}

	return giteeClient.HandleMCPResult(&data)
}
