package repository

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateEnterpriseProject = "create_enterprise_repository"
)

var CreateRepositoryTool = mcp.NewTool(CreateEnterpriseProject,
	mcp.WithDescription("Create a repository in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_name",
		mcp.Description("Repository name"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_namespace_path",
		mcp.Description("Namespace path"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_path",
		mcp.Description("Repository path"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_description",
		mcp.Description("Repository description"),
	),
	mcp.WithNumber(
		"project_public",
		mcp.Description("Whether public: 0: Private, 1: Public, 2: Internal Open"),
	),
	mcp.WithNumber(
		"project_outsourced",
		mcp.Description("Whether outsourced: 0: No, 1: Yes"),
	),
	mcp.WithString(
		"project_program_ids",
		mcp.Description("Program IDs (comma separated)"),
	),
	mcp.WithString(
		"project_member_ids",
		mcp.Description("Member IDs (comma separated)"),
	),
	mcp.WithNumber(
		"import_program_users",
		mcp.Description("Whether to import program members: 0: No, 1: Yes"),
	),
	mcp.WithNumber(
		"readme",
		mcp.Description("Whether to initialize readme: 0: No, 1: Yes"),
	),
	mcp.WithNumber(
		"issue_template",
		mcp.Description("Whether to initialize issue template: 0: No, 1: Yes"),
	),
	mcp.WithNumber(
		"pull_request_template",
		mcp.Description("Whether to initialize PR template: 0: No, 1: Yes"),
	),
)

func CreateRepositoryHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_name", "project_namespace_path", "project_path"); err != nil {
		return checkResult, err
	}
	// Validate required parameters
	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	apiUrl := fmt.Sprintf("/%d/projects", enterpriseID)
	payload := utils.ConvertToHash(request.Params.Arguments, "project", "name", "namespace_path", "path", "description", "public", "outsourced", "program_ids", "member_ids")
	opts = append(opts, utils.WithPayload(payload))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.Repository{}
	return giteeClient.HandleMCPResult(&data)
}

