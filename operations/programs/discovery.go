package programs

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetProgram           = "get_program"
	GetProgramOverview   = "get_program_overview"
	GetProgramComponents = "get_program_components"
)

var GetProgramTool = mcp.NewTool(GetProgram,
	mcp.WithDescription("Get one project (program) by ID, including its basic configuration and metadata."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetProgramOverviewTool = mcp.NewTool(GetProgramOverview,
	mcp.WithDescription("Get a project overview with current work item and progress statistics."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetProgramComponentsTool = mcp.NewTool(GetProgramComponents,
	mcp.WithDescription("List enabled and available components for a project, such as issues, milestones, sprints, and test management."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Project ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func GetProgramHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return getProgram(ctx, request, opts, "")
}

func GetProgramOverviewHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return getProgram(ctx, request, opts, "/overview")
}

func GetProgramComponentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return getProgram(ctx, request, opts, "/components")
}

func getProgram(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, suffix string) (*mcp.CallToolResult, error) {
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
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/programs/%d%s", enterpriseID, programID, suffix),
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
