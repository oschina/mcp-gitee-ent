package issues

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	UpdateIssue = "update_enterprise_issue"
)

var UpdateIssueTool = mcp.NewTool(UpdateIssue,
	mcp.WithDescription("Update an issue in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"issue_id",
		mcp.Description("Issue Ident"),
		mcp.Required(),
	),
	mcp.WithString(
		"title",
		mcp.Description("Issue title"),
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
		"issue_state_id",
		mcp.Description("Issue state ID"),
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
		"scrum_sprint_id",
		mcp.Description("Sprint ID"),
	),
	mcp.WithNumber(
		"scrum_version_id",
		mcp.Description("Version ID"),
	),
	mcp.WithNumber(
		"estimated_duration",
		mcp.Description("Estimated duration (in hours, supports two decimal places)"),
	),
)

func UpdateIssueHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "issue_id"); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	issueIDArg := arguments["issue_id"]
	arguments["qt"] = "ident"

	apiUrl := fmt.Sprintf("/%d/issues/%s", enterpriseID, issueIDArg)
	opts = append(opts, utils.WithPayload(arguments))
	giteeClient := utils.NewGiteeClient("PUT", apiUrl, opts...)

	data := types.BasicIssue{}
	return giteeClient.HandleMCPResult(&data)
}
