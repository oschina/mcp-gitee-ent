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

const docNodeIDDescription = "Document node ID (doc_nodes.id). The same value appears as document_node_id in " +
	"list_enterprise_documents, doc_id in search_documents, and id in get_issue_development_links(type='docs'). " +
	"Never pass document_id, info_id, or page_id."

var LinkIssueDocumentTool = mcp.NewTool(LinkIssueDocument,
	mcp.WithDescription("Attach an existing knowledge-base document or attachment (a document node) to a work item. "+
		"doc_node_id is a document NODE id: use list_enterprise_documents' document_node_id, search_documents' doc_id, "+
		"or get_issue_development_links(type='docs').data[].id - these three fields carry the same value. "+
		"Passing document_id, info_id, or page_id fails with 404 'node not found'. "+
		"Only document (WikiInfo) and attachment (AttachFile) nodes can be linked; folders fail with 400. "+
		"Linking is not idempotent: linking an already-linked node fails with a generic 400, so read "+
		"get_issue_development_links(type='docs') before linking. This does not create or upload documents; "+
		"use create_enterprise_document or upload_enterprise_document_files first."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident, for example 9370067 or I5KTZN."), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description(docNodeIDDescription), mcp.Required(), mcp.Min(1)),
)

var UnlinkIssueDocumentTool = mcp.NewTool(UnlinkIssueDocument,
	mcp.WithDescription("Detach a knowledge-base document or attachment from a work item. "+
		"Pass the same document NODE id that was linked: read the current links with "+
		"get_issue_development_links(type='docs') and use each item's id. "+
		"Unlinking is idempotent: a node that is not linked still succeeds. "+
		"The response returns is_linked for the target node plus the issue's remaining linked_documents and "+
		"linked_documents_total, so the new state is confirmed without a follow-up read. "+
		"It only removes the association; the document itself is not deleted."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("issue_id", mcp.Description("Issue ID or issue ident, for example 9370067 or I5KTZN."), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description(docNodeIDDescription), mcp.Required(), mcp.Min(1)),
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
	client := utils.NewGiteeClient(
		method,
		fmt.Sprintf("/%d/issues/%s/link_doc_nodes/%d", enterpriseID, issueID, docNodeID),
		appendQueryOpt(opts, issueID)...,
	)
	if method != "DELETE" {
		var data any
		return client.HandleMCPResultContext(ctx, &data)
	}
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return unlinkIssueDocumentResult(ctx, enterpriseID, issueID, docNodeID, opts)
}

func unlinkIssueDocumentResult(ctx context.Context, enterpriseID int, issueID string, docNodeID int, opts []utils.Option) (*mcp.CallToolResult, error) {
	query := issueQuery(nil, issueID, nil)
	query["linked"] = true
	query["per_page"] = 100
	client := utils.NewGiteeClient(
		"GET",
		fmt.Sprintf("/%d/issues/%s/link_doc_nodes", enterpriseID, issueID),
		append(append([]utils.Option{}, opts...), utils.WithQuery(query))...,
	)
	var payload struct {
		TotalCount int   `json:"total_count"`
		Data       []any `json:"data"`
	}
	if result, err := client.HandleMCPResultContext(ctx, &payload); result != nil && result.IsError {
		return result, err
	}
	isLinked := false
	for _, item := range payload.Data {
		document, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if id, ok := document["id"].(float64); ok && int(id) == docNodeID {
			isLinked = true
			break
		}
	}
	return utils.MCPJSONResult(map[string]any{
		"doc_node_id":            docNodeID,
		"is_linked":              isLinked,
		"linked_documents":       payload.Data,
		"linked_documents_total": payload.TotalCount,
	})
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
