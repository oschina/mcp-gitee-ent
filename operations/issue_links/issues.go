package issue_links

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GetIssueOverview   = "get_issue_overview"
	ListIssueChildren  = "list_issue_children"
	ListIssueAncestors = "list_issue_ancestors"
)

var GetIssueOverviewTool = mcp.NewTool(GetIssueOverview,
	mcp.WithDescription("Get an issue overview for planning and navigation. "+
		"Use get_enterprise_issue_detail for full issue fields; do not use this for development links."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListIssueChildrenTool = mcp.NewTool(ListIssueChildren,
	mcp.WithDescription("List child work items of an issue. Not for ancestors or other linked issues."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithString("issue_type",
		mcp.Description("Optional child issue type filter."),
		mcp.Enum("all", "feature", "task", "bug"),
	),
	mcp.WithString("assignee_ids", mcp.Description("Comma-separated assignee IDs.")),
	mcp.WithString("issue_state_ids", mcp.Description("Comma-separated issue state IDs.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListIssueAncestorsTool = mcp.NewTool(ListIssueAncestors,
	mcp.WithDescription("List ancestor work items of an issue to trace its parent hierarchy. Not for children or development links."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithString("scope",
		mcp.Description("Ancestor range: part returns the root and parent; all returns the complete chain."),
		mcp.Enum("part", "all"),
	),
	mcp.WithReadOnlyHintAnnotation(true),
)

func GetIssueOverviewHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(arguments, issueID, nil)
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/overview", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListIssueChildrenHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(arguments, issueID,
		[]string{"issue_type", "assignee_ids", "issue_state_ids", "page", "per_page"})
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/children", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListIssueAncestorsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(arguments, issueID, []string{"scope"})
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/ancestors", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseIssueArguments(request mcp.CallToolRequest) (map[string]any, int, string, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, "", err
	}
	for _, key := range []string{"enterprise_id", "issue_id"} {
		if value, ok := arguments[key]; !ok || value == nil {
			return nil, 0, "", fmt.Errorf("Missing required parameter: %s", key)
		}
	}
	enterpriseID, err := utils.SafelyConvertToInt(arguments["enterprise_id"])
	if err != nil {
		return nil, 0, "", err
	}
	issueID, err := utils.SafelyConvertToString(arguments["issue_id"])
	if err != nil {
		return nil, 0, "", err
	}
	if issueID == "" {
		return nil, 0, "", fmt.Errorf("Parameter issue_id must not be empty")
	}
	return arguments, enterpriseID, issueID, nil
}

func issueQuery(arguments map[string]any, issueID string, allowed []string) map[string]interface{} {
	query := make(map[string]interface{}, len(allowed)+1)
	if !utils.IsAllDigits(issueID) {
		query["qt"] = "ident"
	}
	for _, key := range allowed {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	return query
}

const (
	GetIssueDevelopmentLinks = "get_issue_development_links"
	ListIssueOperationLogs   = "list_issue_operation_logs"
	ListIssueNextStates      = "list_issue_next_states"
)

var GetIssueDevelopmentLinksTool = mcp.NewTool(GetIssueDevelopmentLinks,
	mcp.WithDescription("Get one category of development links for an issue: related PRs, branches, commits, tests, docs, or linked work items. "+
		"Not for PR diffs, file contents, or issue hierarchy."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithString("type",
		mcp.Description("Link category to return."),
		mcp.Required(),
		mcp.Enum("overview", "issues", "parent", "children", "pull_requests", "branch", "commits", "test_cases", "docs"),
	),
	mcp.WithNumber("issue_state_id", mcp.Description("Filter related or child work items by issue state ID.")),
	mcp.WithString("state", mcp.Description("Filter linked pull requests by state.")),
	mcp.WithNumber("results", mcp.Description("Filter linked test cases by execution result: 0 untested, 1 passed, 2 blocked, 3 failed, 4 skipped."), mcp.Min(0), mcp.Max(4)),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListIssueOperationLogsTool = mcp.NewTool(ListIssueOperationLogs,
	mcp.WithDescription("List operation history for an issue. Not for comments; use list_enterprise_issue_comments for discussions."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("prev_id", mcp.Description("Pagination cursor: previous log ID.")),
	mcp.WithNumber("last_id", mcp.Description("Pagination cursor: next log ID.")),
	mcp.WithNumber("limit", mcp.Description("Maximum number of log entries to return."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListIssueNextStatesTool = mcp.NewTool(ListIssueNextStates,
	mcp.WithDescription("List states to which this issue can currently transition. Use before changing its state; not for all states configured for an issue type."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

type developmentLink struct {
	path    string
	allowed []string
}

var developmentLinks = map[string]developmentLink{
	"overview":      {path: "overview"},
	"issues":        {path: "issues", allowed: []string{"page", "per_page", "issue_state_id"}},
	"parent":        {path: "parent"},
	"children":      {path: "children", allowed: []string{"page", "per_page", "issue_state_id"}},
	"pull_requests": {path: "pull_requests", allowed: []string{"page", "per_page", "state"}},
	"branch":        {path: "branch"},
	"commits":       {path: "commits", allowed: []string{"page", "per_page"}},
	"test_cases":    {path: "test_cases", allowed: []string{"page", "per_page", "results"}},
	"docs":          {path: "docs", allowed: []string{"page", "per_page"}},
}

func GetIssueDevelopmentLinksHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	typeValue, ok := arguments["type"].(string)
	if !ok || strings.TrimSpace(typeValue) == "" {
		return mcp.NewToolResultError("Missing required parameter: type. Allowed values: overview, issues, parent, children, pull_requests, branch, commits, test_cases, docs."), nil
	}
	link, ok := developmentLinks[typeValue]
	if !ok {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid type %q. Allowed values: overview, issues, parent, children, pull_requests, branch, commits, test_cases, docs.", typeValue)), nil
	}
	query := issueQuery(arguments, issueID, link.allowed)
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/data_links/%s", enterpriseID, issueID, link.path),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListIssueOperationLogsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(arguments, issueID, []string{"prev_id", "last_id", "limit"})
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/operate_logs", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListIssueNextStatesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	_, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(nil, issueID, nil)
	opts = append(opts, utils.WithQuery(query))

	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/issue_states", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}
