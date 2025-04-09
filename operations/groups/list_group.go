package groups

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEntGroup = "list_enterprise_groups"
)

var ListEntGroupsTool = mcp.NewTool(ListEntGroup,
	mcp.WithDescription("List groups in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"qt",
		mcp.Description("Query type (path/id)"),
		mcp.Enum("path", "id"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at: created time, updated_at: updated time)"),
		mcp.Enum("created_at", "updated_at"),
	),
	mcp.WithNumber(
		"program_id",
		mcp.Description("Program ID"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending, desc: descending)"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search string"),
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

func ListEntGroupsHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	apiUrl := fmt.Sprintf("/%d/groups", enterpriseID)
	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	// Handle response
	data := types.PagedResponse[types.Group]{}
	return giteeClient.HandleMCPResult(&data)
}
