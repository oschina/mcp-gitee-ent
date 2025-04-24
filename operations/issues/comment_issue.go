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
		mcp.Description("Issue Ident"),
		mcp.Required(),
	),
	mcp.WithString(
		"body",
		mcp.Description("Comment content"),
		mcp.Required(),
	),
)

func CommentIssueHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "issue_id", "body"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	issueID := request.Params.Arguments["issue_id"]
	request.Params.Arguments["qt"] = "ident"

	apiUrl := fmt.Sprintf("/%d/issues/%s/notes", enterpriseID, issueID)
	opts = append(opts, utils.WithPayload(request.Params.Arguments))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.IssueComment{}
	return giteeClient.HandleMCPResult(&data)
}

