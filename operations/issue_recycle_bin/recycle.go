package issue_recycle_bin

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const RestoreIssues = "restore_issues"

var RestoreIssuesTool = mcp.NewTool(RestoreIssues,
	mcp.WithDescription("Restore deleted issues from the enterprise recycle bin."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithArray("issue_ids", mcp.Description("Issue IDs to restore."), mcp.Required(), mcp.Items(map[string]any{"type": "number"})),
	mcp.WithBoolean("recover_children", mcp.Description("Whether to restore child issues as well.")),
)

func RestoreIssuesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	issueIDs, err := intArray(arguments["issue_ids"])
	if err != nil || len(issueIDs) == 0 {
		return mcp.NewToolResultError("issue_ids must contain at least one issue ID"), nil
	}
	ids := make([]string, 0, len(issueIDs))
	for _, id := range issueIDs {
		ids = append(ids, fmt.Sprintf("%d", id))
	}
	payload := map[string]any{"issue_ids": strings.Join(ids, ",")}
	if value, ok := arguments["recover_children"]; ok && value != nil {
		payload["recover_children"] = value
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"PUT",
		fmt.Sprintf("/%d/issue_recycle_bin/recover_issue", enterpriseID),
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

func intArray(value any) ([]int, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array of numeric IDs")
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		id, err := utils.SafelyConvertToInt(item)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}
