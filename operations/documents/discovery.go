package documents

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	SearchDocuments      = "search_documents"
	ListRecentDocuments  = "list_recent_documents"
	ListSharedDocuments  = "list_shared_documents"
	GetDocumentNode      = "get_document_node"
	ListDocumentVersions = "list_document_versions"
)

var SearchDocumentsTool = mcp.NewTool(SearchDocuments,
	mcp.WithDescription("Search documents and knowledge-base nodes in an enterprise. Supports optional project or parent-folder filtering."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithString("search", mcp.Description("Search keyword."), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Optional project ID.")),
	mcp.WithNumber("parent_id", mcp.Description("Optional parent doc node ID.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListRecentDocumentsTool = mcp.NewTool(ListRecentDocuments,
	mcp.WithDescription("List recently edited documents, optionally scoped to a project."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Optional project ID.")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListSharedDocumentsTool = mcp.NewTool(ListSharedDocuments,
	mcp.WithDescription("List documents shared with the current user."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("program_id", mcp.Description("Optional project ID.")),
	mcp.WithString("search", mcp.Description("Search keyword.")),
	mcp.WithString("file_type", mcp.Description("Comma-separated file types: WikiInfo, Wiki, DocDirectory, AttachFile.")),
	mcp.WithString("sort", mcp.Description("Sort field: name, size, updated_at, created_at, creator_id, public.")),
	mcp.WithString("direction", mcp.Description("Sort direction."), mcp.Enum("asc", "desc")),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

var GetDocumentNodeTool = mcp.NewTool(GetDocumentNode,
	mcp.WithDescription("Get one document or knowledge-base node by ID."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description("Document node ID."), mcp.Required()),
	mcp.WithReadOnlyHintAnnotation(true),
)

var ListDocumentVersionsTool = mcp.NewTool(ListDocumentVersions,
	mcp.WithDescription("List historical versions of an attachment document node."),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID"), mcp.Required()),
	mcp.WithNumber("doc_node_id", mcp.Description("Document node ID."), mcp.Required()),
	mcp.WithNumber("page", mcp.Description("Page number, starting at 1."), mcp.Min(1)),
	mcp.WithNumber("per_page", mcp.Description("Page size, max 100."), mcp.Min(1), mcp.Max(100)),
	mcp.WithReadOnlyHintAnnotation(true),
)

func SearchDocumentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseDocumentEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	search, ok := arguments["search"].(string)
	if !ok || search == "" {
		return mcp.NewToolResultError("Missing required parameter: search"), nil
	}
	query := map[string]interface{}{"search": search}
	for _, key := range []string{"program_id", "parent_id", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/doc_nodes/search/v2", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListRecentDocumentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseDocumentEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"program_id", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/doc_nodes/recent_doc", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListSharedDocumentsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseDocumentEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"program_id", "search", "file_type", "sort", "direction", "page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/doc_nodes/shared", enterpriseID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func GetDocumentNodeHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseDocumentEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	nodeID, err := requiredDocumentInt(arguments, "doc_node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/doc_nodes/%d", enterpriseID, nodeID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func ListDocumentVersionsHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
	arguments, enterpriseID, err := parseDocumentEnterprise(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	nodeID, err := requiredDocumentInt(arguments, "doc_node_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	query := map[string]interface{}{}
	for _, key := range []string{"page", "per_page"} {
		if value, ok := arguments[key]; ok {
			query[key] = value
		}
	}
	opts = append(opts, utils.WithQuery(query))
	client := utils.NewGiteeClient("GET", fmt.Sprintf("/%d/doc_nodes/%d/file_versions", enterpriseID, nodeID), opts...)
	var data any
	return client.HandleMCPResultContext(ctx, &data)
}

func parseDocumentEnterprise(request mcp.CallToolRequest) (map[string]any, int, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, 0, err
	}
	enterpriseID, err := requiredDocumentInt(arguments, "enterprise_id")
	if err != nil {
		return nil, 0, err
	}
	return arguments, enterpriseID, nil
}

func requiredDocumentInt(arguments map[string]any, key string) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("Missing required parameter: %s", key)
	}
	return utils.SafelyConvertToInt(value)
}
