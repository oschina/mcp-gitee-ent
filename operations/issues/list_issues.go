package issues

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListIssues = "list_enterprise_issues"
)

var ListIssuesTool = mcp.NewTool(ListIssues,
	mcp.WithDescription("List issues in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_id",
		mcp.Description("Associated Repository ID"),
	),
	mcp.WithString(
		"program_id",
		mcp.Description("Associated Project ID (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"state",
		mcp.Description("Task status attributes, multiple can be selected, separated by commas. (Open: open Closed: closed Rejected: rejected In Progress: progressing)"),
		mcp.Enum("open", "closed", "rejected", "progressing"),
	),
	mcp.WithString(
		"only_related_me",
		mcp.Description("Whether to list only tasks related to the authorized user (0: No 1: Yes)"),
		mcp.Enum("0", "1"),
	),
	mcp.WithString(
		"assignee_id",
		mcp.Description("Assignee ID"),
	),
	mcp.WithString(
		"author_id",
		mcp.Description("Creator ID"),
	),
	mcp.WithString(
		"collaborator_ids",
		mcp.Description("Collaborator's ids. (Comma-separated ID string)"),
	),
	mcp.WithString(
		"created_at",
		mcp.Description("Creation time, format: (interval)yyyymmddTHH:MM:SS+08:00-yyyymmddTHH:MM:SS+08:00, (specific date)yyyymmddTHH:MM:SS+08:00, (less than specified date)<yyyymmddTHH:MM:SS+08:00, (greater than specified date)>yyyymmddTHH:MM:SS+08:00"),
	),
	mcp.WithString(
		"finished_at",
		mcp.Description("Task completion date, format as above"),
	),
	mcp.WithString(
		"plan_started_at",
		mcp.Description("Planned start time, (format: yyyy-mm-dd)"),
	),
	mcp.WithString(
		"deadline",
		mcp.Description("Task deadline, (format: yyyy-mm-dd)"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search parameter"),
	),
	mcp.WithString(
		"filter_child",
		mcp.Description("Whether to filter subtasks (0: No, 1: Yes)"),
	),
	mcp.WithString(
		"issue_state_ids",
		mcp.Description("Task status IDs, multiple selection, separated by commas."),
	),
	mcp.WithString(
		"issue_type_id",
		mcp.Description("Task type"),
	),
	mcp.WithString(
		"label_ids",
		mcp.Description("Label IDs (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"priority",
		mcp.Description("Priority (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"scrum_sprint_ids",
		mcp.Description("Sprint IDs (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"scrum_version_ids",
		mcp.Description("Version IDs (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"kanban_ids",
		mcp.Description("Kanban IDs (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"kanban_column_ids",
		mcp.Description("Kanban column IDs (multiple can be selected, separated by commas)"),
	),
	mcp.WithString(
		"sort",
		mcp.Description("Sort field (created_at, updated_at, deadline, priority)"),
	),
	mcp.WithString(
		"direction",
		mcp.Description("Sort direction (asc: ascending desc: descending)"),
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

func ListIssuesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	if checkResult, err := utils.CheckRequired(request.Params.Arguments); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	request.Params.Arguments["show_scrum_sprints"] = true

	apiUrl := fmt.Sprintf("/%d/issues", enterpriseID)

	opts = append(opts, utils.WithQuery(request.Params.Arguments))

	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.BasicIssue]{}

	return giteeClient.HandleMCPResult(&data)
}
