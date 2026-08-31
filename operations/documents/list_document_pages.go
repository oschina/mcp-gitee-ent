package documents

import (
	"context"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

var ListDocumentPagesTool = mcp.NewTool(
	ListDocumentPages,
	mcp.WithDescription("List all pages and folders in an enterprise knowledge-base document. Call this before reading or updating a page to resolve its page_id and current version."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents. Do not pass document_node_id or page_id."), mcp.Required(), mcp.Min(1)),
)

func ListDocumentPagesHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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

	response, err := fetchDocumentPages(ctx, enterpriseID, documentID, opts...)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	return utils.MCPJSONResult(response)
}
