package issues

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetIssueDetail = "get_enterprise_issue_detail"
)

var GetIssueDetailTool = mcp.NewTool(GetIssueDetail,
	mcp.WithDescription("Get issue detail"),
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
)

func GetIssueDetailHandleFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Validate required parameters
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "issue_id"); err != nil {
		return checkResult, err
	}
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	issueID := request.Params.Arguments["issue_id"]
	request.Params.Arguments["qt"] = "ident"

	apiUrl := fmt.Sprintf("/%d/issues/%s", enterpriseID, issueID)
	giteeClient := utils.NewGiteeClient("GET", apiUrl, utils.WithQuery(request.Params.Arguments))

	data := types.IssueDetail{}

	return giteeClient.HandleMCPResult(&data)
}
