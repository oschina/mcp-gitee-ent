package check_runs

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListPRCheckRuns        = "list_pr_check_runs"
	ListCommitCheckRuns    = "list_commit_check_runs"
	ListProjectCheckRuns   = "list_project_check_runs"
	GetCheckRunAnnotations = "get_check_run_annotations"
)

var ListPRCheckRunsTool = mcp.NewTool(ListPRCheckRuns,
	mcp.WithDescription("List check runs for one or more pull requests in an enterprise."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("pull_request_ids", mcp.Description("Comma-separated pull request IDs.")),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListCommitCheckRunsTool = mcp.NewTool(ListCommitCheckRuns,
	mcp.WithDescription("List check runs for commits in one repository. Requires project_id."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("project_id", mcp.Description("Repository ID."), mcp.Required()),
	mcp.WithNumber("pull_request_id", mcp.Description("Optional pull request ID.")),
	mcp.WithString("commit_ids", mcp.Description("Comma-separated commit SHAs.")),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListProjectCheckRunsTool = mcp.NewTool(ListProjectCheckRuns,
	mcp.WithDescription("List gate/check runs for a repository, optionally filtered by pull request, check run, or head SHA."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("project_id", mcp.Description("Repository ID."), mcp.Required()),
	mcp.WithNumber("pull_request_id", mcp.Description("Optional pull request ID.")),
	mcp.WithNumber("check_run_id", mcp.Description("Optional check run ID.")),
	mcp.WithString("head_sha", mcp.Description("Optional commit SHA.")),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetCheckRunAnnotationsTool = mcp.NewTool(GetCheckRunAnnotations,
	mcp.WithDescription("Get annotations/details for a check run or commit head SHA."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("project_id", mcp.Description("Repository ID."), mcp.Required()),
	mcp.WithNumber("check_run_id", mcp.Description("Optional check run ID.")),
	mcp.WithString("head_sha", mcp.Description("Optional commit SHA.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListPRCheckRunsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	if value, ok := arguments["pull_request_ids"]; ok {
		query["pull_request_ids"] = value
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/pr_check_runs", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListCommitCheckRunsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	projectID, err := requiredInt(arguments, "project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{"project_id": projectID}
	for _, key := range []string{"pull_request_id", "commit_ids"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/commit_check_runs", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListProjectCheckRunsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, projectID, err := parseProject(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"pull_request_id", "check_run_id", "head_sha"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/projects/%d/check_runs/list", enterpriseID, projectID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetCheckRunAnnotationsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, projectID, err := parseProject(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload := map[string]any{}
	for _, key := range []string{"check_run_id", "head_sha", "page"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	opts = append(opts, utils.WithPayload(payload))
	client := utils.NewGiteeClient(
		"POST",
		fmt.Sprintf("/%d/projects/%d/check_runs/annotations", enterpriseID, projectID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseEnterprise(request mcp.CallToolRequest) (map[string]any, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, err
	}
	enterpriseID, err := requiredInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, err
	}
	return arguments, enterpriseID, nil
}

func parseProject(request mcp.CallToolRequest) (map[string]any, int, int, error) {
	arguments, enterpriseID, err := parseEnterprise(request)
	if err != nil {
		return nil, 0, 0, err
	}
	projectID, err := requiredInt(arguments, "project_id")
	if err != nil {
		return nil, 0, 0, err
	}
	return arguments, enterpriseID, projectID, nil
}

func requiredInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
