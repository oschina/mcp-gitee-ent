package issue_states

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListIssueTypeStates = "list_issue_type_states"
)

var ListIssueTypeStatesTool = mcp.NewTool(ListIssueTypeStates,
	mcp.WithDescription("List issue states of an issue type"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"issue_type_id",
		mcp.Description("Issue Type ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at, updated_at)"),
		mcp.Enum("created_at", "updated_at"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending desc: descending)"),
		mcp.Enum("asc", "desc"),
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

func ListIssueTypeStatesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "issue_type_id"); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	issueTypeIDArg := request.Params.Arguments["issue_type_id"]
	issueTypeID, err := utils.SafelyConvertToInt(issueTypeIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/issue_types/%d/issue_states", enterpriseID, issueTypeID)
	opts = append(opts, utils.WithQuery(request.Params.Arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)
	data := types.PagedResponse[types.IssueState]{}
	return giteeClient.HandleMCPResult(&data)
}

