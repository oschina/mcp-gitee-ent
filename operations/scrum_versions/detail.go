package scrum_versions

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetScrumVersion = "get_scrum_version"
	ListScrumStages = "list_scrum_stages"
)

var GetScrumVersionTool = mcp.NewTool(GetScrumVersion,
	mcp.WithDescription("Get one Scrum version by ID."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithNumber("scrum_version_id", mcp.Description("Version ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListScrumStagesTool = mcp.NewTool(ListScrumStages,
	mcp.WithDescription("List stages of a Scrum version."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithNumber("scrum_version_id", mcp.Description("Version ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func GetScrumVersionHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return versionRequest(ctx, request, opts, "")
}

func ListScrumStagesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return versionRequest(ctx, request, opts, "/scrum_stages")
}

func versionRequest(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, suffix string) (*mcp.CallToolResult, error) {
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
	versionID, err := requiredInt(arguments, "scrum_version_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/programs/%d/scrum_versions/%d%s", enterpriseID, programID, versionID, suffix),
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
