package labels

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEnterpriseLabels = "list_enterprise_labels"
)

var ListEnterpriseLabelsTool = mcp.NewTool(ListEnterpriseLabels,
	mcp.WithDescription("List labels of an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at, updated_at, serial)"),
		mcp.Enum("created_at", "updated_at", "serial"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending, desc: descending)"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search keyword"),
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

func ListEnterpriseLabelsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/labels", enterpriseID)
	opts = append(opts, utils.WithQuery(request.Params.Arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.Label]{}
	return giteeClient.HandleMCPResult(&data)
}

