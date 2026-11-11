package week_reports

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SaveWeekReport = "save_week_report"
)

var SaveWeekReportTool = mcp.NewTool(
	SaveWeekReport,
	mcp.WithDescription("Create or overwrite a user's week report for a given week. This is an upsert: "+
		"if that week's report does not exist it is created, otherwise it is overwritten."),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"username",
		mcp.Description("User personalized domain or login"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"year",
		mcp.Description("Report year"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"week_index",
		mcp.Description("Week index within the year"),
		mcp.Required(),
	),
	mcp.WithString(
		"content",
		mcp.Description("Week report content"),
		mcp.Required(),
	),
	mcp.WithArray(
		"pull_request_ids",
		mcp.Description("Related pull request IDs"),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithArray(
		"issue_ids",
		mcp.Description("Related issue IDs"),
		mcp.Items(map[string]any{"type": "number"}),
	),
	mcp.WithArray(
		"event_ids",
		mcp.Description("Related event IDs"),
		mcp.Items(map[string]any{"type": "number"}),
	),
)

func SaveWeekReportHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "username", "year", "week_index", "content"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	username, err := utils.SafelyConvertToString(arguments["username"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	year, err := utils.SafelyConvertToInt(arguments["year"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	weekIndex, err := utils.SafelyConvertToInt(arguments["week_index"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	content, err := utils.SafelyConvertToString(arguments["content"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	payload := map[string]any{"content": content}
	for _, key := range []string{"pull_request_ids", "issue_ids", "event_ids"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}

	apiUrl := fmt.Sprintf("/%d/week_reports/%s/%d/%d", enterpriseID, url.QueryEscape(username), year, weekIndex)
	opts = append(opts, utils.WithPayload(payload))
	giteeClient := utils.NewGiteeClient("PUT", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
