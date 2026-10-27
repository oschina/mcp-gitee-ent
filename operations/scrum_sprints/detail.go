package scrum_sprints

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const GetScrumSprint = "get_scrum_sprint"

var GetScrumSprintTool = mcp.NewTool(GetScrumSprint,
	mcp.WithDescription("Get one Scrum sprint by ID, including its issue statistics."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithNumber("scrum_sprint_id", mcp.Description("Sprint ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func GetScrumSprintHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	programID, err := requiredInt(arguments, "program_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sprintID, err := requiredInt(arguments, "scrum_sprint_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/programs/%d/scrum_sprints/%d", enterpriseID, programID, sprintID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
