package members

import (
	"context"
	"fmt"
	"net/url"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListMemberEvents = "list_member_events"
)

var ListMemberEventsTool = mcp.NewTool(ListMemberEvents,
	mcp.WithDescription("List events for an enterprise member."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("user_id", mcp.Description("User ID or username."), mcp.Required()),
	mcp.WithString("start_date", mcp.Description("Start date (format: yyyy-mm-dd)")),
	mcp.WithString("end_date", mcp.Description("End date (format: yyyy-mm-dd)")),
	mcp.WithNumber("limit", mcp.Description("Maximum number of events to return."), mcp.Min(1), mcp.Max(100)),
	mcp.WithString("prev_id", mcp.Description("Cursor: return events older than this event ID. Pass the exact (19-digit) event id string from the previous page; ids exceed float64 precision so pass them as strings.")),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListMemberEventsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "user_id"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	userID, err := utils.SafelyConvertToString(arguments["user_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	query := map[string]any{}
	for _, key := range []string{"start_date", "end_date", "prev_id", "limit"} {
		if value, ok := arguments[key]; ok && value != nil {
			query[key] = value
		}
	}
	if !utils.IsAllDigits(userID) {
		query["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/members/%s/events", enterpriseID, url.QueryEscape(userID))
	opts = append(opts, utils.WithQuery(query))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
