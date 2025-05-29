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

func GetIssueDetailHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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

	apiUrl := fmt.Sprintf("/%d/issues/%s", enterpriseID, issueIDArg)
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.IssueDetail{}
	return giteeClient.HandleMCPResult(&data)
}
