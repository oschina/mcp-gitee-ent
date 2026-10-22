package issues

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CommentIssue = "comment_enterprise_issue"
)

var CommentIssueTool = mcp.NewTool(CommentIssue,
	mcp.WithDescription("Comment on an issue"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"issue_id",
		mcp.Description("Issue ID or Ident"),
		mcp.Required(),
	),
	mcp.WithString(
		"body",
		mcp.Description("Comment content"),
		mcp.Required(),
	),
)

func CommentIssueHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "issue_id", "body"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	issueIDArg := arguments["issue_id"]
	issueID, err := utils.SafelyConvertToString(issueIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	body := arguments["body"]

	apiUrl := fmt.Sprintf("/%d/issues/%s/notes", enterpriseID, issueID)

	payload := map[string]interface{}{
		"body": body,
	}
	if !utils.IsAllDigits(issueID) {
		payload["qt"] = "ident"
	}

	opts = append(opts, utils.WithPayload(payload))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.IssueComment{}
	return giteeClient.HandleMCPResult(&data)
}
