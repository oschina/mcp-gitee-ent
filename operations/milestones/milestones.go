package milestones

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListMilestones = "list_milestones"
	GetMilestone   = "get_milestone"
)

var ListMilestonesTool = mcp.NewTool(ListMilestones,
	mcp.WithDescription("List milestones or search them by keyword. Not for Scrum sprints or Scrum versions."),
	mcp.WithNumber("enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithNumber("program_id", mcp.Description("Optional program ID.")),
	mcp.WithNumber("project_id", mcp.Description("Optional repository ID.")),
	mcp.WithString("state",
		mcp.Description("Milestone state: active, pending, processing, closed, overdued, not_overdue."),
		mcp.Enum("active", "pending", "processing", "closed", "overdued", "not_overdue"),
	),
	mcp.WithString("search", mcp.Description("Keyword to search milestone titles.")),
	mcp.WithString("sort", mcp.Description("Sort field: title, created_at, start_date, due_date.")),
	mcp.WithString("direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetMilestoneTool = mcp.NewTool(GetMilestone,
	mcp.WithDescription("Get one milestone by ID. Use list_milestones to discover milestone IDs."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("milestone_id", mcp.Description("Milestone ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListMilestonesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	if checkResult, err := utils.CheckRequired(arguments); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	query := make(map[string]interface{})
	for _, key := range []string{"program_id", "project_id", "state", "search", "sort", "direction", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/milestones", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetMilestoneHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	if checkResult, err := utils.CheckRequired(arguments, "milestone_id"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	milestoneID, err := utils.SafelyConvertToInt(arguments["milestone_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/milestones/%d", enterpriseID, milestoneID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}
