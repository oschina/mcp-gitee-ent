package documents

import (
	"context"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

var GetDocumentPageTool = mcp.NewTool(
	GetDocumentPage,
	mcp.WithDescription("Get the complete content and metadata of one page in an enterprise knowledge-base document. Use list_enterprise_document_pages first if the page_id is unknown."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("page_id", mcp.Description("Page ID returned by list_enterprise_document_pages."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("version", mcp.Description("Historical version or commit ID. Omit to read the latest version.")),
	mcp.WithBoolean("include_rendered_html", mcp.Description("Include rendered HTML in addition to the source content. Default: false."), mcp.DefaultBool(false)),
)

func GetDocumentPageHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	version, err := optionalString(arguments, "version", "")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	includeHTML, err := optionalBool(arguments, "include_rendered_html", false)
	if err != nil {
		return utils.MCPErrorResult(err)
	}

	detail, err := fetchDocumentPage(ctx, enterpriseID, documentID, pageID, version, opts...)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if !includeHTML {
		detail.ContentHTML = ""
	}

	return utils.MCPJSONResult(types.DocumentPageResponse{DocumentID: documentID, Page: detail})
}
