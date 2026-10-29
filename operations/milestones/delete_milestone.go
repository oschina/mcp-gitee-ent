package milestones

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const DeleteMilestone = "delete_milestone"

var DeleteMilestoneTool = mcp.NewTool(DeleteMilestone,
	mcp.WithDescription("Delete one milestone permanently. Work items that referenced it are unlinked, not deleted."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("milestone_id", mcp.Description("Milestone ID from list_milestones."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteMilestoneHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	milestoneID, err := utils.SafelyConvertToInt(arguments["milestone_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/milestones/%d", enterpriseID, milestoneID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"milestone_id": milestoneID, "deleted": true})
}
