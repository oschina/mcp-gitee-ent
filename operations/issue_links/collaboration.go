package issue_links

import (
	"context"
	"fmt"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SetIssueStarred      = "set_issue_starred"
	AddIssueReaction     = "add_issue_reaction"
	RemoveIssueReaction  = "remove_issue_reaction"
	AddIssueRelation     = "add_issue_relation"
	UpdateIssueRelation  = "update_issue_relation"
	RemoveIssueRelation  = "remove_issue_relation"
	LinkIssueDocument    = "link_issue_document"
	UnlinkIssueDocument  = "unlink_issue_document"
	LinkIssueTestCases   = "link_issue_test_cases"
	UnlinkIssueTestCases = "unlink_issue_test_cases"
)

var SetIssueStarredTool = mcp.NewTool(SetIssueStarred,
	mcp.WithDescription("Star or unstar an issue for the current user."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithBoolean("starred", mcp.Description("True to star, false to unstar."), mcp.Required()),
)

var AddIssueReactionTool = mcp.NewTool(AddIssueReaction,
	mcp.WithDescription("Add an emoji reaction to an issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithString("text", mcp.Description("Reaction text."), mcp.Required(), mcp.Enum("+1", "-1", "smile", "tada", "confused", "heart", "rocket", "eyes")),
)

var RemoveIssueReactionTool = mcp.NewTool(RemoveIssueReaction,
	mcp.WithDescription("Remove one issue reaction by reaction ID."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("reaction_id", mcp.Description("Reaction ID."), mcp.Required()),
)

var AddIssueRelationTool = mcp.NewTool(AddIssueRelation,
	mcp.WithDescription("Create a relation between two issues."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("link_issue_id", mcp.Description("Issue ID to link."), mcp.Required()),
	mcp.WithString("ref_type", mcp.Description("Relation type."), mcp.Required(), mcp.Enum("normal", "finish_to_finish", "finish_to_start", "start_to_start", "start_to_finish")),
	mcp.WithString("direction", mcp.Description("Relation direction."), mcp.Enum("none", "pre", "latter")),
)

var UpdateIssueRelationTool = mcp.NewTool(UpdateIssueRelation,
	mcp.WithDescription("Update an existing issue relation."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("link_issue_id", mcp.Description("Linked issue ID."), mcp.Required()),
	mcp.WithString("ref_type", mcp.Description("Relation type."), mcp.Enum("normal", "finish_to_finish", "finish_to_start", "start_to_start", "start_to_finish")),
	mcp.WithString("direction", mcp.Description("Relation direction."), mcp.Enum("none", "pre", "latter")),
)

var RemoveIssueRelationTool = mcp.NewTool(RemoveIssueRelation,
	mcp.WithDescription("Remove a relation between two issues."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("link_issue_id", mcp.Description("Linked issue ID."), mcp.Required()),
)

var LinkIssueDocumentTool = mcp.NewTool(LinkIssueDocument,
	mcp.WithDescription("Link a document node to an issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description("Document node ID."), mcp.Required()),
)

var UnlinkIssueDocumentTool = mcp.NewTool(UnlinkIssueDocument,
	mcp.WithDescription("Unlink a document node from an issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description("Document node ID."), mcp.Required()),
)

var LinkIssueTestCasesTool = mcp.NewTool(LinkIssueTestCases,
	mcp.WithDescription("Link test plan cases to an issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithArray("test_plan_case_ids", mcp.Description("Test plan case IDs."), mcp.Required(), mcp.Items(map[string]any{"type": "number"})),
)

var UnlinkIssueTestCasesTool = mcp.NewTool(UnlinkIssueTestCases,
	mcp.WithDescription("Unlink test plan cases from an issue."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident."), mcp.Required()),
	mcp.WithArray("test_plan_case_ids", mcp.Description("Test plan case IDs."), mcp.Required(), mcp.Items(map[string]any{"type": "number"})),
)

func SetIssueStarredHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	starred, ok := arguments["starred"].(bool)
	if !ok {
		return mcp.NewToolResultError("Missing required parameter: starred"), nil
	}
	method := "POST"
	if !starred {
		method = "DELETE"
	}
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient(method, fmt.Sprintf("/%d/issues/%s/star", enterpriseID, issueID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func AddIssueReactionHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	text, ok := arguments["text"].(string)
	if !ok || strings.TrimSpace(text) == "" {
		return mcp.NewToolResultError("Missing required parameter: text"), nil
	}
	opts = append(opts, utils.WithPayload(map[string]any{"text": text}))
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/issues/%s/reaction", enterpriseID, issueID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func RemoveIssueReactionHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	enterpriseID, err := requiredIssueArgumentInt(arguments, "enterprise_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	reactionID, err := requiredIssueArgumentInt(arguments, "reaction_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/reactions/%d", enterpriseID, reactionID), opts...)
	return client.HandleMCPResultContext(ctx, nil)
}

func AddIssueRelationHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload, err := relationPayload(arguments, true)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithPayload(payload))
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient("POST", fmt.Sprintf("/%d/issues/%s/related_issue", enterpriseID, issueID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func UpdateIssueRelationHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkIssueID, err := requiredIssueArgumentInt(arguments, "link_issue_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	payload, err := relationPayload(arguments, false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = append(opts, utils.WithPayload(payload))
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient("PUT", fmt.Sprintf("/%d/issues/%s/related_issue/%d", enterpriseID, issueID, linkIssueID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func RemoveIssueRelationHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	linkIssueID, err := requiredIssueArgumentInt(arguments, "link_issue_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/issues/%s/related_issue/%d", enterpriseID, issueID, linkIssueID), opts...)
	return client.HandleMCPResultContext(ctx, nil)
}

func LinkIssueDocumentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changeIssueDocument(ctx, request, opts, "POST")
}

func UnlinkIssueDocumentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changeIssueDocument(ctx, request, opts, "DELETE")
}

func changeIssueDocument(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, method string) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	docNodeID, err := requiredIssueArgumentInt(arguments, "doc_node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient(method, fmt.Sprintf("/%d/issues/%s/link_doc_nodes/%d", enterpriseID, issueID, docNodeID), opts...)
	if method == "DELETE" {
		return client.HandleMCPResultContext(ctx, nil)
	}
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func LinkIssueTestCasesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changeIssueTestCases(ctx, request, opts, "POST")
}

func UnlinkIssueTestCasesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	return changeIssueTestCases(ctx, request, opts, "DELETE")
}

func changeIssueTestCases(ctx context.Context, request mcp.CallToolRequest, opts []utils.Option, method string) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, issueID, err := parseIssueArguments(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	caseIDs, err := intArrayArgument(arguments["test_plan_case_ids"])
	if err != nil || len(caseIDs) == 0 {
		return mcp.NewToolResultError("test_plan_case_ids must contain at least one ID"), nil
	}
	opts = append(opts, utils.WithPayload(map[string]any{"test_plan_case_ids": caseIDs}))
	opts = appendQueryOpt(opts, issueID)
	client := utils.NewGiteeClient(method, fmt.Sprintf("/%d/issues/%s/test_plan_cases", enterpriseID, issueID), opts...)
	return client.HandleMCPResultContext(ctx, nil)
}

func relationPayload(arguments map[string]any, requireRefType bool) (map[string]any, error) {
	payload := map[string]any{}
	linkIssueID, err := requiredIssueArgumentInt(arguments, "link_issue_id")
	if err != nil {
		return nil, err
	}
	payload["link_issue_id"] = linkIssueID
	for _, key := range []string{"ref_type", "direction"} {
		if value, ok := arguments[key]; ok && value != nil {
			payload[key] = value
		}
	}
	if requireRefType {
		if _, ok := payload["ref_type"]; !ok {
			return nil, fmt.Errorf("Missing required parameter: ref_type")
		}
	}
	if !requireRefType && len(payload) == 1 {
		return nil, fmt.Errorf("At least one of ref_type or direction is required")
	}
	return payload, nil
}

func requiredIssueArgumentInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}

func intArrayArgument(value any) ([]int, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array of numeric IDs")
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		id, err := utils.SafelyConvertToInt(item)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}
