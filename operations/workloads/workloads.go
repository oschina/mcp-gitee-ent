package workloads

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const ListWorkloads = "list_workloads"

var ListWorkloadsTool = mcp.NewTool(ListWorkloads,
	mcp.WithDescription("List workload entries within an explicit date range using the v2 workload API. "+
		"Not for member details or issue operation logs."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("start_date", mcp.Description("Start date in YYYY-MM-DD format."), mcp.Required()),
	mcp.WithString("end_date", mcp.Description("End date in YYYY-MM-DD format."), mcp.Required()),
	mcp.WithArray("user_ids",
		mcp.Description("Optional user IDs to filter by."),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithArray("program_ids",
		mcp.Description("Optional program IDs to filter by."),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithNumber("prev_id", mcp.Description("Pagination cursor: last workload log ID from the previous page.")),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListWorkloadsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if checkResult, err := utils.CheckRequired(arguments, "start_date", "end_date"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	startDate, start, err := parseDateArgument(arguments, "start_date")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	endDate, end, err := parseDateArgument(arguments, "end_date")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if end.Before(start) {
		return mcp.NewToolResultError("end_date must be on or after start_date."), nil
	}

	query := map[string]interface{}{
		"registered_dates": startDate + "," + endDate,
	}
	if value, ok := arguments["user_ids"]; ok && value != nil {
		ids, err := joinIDArray(value)
		if err != nil {
			return mcp.NewToolResultError("invalid user_ids: " + err.Error()), nil
		}
		query["user_ids"] = ids
	}
	if value, ok := arguments["program_ids"]; ok && value != nil {
		ids, err := joinIDArray(value)
		if err != nil {
			return mcp.NewToolResultError("invalid program_ids: " + err.Error()), nil
		}
		query["program_ids"] = ids
	}
	for _, key := range []string{"prev_id", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/workloads/v2", enterpriseID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseDateArgument(arguments map[string]any, key string) (string, time.Time, error) {
	value, ok := arguments[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", time.Time{}, fmt.Errorf("Parameter %s must be a YYYY-MM-DD string", key)
	}
	value = strings.TrimSpace(value)
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("Parameter %s must use YYYY-MM-DD format: %w", key, err)
	}
	return value, parsed, nil
}

func joinIDArray(value any) (string, error) {
	items, ok := value.([]any)
	if !ok {
		return "", fmt.Errorf("expected an array of numeric IDs")
	}
	if len(items) == 0 {
		return "", nil
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		id, err := utils.SafelyConvertToInt(item)
		if err != nil {
			return "", err
		}
		ids = append(ids, fmt.Sprintf("%d", id))
	}
	return strings.Join(ids, ","), nil
}
