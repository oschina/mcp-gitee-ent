package scrum_sprints

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const DeleteScrumSprint = "delete_scrum_sprint"

var DeleteScrumSprintTool = mcp.NewTool(DeleteScrumSprint,
	mcp.WithDescription("Delete one Scrum sprint permanently. The sprint directory is removed and its work items are detached; "+
		"deletion is blocked when the sprint belongs to a baseline."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID the sprint belongs to."), mcp.Required()),
	mcp.WithNumber("scrum_sprint_id", mcp.Description("Sprint ID from list_scrum_sprints."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteScrumSprintHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	programID, err := utils.SafelyConvertToInt(arguments["program_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sprintID, err := utils.SafelyConvertToInt(arguments["scrum_sprint_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient(
		"DELETE",
		fmt.Sprintf("/%d/programs/%d/scrum_sprints/%d", enterpriseID, programID, sprintID),
		opts...,
	)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"scrum_sprint_id": sprintID, "deleted": true})
}
