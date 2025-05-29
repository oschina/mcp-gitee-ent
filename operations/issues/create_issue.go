package issues

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateIssue = "create_enterprise_issue"
)

var CreateIssueTool = mcp.NewTool(CreateIssue,
	mcp.WithDescription("Create an issue in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"title",
		mcp.Description("Issue title"),
		mcp.Required(),
	),
	mcp.WithString(
		"description",
		mcp.Description("Issue description"),
	),
	mcp.WithNumber(
		"assignee_id",
		mcp.Description("Assignee user ID"),
	),
	mcp.WithString(
		"collaborator_ids",
		mcp.Description("Collaborator user IDs, comma separated (e.g. 1,2,3)"),
	),
	mcp.WithNumber(
		"issue_type_id",
		mcp.Description("Issue type ID"),
	),
	mcp.WithNumber(
		"program_id",
		mcp.Description("Program ID"),
	),
	mcp.WithNumber(
		"project_id",
		mcp.Description("Project ID"),
	),
	mcp.WithString(
		"label_ids",
		mcp.Description("Label IDs, comma separated (e.g. 1,2,3)"),
	),
	mcp.WithNumber(
		"priority",
		mcp.Description("Priority (0: Unspecified, 1: Trivial, 2: Minor, 3: Major, 4: Critical)"),
	),
	mcp.WithNumber(
		"parent_id",
		mcp.Description("Parent issue ID"),
	),
	mcp.WithString(
		"branch",
		mcp.Description("Related branch name"),
	),
	mcp.WithString(
		"plan_started_at",
		mcp.Description("Planned start date (format: yyyy-mm-ddTHH:MM:SS)"),
	),
	mcp.WithString(
		"deadline",
		mcp.Description("Planned completion date (format: yyyy-mm-ddTHH:MM:SS)"),
	),
	mcp.WithString(
		"started_at",
		mcp.Description("Actual start time (format: yyyy-mm-ddTHH:MM:SS)"),
	),
	mcp.WithString(
		"finished_at",
		mcp.Description("Actual completion time (format: yyyy-mm-ddTHH:MM:SS)"),
	),
	mcp.WithNumber(
		"kanban_id",
		mcp.Description("Kanban ID"),
	),
	mcp.WithNumber(
		"kanban_column_id",
		mcp.Description("Kanban column ID"),
	),
	mcp.WithNumber(
		"scrum_sprint_id",
		mcp.Description("Sprint ID"),
	),
	mcp.WithNumber(
		"scrum_version_id",
		mcp.Description("Version ID"),
	),
	mcp.WithNumber(
		"duration",
		mcp.Description("Estimated duration in minutes"),
	),
	mcp.WithNumber(
		"pull_request_id",
		mcp.Description("Pull Request ID"),
	),
	mcp.WithString(
		"category",
		mcp.Description("Issue category"),
		mcp.Enum("task", "bug", "requirement"),
	),
	mcp.WithNumber(
		"link_issue_id",
		mcp.Description("Related issue ID to link"),
	),
)

func CreateIssueHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate required parameters
	if checkResult, err := utils.CheckRequired(arguments, "title"); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/issues", enterpriseID)
	opts = append(opts, utils.WithPayload(arguments))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.BasicIssue{}
	return giteeClient.HandleMCPResult(&data)
}
