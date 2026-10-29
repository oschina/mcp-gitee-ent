package members

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SearchMembers         = "search_members"
	GetMember             = "get_member"
	GetMemberByEmail      = "get_member_by_email"
	ListMemberMemberships = "list_member_memberships"
	GetGroup              = "get_group"
)

var SearchMembersTool = mcp.NewTool(SearchMembers,
	mcp.WithDescription("Search enterprise members. Use get_member_by_email for an exact email lookup, "+
		"or list_enterprise_members for an unfiltered list."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("search", mcp.Description("Search by name, username, remark, pinyin, or email.")),
	mcp.WithArray("program_ids",
		mcp.Description("Optional program IDs to restrict the search to."),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithArray("project_ids",
		mcp.Description("Optional repository IDs to restrict the search to."),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetMemberTool = mcp.NewTool(GetMember,
	mcp.WithDescription("Get one enterprise member by user ID or username. "+
		"Not for listing the member's projects, programs, or groups; use list_member_memberships for that."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("user_id", mcp.Description("User ID or username."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetMemberByEmailTool = mcp.NewTool(GetMemberByEmail,
	mcp.WithDescription("Find an enterprise member by exact email address."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("email", mcp.Description("Exact member email address."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListMemberMembershipsTool = mcp.NewTool(ListMemberMemberships,
	mcp.WithDescription("List one membership collection for a member: projects (repositories), programs, or groups. "+
		"Use type=projects for repositories; the API does not accept repositories as a type value."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("user_id", mcp.Description("User ID or username."), mcp.Required()),
	mcp.WithString("type",
		mcp.Description("Membership collection to list."),
		mcp.Required(),
		mcp.Enum("projects", "programs", "groups"),
	),
	mcp.WithString("scope", mcp.Description("Optional scope filter. Use not_joined for resources the member has not joined; groups also accept admin.")),
	mcp.WithString("search", mcp.Description("Optional keyword filter.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetGroupTool = mcp.NewTool(GetGroup,
	mcp.WithDescription("Get enterprise team details by group ID or path. Use list_enterprise_groups to discover teams."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("group_id", mcp.Description("Group ID or group path."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func SearchMembersHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := make(map[string]any, len(arguments))
	for key, value := range arguments {
		if key == "enterprise_id" || value == nil {
			continue
		}
		payload[key] = value
	}
	opts = append(opts, utils.WithPayload(payload))

	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/members/search_members", enterpriseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetMemberHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userID, err := requiredMemberIdentifier(arguments, "user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/members/%s", enterpriseID, userID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetMemberByEmailHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	email, err := requiredMemberIdentifier(arguments, "email")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(map[string]interface{}{"email": email}))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/members/by_email", enterpriseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListMemberMembershipsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	userID, err := requiredMemberIdentifier(arguments, "user_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	membershipType, err := requiredMemberIdentifier(arguments, "type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	switch membershipType {
	case "projects", "programs", "groups":
	default:
		return mcp.NewToolResultError(fmt.Sprintf("Invalid type %q. Allowed values: projects, programs, groups.", membershipType)), nil
	}

	query := make(map[string]interface{})
	for _, key := range []string{"scope", "search", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/members/%s/%s", enterpriseID, userID, membershipType),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetGroupHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	groupID, err := requiredMemberIdentifier(arguments, "group_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !utils.IsAllDigits(groupID) {
		opts = append(opts, utils.WithQuery(map[string]interface{}{"qt": "path"}))
	}

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/groups/%s", enterpriseID, groupID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseEnterprise(request mcp.CallToolRequest) (map[string]any, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, err
	}
	value, ok := arguments["enterprise_id"]
	if !ok || value == nil {
		return nil, 0, fmt.Errorf("Missing required parameter: enterprise_id")
	}
	enterpriseID, err := utils.SafelyConvertToInt(value)
	if err != nil {
		return nil, 0, err
	}
	return arguments, enterpriseID, nil
}

func requiredMemberIdentifier(arguments map[string]any, key string) (string, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return "", fmt.Errorf("Missing required parameter: %s", key)
	}
	identifier, err := utils.SafelyConvertToString(value)
	if err != nil {
		return "", err
	}
	if identifier == "" {
		return "", fmt.Errorf("Parameter %s must not be empty", key)
	}
	return identifier, nil
}
