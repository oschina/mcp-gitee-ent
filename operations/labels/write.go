package labels

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateLabel = "create_label"
	UpdateLabel = "update_label"
)

var labelColorPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

var CreateLabelTool = mcp.NewTool(CreateLabel,
	mcp.WithDescription("Create an enterprise label with a name and color. Use list_enterprise_labels to discover existing labels."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("name", mcp.Description("Label name."), mcp.Required()),
	mcp.WithString("color", mcp.Description("Label color in #RGB or #RRGGBB format."), mcp.Required()),
)

var UpdateLabelTool = mcp.NewTool(UpdateLabel,
	mcp.WithDescription("Update an existing label. Both name and color are required by the Gitee API."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("label_id", mcp.Description("Label ID."), mcp.Required()),
	mcp.WithString("name", mcp.Description("Label name."), mcp.Required()),
	mcp.WithString("color", mcp.Description("Label color in #RGB or #RRGGBB format."), mcp.Required()),
)

func CreateLabelHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterpriseAndLabel(nil, request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload, err := labelPayload(arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithPayload(payload))

	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/labels", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateLabelHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterpriseAndLabel([]string{"label_id"}, request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	labelID, err := utils.SafelyConvertToInt(arguments["label_id"])
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload, err := labelPayload(arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithPayload(payload))

	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/labels/%d", enterpriseID, labelID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseEnterpriseAndLabel(required []string, request mcp.CallToolRequest) (map[string]any, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, err
	}
	required = append(required, "enterprise_id", "name", "color")
	for _, key := range required {
		if value, ok := arguments[key]; !ok || value == nil {
			return nil, 0, fmt.Errorf("Missing required parameter: %s", key)
		}
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return nil, 0, err
	}
	return arguments, enterpriseID, nil
}

func labelPayload(arguments map[string]any) (map[string]any, error) {
	name, ok := arguments["name"].(string)
	if !ok || strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("Parameter name must be a non-empty string")
	}
	color, ok := arguments["color"].(string)
	if !ok || !labelColorPattern.MatchString(strings.TrimSpace(color)) {
		return nil, fmt.Errorf("Parameter color must use #RGB or #RRGGBB format")
	}
	return map[string]any{"name": strings.TrimSpace(name), "color": strings.TrimSpace(color)}, nil
}
