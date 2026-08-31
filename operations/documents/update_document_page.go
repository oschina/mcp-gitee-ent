package documents

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

var UpdateDocumentPageTool = mcp.NewTool(
	UpdateDocumentPage,
	mcp.WithDescription("Replace the complete content of one page in an enterprise knowledge-base document. Read the page first and pass its current version; the tool re-reads the page and rejects content that is already stale before submission. The content parameter is a full replacement, not a patch."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(true),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("page_id", mcp.Description("Page ID returned by list_enterprise_document_pages."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("content", mcp.Description("Complete replacement page content in the content_type returned by get_enterprise_document_page. This is not a patch; an empty string clears the page."), mcp.Required()),
	mcp.WithString("expected_version", mcp.Description("Current page version returned by get_enterprise_document_page, used for a preflight stale-content check."), mcp.Required(), mcp.MinLength(1)),
	mcp.WithString("page_name", mcp.Description("New page name without a parent path. Omit to keep the current name.")),
	mcp.WithString("commit_message", mcp.Description("Short description of the change recorded in document history."), mcp.Required(), mcp.MinLength(1)),
	mcp.WithArray(
		"mentioned_user_ids",
		mcp.Description("Enterprise user IDs returned by list_enterprise_members to mention in the updated page."),
		mcp.Items(map[string]interface{}{"type": "number", "minimum": 1}),
	),
)

func UpdateDocumentPageHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, result, err := getArguments(request)
	if err != nil {
		return result, err
	}
	enterpriseID, err := requiredPositiveInt(arguments, "enterprise_id")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	documentID, err := requiredPositiveInt(arguments, "document_id")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	pageID, err := requiredPositiveInt(arguments, "page_id")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	content, err := requiredString(arguments, "content")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	expectedVersion, err := requiredString(arguments, "expected_version")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if strings.TrimSpace(expectedVersion) == "" {
		return utils.MCPErrorResult(utils.NewParamError("expected_version", "parameter must not be empty"))
	}
	commitMessage, err := requiredString(arguments, "commit_message")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if strings.TrimSpace(commitMessage) == "" {
		return utils.MCPErrorResult(utils.NewParamError("commit_message", "parameter must not be empty"))
	}
	pageName, err := optionalString(arguments, "page_name", "")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	mentionedUserIDs, err := intArray(arguments, "mentioned_user_ids")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	for _, userID := range mentionedUserIDs {
		if userID <= 0 {
			return utils.MCPErrorResult(utils.NewParamError("mentioned_user_ids", "all items must be positive integers"))
		}
	}

	current, err := fetchDocumentPage(ctx, enterpriseID, documentID, pageID, "", opts...)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if current.Version != expectedVersion {
		return utils.MCPErrorResult(utils.NewParamError(
			"expected_version",
			fmt.Sprintf("page version changed from %q to %q; read the page again before updating", expectedVersion, current.Version),
		))
	}

	payload := map[string]interface{}{
		"content": content,
		"message": commitMessage,
	}
	if pageName != "" {
		payload["name"] = pageName
	}
	if _, present := arguments["mentioned_user_ids"]; present {
		payload["mentioned_user_ids"] = pageIDList(mentionedUserIDs)
	}

	path := fmt.Sprintf("/%d/docs/%d/files/%d", enterpriseID, documentID, pageID)
	raw := types.DocumentPageDetailNode{}
	if err := getJSON(ctx, http.MethodPut, path, &raw, append(opts, utils.WithPayload(payload))...); err != nil {
		return utils.MCPErrorResult(err)
	}
	updated := types.NewDocumentPageDetail(raw)
	updated.ContentHTML = ""

	return utils.MCPJSONResult(types.DocumentPageResponse{DocumentID: documentID, Page: updated})
}
