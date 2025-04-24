package repository

import (
	"context"
	"fmt"
	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"net/url"
)

const (
	CreateRelease = "create_enterprise_repo_release"
)

var CreateReleaseTool = mcp.NewTool(CreateRelease,
	mcp.WithDescription("Create a release for repository in an enterprise"),
	mcp.WithNumber(
		"enterprise_id",
		mcp.Description("Enterprise ID"),
		mcp.Required(),
	),
	mcp.WithString(
		"project_id",
		mcp.Description("Project ID or PathWithNamespace"),
		mcp.Required(),
	),
	mcp.WithString(
		"release_tag_version",
		mcp.Description("Tag version"),
		mcp.Required(),
	),
	mcp.WithString(
		"release_title",
		mcp.Description("Release title"),
		mcp.Required(),
	),
	mcp.WithString(
		"release_ref",
		mcp.Description("Reference (branch or tag)"),
	),
	mcp.WithString(
		"release_description",
		mcp.Description("Release description"),
		mcp.Required(),
	),
	mcp.WithString(
		"release_release_type",
		mcp.Description("Release type, 0: release, 1: pre-release"),
		mcp.Enum("0", "1"),
		mcp.DefaultString("0"),
	),
)

func CreateReleaseHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	if checkResult, err := utils.CheckRequired(request.Params.Arguments, "project_id", "release_tag_version", "release_title", "release_description"); err != nil {
		return checkResult, err
	}

	enterpriseIDArg := request.Params.Arguments["enterprise_id"]
	enterpriseID, err := utils.SafelyConvertToInt(enterpriseIDArg)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}
	projectIDArg := request.Params.Arguments["project_id"]
	if !utils.IsAllDigits(projectIDArg.(string)) {
		request.Params.Arguments["qt"] = "path"
	}

	apiUrl := fmt.Sprintf("/%d/projects/%s/releases", enterpriseID, url.QueryEscape(projectIDArg.(string)))

	payload := utils.ConvertToHash(request.Params.Arguments, "release", "tag_version", "title", "ref", "description", "release_type")
	opts = append(opts, utils.WithPayload(payload))
	giteeClient := utils.NewGiteeClient("POST", apiUrl, opts...)

	data := types.ReleaseDetail{}
	return giteeClient.HandleMCPResult(&data)
}

