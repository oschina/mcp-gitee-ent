package common_views

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const ListCommonViews = "list_common_views"

var ListCommonViewsTool = mcp.NewTool(ListCommonViews,
	mcp.WithDescription("List saved common views in an enterprise. Use type program_view for project work item views."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("type", mcp.Description("View type."), mcp.Required(), mcp.Enum("program_view")),
	mcp.WithString("search", mcp.Description("View title keyword.")),
	mcp.WithString("scope", mcp.Description("Visibility scope."), mcp.Enum("hidden", "show", "all")),
	mcp.WithString("public", mcp.Description("Comma-separated visibility: 0 system, 1 private, 2 public.")),
	mcp.WithNumber("program_id", mcp.Description("Optional project ID.")),
	mcp.WithBoolean("use_pagination", mcp.Description("Whether to paginate results.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListCommonViewsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	viewType, ok := arguments["type"].(string)
	if !ok || viewType == "" {
		return mcp.NewToolResultError("Missing required parameter: type"), nil
	}
	query := map[string]interface{}{"type": viewType}
	for _, key := range []string{"search", "scope", "public", "program_id", "use_pagination", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/common_views", enterpriseID), opts...)
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
