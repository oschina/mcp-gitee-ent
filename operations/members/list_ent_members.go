package members

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListEntMembers = "list_enterprise_members"
)

var ListEntMembersTool = mcp.NewTool(ListEntMembers,
	mcp.WithDescription("List members of an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"ident",
		mcp.Description("Role type (viewer: viewer, member: member, outsourced_member: outsourced member, human_resources: HR admin, admin: admin, super_admin: super admin, enterprise_owner: enterprise owner)"),
		mcp.Enum("viewer", "member", "outsourced_member", "human_resources", "admin", "super_admin", "enterprise_owner"),
	),
	mcp.WithNumber(
		"is_block",
		mcp.Description("Filter blocked users (1: blocked users)"),
	),
	mcp.WithNumber(
		"group_id",
		mcp.Description("Filter members by team ID"),
	),
	mcp.WithNumber(
		"role_id",
		mcp.Description("Filter members by role ID"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search keyword"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at: creation time, remark: enterprise remark, role: role, occupation: position, block_status: block status)"),
		mcp.Enum("created_at", "remark", "role", "occupation", "block_status"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending, desc: descending)"),
		mcp.Enum("asc", "desc"),
	),
	mcp.WithBoolean(
		"include_member_histories",
		mcp.Description("Include resigned member histories (true/false)"),
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

func ListEntMembersHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/members", enterpriseID)
	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	// Handle response
	data := types.PagedResponse[types.EnterpriseMember]{}
	return giteeClient.HandleMCPResult(&data)
}
