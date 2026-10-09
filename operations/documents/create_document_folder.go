package documents

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateDocumentFolderTool = mcp.NewTool(
	CreateDocumentFolder,
	mcp.WithDescription("Create a new folder inside an existing enterprise knowledge-base document. Omit parent_id to create the folder at the document root."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("name", mcp.Description("Folder name."), mcp.Required(), mcp.MinLength(1)),
	mcp.WithNumber("parent_id", mcp.Description("Parent folder's page_id returned by list_enterprise_document_pages. Omit to create at the document root."), mcp.Min(1)),
)

func CreateDocumentFolderHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	name, err := requiredString(arguments, "name")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if name = strings.TrimSpace(name); name == "" {
		return utils.MCPErrorResult(utils.NewParamError("name", "parameter is required"))
	}
	parentID, err := optionalInt(arguments, "parent_id", 0)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if _, present := arguments["parent_id"]; present && parentID <= 0 {
		return utils.MCPErrorResult(utils.NewParamError("parent_id", "parameter must be a positive integer"))
	}

	payload := map[string]interface{}{"name": name}
	if parentID > 0 {
		payload["parent_id"] = parentID
	}

	path := fmt.Sprintf("/%d/docs/%d/create_folder", enterpriseID, documentID)
	return documentChildResult(ctx, http.MethodPost, path, documentID, payload, opts...)
}
