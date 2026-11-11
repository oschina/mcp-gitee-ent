package week_reports

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetWeekReport = "get_week_report"
)

var GetWeekReportTool = mcp.NewTool(
	GetWeekReport,
	mcp.WithDescription("Get one week report for a user."),
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
)

func GetWeekReportHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "username", "year", "week_index"); err != nil {
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

	apiUrl := fmt.Sprintf("/%d/week_reports/%s/%d/%d", enterpriseID, url.QueryEscape(username), year, weekIndex)
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
