package scrum_sprints

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateScrumSprint = "create_scrum_sprint"
)

var CreateScrumSprintTool = mcp.NewTool(CreateScrumSprint,
	mcp.WithDescription("Create a Scrum Sprint"),
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
		"title",
		mcp.Description("Title"),
		mcp.Required(),
	),
	mcp.WithNumber(
		"assignee_id",
		mcp.Description("Assignee ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"started_at",
		mcp.Description("Started At"),
		mcp.Required(),
	),
	mcp.WithString(
		"finished_at",
		mcp.Description("Finished At"),
		mcp.Required(),
	),
	mcp.WithString(
		"description",
		mcp.Description("Description"),
	),
	mcp.WithNumber(
		"time_scale",
		mcp.Description("Time Scale, unit: hours, support two decimal places"),
	),
)

func CreateScrumSprintHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	// 安全转换参数类型
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if checkResult, err := utils.CheckRequired(arguments, "program_id", "title", "assignee_id", "started_at", "finished_at"); err != nil {
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

	apiUrl := fmt.Sprintf("/%d/programs/%d/scrum_sprints", enterpriseID, programID)
	opts = append(opts, utils.WithPayload(arguments))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.ScrumSprint{}
	return giteeClient.HandleMCPResult(&data)
}
