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

var ListDocumentsTool = mcp.NewTool(
	ListDocuments,
	mcp.WithDescription("List knowledge-base documents accessible to the authenticated user in a Gitee enterprise. Use the returned document_id with document page, upload, download, and update tools. Each item also carries document_node_id, the document node ID to pass as doc_node_id to link_issue_document, unlink_issue_document, and get_document_node; document_id is a different identifier and those tools reject it. A listed document may still be read-only. This does not list repository Wikis or uploaded attachments. With search set, the tool returns page-level keyword hits instead of the flat document list; each hit carries document_node_id and page_id, and program_id, created_by_me, direction, and per_page are not supported."),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(true),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("program_id", mcp.Description("Program ID returned by list_programs. Omit to list enterprise-level documents."), mcp.Min(1)),
	mcp.WithBoolean("created_by_me", mcp.Description("Return only documents created by the authenticated user. Default: false."), mcp.DefaultBool(false)),
	mcp.WithString("search", mcp.Description("Keyword to search. Matches document names and page names; page content is searched only when the server has full-text search enabled. Mutually exclusive with program_id, created_by_me, direction, and per_page.")),
	mcp.WithString("sort", mcp.Description("Sort field: name, created_at, or updated_at. Default: updated_at."), mcp.Enum("name", "created_at", "updated_at"), mcp.DefaultString("updated_at")),
	mcp.WithString("direction", mcp.Description("Sort direction: asc or desc. Default: desc."), mcp.Enum("asc", "desc"), mcp.DefaultString("desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting from 1. Default: 1."), mcp.Min(1), mcp.DefaultNumber(1)),
	mcp.WithNumber("per_page", mcp.Description("Items per page. Default: 20, maximum: 100."), mcp.Min(1), mcp.Max(100), mcp.DefaultNumber(20)),
)

func ListDocumentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, result, err := getArguments(request)
	if err != nil {
		return result, err
	}

	enterpriseID, err := requiredPositiveInt(arguments, "enterprise_id")
	if err != nil {
		return utils.MCPErrorResult(err)
	}

	keyword, err := optionalString(arguments, "search", "")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		return searchDocuments(ctx, arguments, enterpriseID, keyword, opts...)
	}

	programID, err := optionalInt(arguments, "program_id", 0)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if _, present := arguments["program_id"]; present && programID <= 0 {
		return utils.MCPErrorResult(utils.NewParamError("program_id", "parameter must be a positive integer"))
	}
	createdByMe, err := optionalBool(arguments, "created_by_me", false)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	sortField, err := optionalString(arguments, "sort", "updated_at")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if sortField != "name" && sortField != "created_at" && sortField != "updated_at" {
		return utils.MCPErrorResult(utils.NewParamError("sort", "must be name, created_at, or updated_at"))
	}
	direction, err := optionalString(arguments, "direction", "desc")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if direction != "asc" && direction != "desc" {
		return utils.MCPErrorResult(utils.NewParamError("direction", "must be asc or desc"))
	}
	page, err := optionalInt(arguments, "page", 1)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	perPage, err := optionalInt(arguments, "per_page", 20)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if page < 1 {
		return utils.MCPErrorResult(utils.NewParamError("page", "must be at least 1"))
	}
	if perPage < 1 || perPage > 100 {
		return utils.MCPErrorResult(utils.NewParamError("per_page", "must be between 1 and 100"))
	}

	query := map[string]interface{}{
		"file_type":  "WikiInfo",
		"program_id": programID,
		"sort":       sortField,
		"direction":  direction,
		"page":       page,
		"per_page":   perPage,
	}
	if createdByMe {
		query["scope"] = "create_by_me"
	}

	raw := types.PagedResponse[types.DocumentNode]{}
	path := fmt.Sprintf("/%d/doc_nodes/tile", enterpriseID)
	if err := getJSON(ctx, http.MethodGet, path, &raw, append(opts, utils.WithQuery(query))...); err != nil {
		return utils.MCPErrorResult(err)
	}

	documents := make([]types.Document, 0, len(raw.Data))
	for _, node := range raw.Data {
		if node.FileType == "WikiInfo" {
			documents = append(documents, types.NewDocument(node))
		}
	}

	return utils.EnrichJSONResult(path, types.DocumentListResponse{
		TotalCount: raw.TotalCount,
		Documents:  documents,
	})
}

// searchDocuments calls GET /{enterprise_id}/doc_nodes/search, a page-level
// keyword search whose upstream contract pins order to desc and page size to a
// server constant, hence the rejected parameters below.
func searchDocuments(ctx context.Context, arguments map[string]interface{}, enterpriseID int, keyword string, opts ...utils.Option) (*mcp.CallToolResult, error) {
	if _, present := arguments["program_id"]; present {
		return utils.MCPErrorResult(utils.NewParamError("program_id", "not supported together with search"))
	}
	createdByMe, err := optionalBool(arguments, "created_by_me", false)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if createdByMe {
		return utils.MCPErrorResult(utils.NewParamError("created_by_me", "not supported together with search"))
	}
	if _, present := arguments["direction"]; present {
		return utils.MCPErrorResult(utils.NewParamError("direction", "not supported together with search; results are always sorted descending"))
	}
	if _, present := arguments["per_page"]; present {
		return utils.MCPErrorResult(utils.NewParamError("per_page", "not supported together with search; the server uses a fixed page size"))
	}
	sortField, err := optionalString(arguments, "sort", "updated_at")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if sortField != "name" && sortField != "created_at" && sortField != "updated_at" {
		return utils.MCPErrorResult(utils.NewParamError("sort", "must be name, created_at, or updated_at"))
	}
	page, err := optionalInt(arguments, "page", 1)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if page < 1 {
		return utils.MCPErrorResult(utils.NewParamError("page", "must be at least 1"))
	}

	raw := types.PagedResponse[types.DocumentSearchNode]{}
	path := fmt.Sprintf("/%d/doc_nodes/search", enterpriseID)
	query := map[string]interface{}{
		"search": keyword,
		"sort":   sortField,
		"page":   page,
	}
	if err := getJSON(ctx, http.MethodGet, path, &raw, append(opts, utils.WithQuery(query))...); err != nil {
		return utils.MCPErrorResult(err)
	}

	hits := make([]types.DocumentSearchHit, 0, len(raw.Data))
	for _, node := range raw.Data {
		hits = append(hits, types.NewDocumentSearchHit(node))
	}

	return utils.MCPJSONResult(types.DocumentSearchResponse{
		Search:     keyword,
		TotalCount: raw.TotalCount,
		Hits:       hits,
	})
}
