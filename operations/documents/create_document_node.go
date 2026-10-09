package documents

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

var CreateDocumentNodeTool = mcp.NewTool(
	CreateDocumentNode,
	mcp.WithDescription("Create a page or folder inside an existing enterprise knowledge-base document. Set type to \"page\" or \"folder\". Omit parent_id to create it at the document root. A new page starts empty; read it with get_enterprise_document_page to obtain its version, then add content with update_enterprise_document_page."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("type", mcp.Description("What to create: page for a page, folder for a folder."), mcp.Required(), mcp.Enum("page", "folder")),
	mcp.WithString("name", mcp.Description("Page or folder name."), mcp.Required(), mcp.MinLength(1)),
	mcp.WithNumber("parent_id", mcp.Description("Parent folder's page_id returned by list_enterprise_document_pages. Omit to create at the document root."), mcp.Min(1)),
)

func CreateDocumentNodeHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	nodeType, err := requiredString(arguments, "type")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if nodeType != "page" && nodeType != "folder" {
		return utils.MCPErrorResult(utils.NewParamError("type", "must be page or folder"))
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

	action := "create_folder"
	if nodeType == "page" {
		action = "create_file"
	}
	path := fmt.Sprintf("/%d/docs/%d/%s", enterpriseID, documentID, action)
	return documentChildResult(ctx, http.MethodPost, path, documentID, payload, opts...)
}
