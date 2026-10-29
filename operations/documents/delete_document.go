package documents

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DeleteDocument     = "delete_enterprise_document"
	DeleteDocumentPage = "delete_enterprise_document_page"
)

var DeleteDocumentTool = mcp.NewTool(DeleteDocument,
	mcp.WithDescription("Delete an enterprise knowledge-base document (or folder) and everything inside it. "+
		"doc_node_id is the document node ID: document_node_id from list_enterprise_documents or id from "+
		"list_recent_documents/list_shared_documents, not document_id."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description("Document node ID (document_node_id), not document_id."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

var DeleteDocumentPageTool = mcp.NewTool(DeleteDocumentPage,
	mcp.WithDescription("Delete one page or folder inside a document. Use list_enterprise_document_pages to resolve page_id; "+
		"document_id comes from list_enterprise_documents."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("document_id", mcp.Description("Document ID (document_id) that owns the page."), mcp.Required()),
	mcp.WithNumber("page_id", mcp.Description("Page or folder ID inside the document."), mcp.Required()),
	mcp.WithDestructiveHintAnnotation(true),
)

func DeleteDocumentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, result, err := getArguments(request)
	if err != nil {
		return result, err
	}
	enterpriseID, err := requiredPositiveInt(arguments, "enterprise_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	docNodeID, err := requiredPositiveInt(arguments, "doc_node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient("DELETE", fmt.Sprintf("/%d/doc_nodes/%d", enterpriseID, docNodeID), opts...)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"doc_node_id": docNodeID, "deleted": true})
}

func DeleteDocumentPageHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, result, err := getArguments(request)
	if err != nil {
		return result, err
	}
	enterpriseID, err := requiredPositiveInt(arguments, "enterprise_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	documentID, err := requiredPositiveInt(arguments, "document_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	pageID, err := requiredPositiveInt(arguments, "page_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient(
		"DELETE",
		fmt.Sprintf("/%d/docs/%d/files/%d", enterpriseID, documentID, pageID),
		opts...,
	)
	if result, err := client.HandleMCPResultContext(ctx, nil); result != nil && result.IsError {
		return result, err
	}
	return utils.MCPActionResult(map[string]any{"document_id": documentID, "page_id": pageID, "deleted": true})
}
