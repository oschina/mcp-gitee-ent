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

var CreateDocumentTool = mcp.NewTool(
	CreateDocument,
	mcp.WithDescription("Create a new knowledge-base document in a Gitee enterprise. Returns the created document_id, which you can pass to the other document tools. This creates an empty document; add content with upload_enterprise_document_files or update_enterprise_document_page. Creation may initialize the document repository asynchronously."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("name", mcp.Description("Document name."), mcp.Required(), mcp.MinLength(1)),
	mcp.WithNumber("parent_id", mcp.Description("Parent folder's doc_node id. Omit to create at the enterprise knowledge-base root."), mcp.Min(1)),
	mcp.WithNumber("program_id", mcp.Description("Program ID returned by list_programs, to create the document under a program knowledge base. Omit for an enterprise-level document."), mcp.Min(1)),
	mcp.WithString("auth_type", mcp.Description("Access scope: inherit, private, read_only, or read_write. Default: inherit."), mcp.Enum("inherit", "private", "read_only", "read_write"), mcp.DefaultString("inherit")),
	mcp.WithNumber("scrum_sprint_id", mcp.Description("Scrum sprint ID. When set, the document is created under the sprint's folder and program_id is required."), mcp.Min(1)),
)

func CreateDocumentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, result, err := getArguments(request)
	if err != nil {
		return result, err
	}

	enterpriseID, err := requiredPositiveInt(arguments, "enterprise_id")
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
	programID, err := optionalInt(arguments, "program_id", 0)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if _, present := arguments["program_id"]; present && programID <= 0 {
		return utils.MCPErrorResult(utils.NewParamError("program_id", "parameter must be a positive integer"))
	}
	authType, err := optionalString(arguments, "auth_type", "inherit")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	switch authType {
	case "inherit", "private", "read_only", "read_write":
	default:
		return utils.MCPErrorResult(utils.NewParamError("auth_type", "must be inherit, private, read_only, or read_write"))
	}
	sprintID, err := optionalInt(arguments, "scrum_sprint_id", 0)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if _, present := arguments["scrum_sprint_id"]; present && sprintID <= 0 {
		return utils.MCPErrorResult(utils.NewParamError("scrum_sprint_id", "parameter must be a positive integer"))
	}
	if sprintID > 0 && programID <= 0 {
		return utils.MCPErrorResult(utils.NewParamError("scrum_sprint_id", "requires program_id"))
	}

	payload := map[string]interface{}{"name": name}
	if parentID > 0 {
		payload["parent_id"] = parentID
	}
	if programID > 0 {
		payload["program_id"] = programID
	}
	if authType != "" {
		payload["auth_type"] = authType
	}
	if sprintID > 0 {
		payload["scrum_sprint_id"] = sprintID
	}

	raw := types.DocumentNode{}
	path := fmt.Sprintf("/%d/doc_nodes/create", enterpriseID)
	if err := getJSON(ctx, http.MethodPost, path, &raw, append(opts, utils.WithPayload(payload))...); err != nil {
		return utils.MCPErrorResult(err)
	}

	return utils.MCPJSONResult(types.NewDocument(raw))
}
