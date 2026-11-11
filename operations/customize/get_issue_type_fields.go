package customize

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetIssueTypeCustomFields = "get_issue_type_custom_fields"
)

var GetIssueTypeCustomFieldsTool = mcp.NewTool(
	GetIssueTypeCustomFields,
	mcp.WithDescription("List the custom fields configured for one issue type. Returns each field's id, name, field_type, options and description, plus builtin fields, so callers know which fields can be set or filtered before creating or querying issues."),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"issue_type_id",
		mcp.Description("Issue type ID"),
		mcp.Required(),
	),
)

func GetIssueTypeCustomFieldsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "issue_type_id"); err != nil {
		return checkResult, err
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	issueTypeID, err := utils.SafelyConvertToInt(arguments["issue_type_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/customize/%d/fields", enterpriseID, issueTypeID)
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	var data any
	return giteeClient.HandleMCPResultContext(ctx, &data)
}
