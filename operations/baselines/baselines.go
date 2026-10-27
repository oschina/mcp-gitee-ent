package baselines

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListBaselines = "list_baselines"
	GetBaseline   = "get_baseline"
)

var ListBaselinesTool = mcp.NewTool(ListBaselines,
	mcp.WithDescription("List project baselines. Filter by status, baseline type, keyword, sprint, assignee, or creator."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithNumber("status", mcp.Description("Baseline status: 0 planning, 1 published."), mcp.Min(0), mcp.Max(1)),
	mcp.WithNumber("baseline_type", mcp.Description("Baseline type: 0 project baseline, 1 process baseline."), mcp.Min(0), mcp.Max(1)),
	mcp.WithString("search", mcp.Description("Baseline title keyword.")),
	mcp.WithNumber("sprint_id", mcp.Description("Filter process baselines by sprint ID.")),
	mcp.WithString("assignee_ids", mcp.Description("Comma-separated assignee IDs.")),
	mcp.WithString("creator_ids", mcp.Description("Comma-separated creator IDs.")),
	mcp.WithString("parent_ids", mcp.Description("Comma-separated parent baseline IDs.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetBaselineTool = mcp.NewTool(GetBaseline,
	mcp.WithDescription("Get one project baseline by ID."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithNumber("baseline_id", mcp.Description("Baseline ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListBaselinesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	for _, key := range []string{
		"status", "baseline_type", "search", "sprint_id",
		"assignee_ids", "creator_ids", "parent_ids", "page", "per_page",
	} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/programs/%d/baselines", enterpriseID, programID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetBaselineHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, programID, err := parseScope(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	baselineID, err := requiredInt(arguments, "baseline_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"program_id": programID}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/programs/%d/baselines/%d", enterpriseID, programID, baselineID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseScope(request mcp.CallToolRequest) (map[string]any, int, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, 0, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, 0, err
	}
	programID, err := requiredInt(arguments, "program_id")
	if err != nil {
		return nil, 0, 0, err
	}
	return arguments, enterpriseID, programID, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
