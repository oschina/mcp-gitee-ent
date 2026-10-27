package issue_links

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	UpdateIssueNote        = "update_issue_note"
	DeleteIssueNote        = "delete_issue_note"
	LinkIssuePullRequest   = "link_issue_pull_request"
	UnlinkIssuePullRequest = "unlink_issue_pull_request"
)

var UpdateIssueNoteTool = mcp.NewTool(UpdateIssueNote,
	mcp.WithDescription("Edit one issue comment. Not for editing the issue itself."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("note_id", mcp.Description("Comment ID."), mcp.Required()),
	mcp.WithString("body", mcp.Description("New comment content."), mcp.Required()),
)

var DeleteIssueNoteTool = mcp.NewTool(DeleteIssueNote,
	mcp.WithDescription("Delete one issue comment without deleting the issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("note_id", mcp.Description("Comment ID."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

var LinkIssuePullRequestTool = mcp.NewTool(LinkIssuePullRequest,
	mcp.WithDescription("Associate an existing pull request with an issue. This does not merge, close, or otherwise modify the pull request."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("pull_request_id", mcp.Description("Pull request ID expected by the link endpoint; do not pass the display iid."), mcp.Required()),
)

var UnlinkIssuePullRequestTool = mcp.NewTool(UnlinkIssuePullRequest,
	mcp.WithDescription("Remove the association between an issue and a pull request. This does not close or delete either resource."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("pull_request_id", mcp.Description("Pull request ID expected by the link endpoint; do not pass the display iid."), mcp.Required()),
)

func UpdateIssueNoteHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	noteID, err := requiredIssueArgument(arguments, "note_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	body, ok := arguments["body"].(string)
	if !ok || strings.TrimSpace(body) == "" {
		return mcp.NewToolResultError("Parameter body must be a non-empty string"), nil
	}
	opts = append(opts, utils.WithPayload(map[string]any{"body": strings.TrimSpace(body)}))
	opts = appendQueryOpt(opts, issueID)

	client := utils.NewGiteeClient(
		"PUT",
		fmt.Sprintf("/%d/issues/%s/notes/%s", enterpriseID, issueID, noteID),
		opts...,
	)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func DeleteIssueNoteHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	noteID, err := requiredIssueArgument(arguments, "note_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = appendQueryOpt(opts, issueID)

	client := utils.NewGiteeClient(
		"DELETE",
		fmt.Sprintf("/%d/issues/%s/notes/%s", enterpriseID, issueID, noteID),
		opts...,
	)
	return client.HandleMCPResultContext(ctx, nil)
}

func LinkIssuePullRequestHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return issuePullRequestLink(ctx, request, opts, "POST")
}

func UnlinkIssuePullRequestHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return issuePullRequestLink(ctx, request, opts, "DELETE")
}

func issuePullRequestLink(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, method string) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pullRequestID, err := requiredIssueArgument(arguments, "pull_request_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = appendQueryOpt(opts, issueID)

	client := utils.NewGiteeClient(
		method,
		fmt.Sprintf("/%d/issues/%s/link_pull_request/%s", enterpriseID, issueID, pullRequestID),
		opts...,
	)
	if method == "DELETE" {
		return client.HandleMCPResultContext(ctx, nil)
	}
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func requiredIssueArgument(arguments map[string]any, key string) (string, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return "", fmt.Errorf("Missing required parameter: %s", key)
	}
	text, err := utils.SafelyConvertToString(value)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("Parameter %s must not be empty", key)
	}
	return strings.TrimSpace(text), nil
}

func appendQueryOpt(opts []utils.Option, issueID string) []utils.Option {
	query := issueQuery(nil, issueID, nil)
	if len(query) == 0 {
		return opts
	}
	return append(opts, utils.WithQuery(query))
}
