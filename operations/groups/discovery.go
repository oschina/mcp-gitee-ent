package groups

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListGroupMembers   = "list_group_members"
	ListGroupProjects  = "list_group_projects"
	ListGroupPrograms  = "list_group_programs"
	GetGroupStatistics = "get_group_statistics"
)

var ListGroupMembersTool = mcp.NewTool(ListGroupMembers,
	mcp.WithDescription("List members of an enterprise team (group)."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("group_id", mcp.Description("Group ID or group path."), mcp.Required()),
	mcp.WithString("search", mcp.Description("Member keyword.")),
	mcp.WithNumber("role_id", mcp.Description("Filter by role ID.")),
	mcp.WithString("scope", mcp.Description("Use not_in to list enterprise members not in this group."), mcp.Enum("not_in")),
	mcp.WithString("sort", mcp.Description("Sort field: created_at or remark.")),
	mcp.WithString("direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListGroupProjectsTool = mcp.NewTool(ListGroupProjects,
	mcp.WithDescription("List repositories associated with an enterprise team (group)."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("group_id", mcp.Description("Group ID or group path."), mcp.Required()),
	mcp.WithString("search", mcp.Description("Repository name keyword.")),
	mcp.WithString("sort", mcp.Description("Sort field: created_at or last_push_at.")),
	mcp.WithString("direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListGroupProgramsTool = mcp.NewTool(ListGroupPrograms,
	mcp.WithDescription("List projects (programs) associated with an enterprise team (group)."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("group_id", mcp.Description("Group ID or group path."), mcp.Required()),
	mcp.WithString("search", mcp.Description("Project name keyword.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetGroupStatisticsTool = mcp.NewTool(GetGroupStatistics,
	mcp.WithDescription("Get repository statistics for an enterprise team: total, forked, and project-group counts."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("group_id", mcp.Description("Group ID or group path."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListGroupMembersHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, groupID, query, err := parseGroupScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	for _, key := range []string{"search", "role_id", "scope", "sort", "direction", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/groups/%s/members", enterpriseID, groupID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListGroupProjectsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, groupID, query, err := parseGroupScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	for _, key := range []string{"search", "sort", "direction", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/groups/%s/projects", enterpriseID, groupID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListGroupProgramsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, groupID, query, err := parseGroupScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	for _, key := range []string{"search", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/groups/%s/group_relate_programs", enterpriseID, groupID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetGroupStatisticsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	_, enterpriseID, groupID, query, err := parseGroupScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/groups/%s/project_statistics", enterpriseID, groupID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseGroupScope(request mcp.CallToolRequest) (map[string]any, int, string, map[string]interface{}, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, "", nil, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, "", nil, err
	}
	value, ok := arguments["group_id"]
	if !ok || value == nil {
		return nil, 0, "", nil, fmt.Errorf("Missing required parameter: group_id")
	}
	groupID, err := utils.SafelyConvertToString(value)
	if err != nil || groupID == "" {
		return nil, 0, "", nil, fmt.Errorf("Missing required parameter: group_id")
	}
	query := map[string]interface{}{}
	if !utils.IsAllDigits(groupID) {
		query["qt"] = "path"
	}
	return arguments, enterpriseID, groupID, query, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
