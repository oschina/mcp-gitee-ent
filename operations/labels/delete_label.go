package labels

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const DeleteLabel = "delete_label"

var DeleteLabelTool = mcp.NewTool(DeleteLabel,
	mcp.WithDescription("Delete an enterprise label permanently. The Gitee API refuses to delete a label that is still "+
		"used by work items, so remove it from those issues first."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("label_id", mcp.Description("Label ID from list_enterprise_labels."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteLabelHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	labelID, err := utils.SafelyConvertToInt(arguments["label_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/labels/%d", enterpriseID, labelID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"label_id": labelID, "deleted": true})
}
