package scrum_versions

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListScrumVersions = "list_scrum_versions"
)

var ListScrumVersionsTool = mcp.NewTool(ListScrumVersions,
	mcp.WithDescription("List Scrum Versions"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"program_id",
		mcp.Description("Program ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"states",
		mcp.Description("States"),
		mcp.Enum("open", "progressing", "closed"),
	),
	mcp.WithString(
		"search",
		mcp.Description("Search"),
	),
	mcp.WithNumber(
		"page",
		mcp.Description("Page"),
		mcp.DefaultNumber(1),
	),
	mcp.WithNumber(
		"per_page",
		mcp.Description("Per Page"),
		mcp.DefaultNumber(20),
	),
)

func ListScrumVersionsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "program_id"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	programIDArg := arguments["program_id"]
	programID, err := utils.SafelyConvertToInt(programIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/programs/%d/scrum_versions", enterpriseID, programID)
	opts = append(opts, utils.WithQuery(arguments))
	giteeClient := utils.NewGiteeClient("GET", apiUrl, opts...)

	data := types.PagedResponse[types.ScrumVersion]{}
	return giteeClient.HandleMCPResult(&data)
}
