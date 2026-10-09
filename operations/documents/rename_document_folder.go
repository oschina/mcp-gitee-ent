package documents

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const maxDocumentFolderNameRunes = 64

var RenameDocumentFolderTool = mcp.NewTool(
	RenameDocumentFolder,
	mcp.WithDescription("Rename a folder in an existing enterprise knowledge-base document. This endpoint handles folders only; to rename a page, pass page_name to update_enterprise_document_page instead."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(true),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("page_id", mcp.Description("Folder page_id returned by list_enterprise_document_pages."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("name", mcp.Description("New folder name. Must not contain '/', must differ from the current name, and is limited to 64 characters."), mcp.Required(), mcp.MinLength(1)),
)

func RenameDocumentFolderHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	name, err := requiredString(arguments, "name")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return utils.MCPErrorResult(utils.NewParamError("name", "parameter is required"))
	}
	if strings.Contains(name, "/") {
		return utils.MCPErrorResult(utils.NewParamError("name", "must not contain '/'"))
	}
	if utf8.RuneCountInString(name) > maxDocumentFolderNameRunes {
		return utils.MCPErrorResult(utils.NewParamError("name", fmt.Sprintf("must be at most %d characters", maxDocumentFolderNameRunes)))
	}

	path := fmt.Sprintf("/%d/docs/%d/files/%d/rename", enterpriseID, documentID, pageID)
	return documentChildResult(ctx, http.MethodPut, path, documentID, map[string]interface{}{"name": name}, opts...)
}
