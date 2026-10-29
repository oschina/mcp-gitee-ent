package issue_links

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	ListIssueAttachments    = "list_issue_attachments"
	ListIssueRelatedIssues  = "list_issue_related_issues"
	ListIssueAuthorizations = "list_issue_authorizations"
)

var ListIssueAttachmentsTool = mcp.NewTool(ListIssueAttachments,
	mcp.WithDescription("List files uploaded as attachments to an issue, with pagination. Returns issue attachment records, "+
		"not knowledge-base documents: the attachment id in this result is not a doc_node_id and cannot be passed to "+
		"link_issue_document. Use get_issue_development_links(type='docs') to list documents linked to the issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListIssueRelatedIssuesTool = mcp.NewTool(ListIssueRelatedIssues,
	mcp.WithDescription("List work items related to an issue, with optional dependency direction and filters."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithString("direction", mcp.Description("Dependency direction."), mcp.Enum("none", "pre", "latter")),
	mcp.WithString("assignee_ids", mcp.Description("Comma-separated assignee IDs.")),
	mcp.WithString("issue_state_ids", mcp.Description("Comma-separated issue state IDs.")),
	mcp.WithString("issue_type_ids", mcp.Description("Comma-separated issue type IDs.")),
	mcp.WithString("priority", mcp.Description("Comma-separated priorities.")),
	mcp.WithString("sort", mcp.Description("Sort field: program, created_at, updated_at, priority, issue_type, issue_state.")),
	mcp.WithString("sort_direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListIssueAuthorizationsTool = mcp.NewTool(ListIssueAuthorizations,
	mcp.WithDescription("Get the current user's permissions on an issue, such as read, update, destroy, and create comment."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

func ListIssueAttachmentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	// The attachment endpoints resolve issue_id by numeric ID only and ignore qt=ident,
	// so an issue ident must be resolved to its numeric ID first.
	numericIssueID, err := resolveNumericIssueID(ctx, enterpriseID, issueID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/paged_attach_files", enterpriseID, numericIssueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func resolveNumericIssueID(ctx context.Context, enterpriseID int, issueID string) (string, error) {
	if utils.IsAllDigits(issueID) {
		return issueID, nil
	}
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s", enterpriseID, issueID),
		utils.WithQuery(map[string]interface{}{"qt": "ident"}),
	)
	if _, err := client.DoContext(ctx); err != nil {
		return "", err
	}
	body, err := client.GetRespBody()
	if err != nil {
		return "", err
	}
	var detail struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(body, &detail); err != nil {
		return "", fmt.Errorf("could not resolve issue ident %q: %w", issueID, err)
	}
	if detail.ID <= 0 {
		return "", fmt.Errorf("could not resolve issue ident %q to a numeric issue ID", issueID)
	}
	return strconv.Itoa(detail.ID), nil
}

func ListIssueRelatedIssuesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(arguments, issueID, []string{
		"direction", "assignee_ids", "issue_state_ids", "issue_type_ids",
		"priority", "sort", "sort_direction", "page", "per_page",
	})
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/link_issues", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListIssueAuthorizationsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	_, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := issueQuery(nil, issueID, nil)
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/auths", enterpriseID, issueID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}
