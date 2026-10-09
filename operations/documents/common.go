package documents

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	CreateDocument       = "create_enterprise_document"
	CreateDocumentNode   = "create_enterprise_document_node"
	RenameDocumentFolder = "rename_enterprise_document_folder"
	ListDocuments        = "list_enterprise_documents"
	ListDocumentPages    = "list_enterprise_document_pages"
	GetDocumentPage      = "get_enterprise_document_page"
	UploadDocument       = "upload_enterprise_document_files"
	DownloadDocument     = "download_enterprise_document"
	UpdateDocumentPage   = "update_enterprise_document_page"
)

func getArguments(request mcp.CallToolRequest) (map[string]interface{}, *mcp.CallToolResult, error) {
	arguments, err := utils.ConvertArgumentsToMap(request.Params.Arguments)
	if err != nil {
		return nil, mcp.NewToolResultError(err.Error()), err
	}
	return arguments, nil, nil
}

func requiredInt(arguments map[string]interface{}, key string) (int, error) {
	value, ok := arguments[key]
	if !ok {
		return 0, utils.NewParamError(key, "parameter is required")
	}
	return utils.SafelyConvertToInt(value)
}

func requiredPositiveInt(arguments map[string]interface{}, key string) (int, error) {
	value, err := requiredInt(arguments, key)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, utils.NewParamError(key, "parameter must be a positive integer")
	}
	return value, nil
}

func optionalInt(arguments map[string]interface{}, key string, fallback int) (int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return fallback, nil
	}
	return utils.SafelyConvertToInt(value)
}

func optionalString(arguments map[string]interface{}, key, fallback string) (string, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return fallback, nil
	}
	return utils.SafelyConvertToString(value)
}

func requiredString(arguments map[string]interface{}, key string) (string, error) {
	value, ok := arguments[key]
	if !ok {
		return "", utils.NewParamError(key, "parameter is required")
	}
	return utils.SafelyConvertToString(value)
}

func optionalBool(arguments map[string]interface{}, key string, fallback bool) (bool, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return fallback, nil
	}
	parsed, ok := value.(bool)
	if !ok {
		return false, utils.NewParamError(key, "parameter must be a boolean")
	}
	return parsed, nil
}

func stringArray(arguments map[string]interface{}, key string) ([]string, error) {
	value, ok := arguments[key]
	if !ok {
		return nil, utils.NewParamError(key, "parameter is required")
	}

	switch values := value.(type) {
	case []string:
		return values, nil
	case []interface{}:
		result := make([]string, 0, len(values))
		for _, item := range values {
			value, err := utils.SafelyConvertToString(item)
			if err != nil {
				return nil, utils.NewParamError(key, "all items must be strings")
			}
			result = append(result, value)
		}
		return result, nil
	default:
		return nil, utils.NewParamError(key, "parameter must be an array of strings")
	}
}

func intArray(arguments map[string]interface{}, key string) ([]int, error) {
	value, ok := arguments[key]
	if !ok || value == nil {
		return nil, nil
	}

	values, ok := value.([]interface{})
	if !ok {
		if ints, ok := value.([]int); ok {
			return ints, nil
		}
		return nil, utils.NewParamError(key, "parameter must be an array of numbers")
	}

	result := make([]int, 0, len(values))
	for _, item := range values {
		parsed, err := utils.SafelyConvertToInt(item)
		if err != nil {
			return nil, utils.NewParamError(key, "all items must be numbers")
		}
		result = append(result, parsed)
	}
	return result, nil
}

func getJSON(ctx context.Context, method, path string, target interface{}, opts ...utils.Option) error {
	client := utils.NewGiteeClient(method, path, opts...)
	_, err := client.DoContext(ctx)
	if err != nil {
		return err
	}
	defer client.Response.Body.Close()

	body, err := io.ReadAll(client.Response.Body)
	if err != nil {
		return utils.NewInternalError(fmt.Errorf("read response body: %w", err))
	}
	if err := json.Unmarshal(body, target); err != nil {
		return utils.NewInternalError(fmt.Errorf("parse response body: %w", err))
	}
	return nil
}

func fetchDocumentPages(ctx context.Context, enterpriseID, documentID int, opts ...utils.Option) (types.DocumentPagesResponse, error) {
	path := fmt.Sprintf("/%d/docs/%d/files", enterpriseID, documentID)
	raw := types.PagedResponse[types.DocumentPageNode]{}
	if err := getJSON(ctx, http.MethodGet, path, &raw, opts...); err != nil {
		return types.DocumentPagesResponse{}, err
	}

	pages := make([]types.DocumentPage, 0)
	for _, node := range raw.Data {
		flattenDocumentPage(node, "", &pages)
	}

	return types.DocumentPagesResponse{
		DocumentID: documentID,
		TotalCount: len(pages),
		Pages:      pages,
	}, nil
}

func flattenDocumentPage(node types.DocumentPageNode, parentPath string, pages *[]types.DocumentPage) {
	pagePath := node.Name
	if parentPath != "" {
		pagePath = parentPath + "/" + node.Name
	}

	pageType := "page"
	if node.FileType == 2 {
		pageType = "folder"
	}

	*pages = append(*pages, types.DocumentPage{
		PageID:       node.ID,
		ParentPageID: node.ParentID,
		Path:         pagePath,
		Name:         node.Name,
		Type:         pageType,
		ContentType:  node.ContentType,
		Version:      node.Version,
		LastEditTime: node.LastEditTime,
		CreatedAt:    node.CreatedAt,
		UpdatedAt:    node.UpdatedAt,
	})

	for _, child := range node.Children {
		flattenDocumentPage(child, pagePath, pages)
	}
}

func fetchDocumentPage(ctx context.Context, enterpriseID, documentID, pageID int, version string, opts ...utils.Option) (types.DocumentPageDetail, error) {
	path := fmt.Sprintf("/%d/docs/%d/files/%d", enterpriseID, documentID, pageID)
	query := map[string]interface{}{}
	if version != "" {
		query["version_id"] = version
	}
	if len(query) > 0 {
		opts = append(opts, utils.WithQuery(query))
	}

	raw := types.DocumentPageDetailNode{}
	if err := getJSON(ctx, http.MethodGet, path, &raw, opts...); err != nil {
		return types.DocumentPageDetail{}, err
	}
	return types.NewDocumentPageDetail(raw), nil
}

// documentChildResult maps the WikiSortDetail returned by the page and folder write endpoints.
func documentChildResult(ctx context.Context, method, path string, documentID int, payload map[string]interface{}, opts ...utils.Option) (*mcp.CallToolResult, error) {
	raw := types.DocumentPageDetailNode{}
	if err := getJSON(ctx, method, path, &raw, append(opts, utils.WithPayload(payload))...); err != nil {
		return utils.MCPErrorResult(err)
	}

	detail := types.NewDocumentPageDetail(raw)
	detail.ContentHTML = ""
	return utils.MCPJSONResult(types.DocumentPageResponse{DocumentID: documentID, Page: detail})
}

func pageIDList(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return strings.Join(parts, ",")
}
