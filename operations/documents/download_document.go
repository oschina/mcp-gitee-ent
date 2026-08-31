package documents

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

var DownloadDocumentTool = mcp.NewTool(
	DownloadDocument,
	mcp.WithDescription("Export an enterprise knowledge-base document and save it to the machine running the MCP server. Markdown and HTML exports may produce ZIP files; PDF exports produce a PDF file."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(true),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithString("format", mcp.Description("Export format: markdown, html, or pdf. Default: markdown."), mcp.Enum("markdown", "html", "pdf"), mcp.DefaultString("markdown")),
	mcp.WithString("output_directory", mcp.Description("Absolute destination directory on the machine running the MCP server."), mcp.Required()),
	mcp.WithBoolean("overwrite", mcp.Description("Overwrite an existing file with the same name. Default: false."), mcp.DefaultBool(false)),
)

func DownloadDocumentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	exportFormat, err := optionalString(arguments, "format", "markdown")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if exportFormat != "markdown" && exportFormat != "html" && exportFormat != "pdf" {
		return utils.MCPErrorResult(utils.NewParamError("format", "must be markdown, html, or pdf"))
	}
	outputDirectory, err := requiredString(arguments, "output_directory")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	outputDirectory, err = utils.ResolveWritableDirectory(outputDirectory)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	overwrite, err := optionalBool(arguments, "overwrite", false)
	if err != nil {
		return utils.MCPErrorResult(err)
	}

	path := fmt.Sprintf("/%d/docs/%d/export", enterpriseID, documentID)
	query := map[string]interface{}{"export_type": exportFormat}
	client := utils.NewGiteeClient(http.MethodGet, path, append(opts, utils.WithQuery(query))...)
	_, err = client.DoContext(ctx)
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	defer client.Response.Body.Close()

	filename := downloadFilename(client.Response, documentID, exportFormat)
	destination := filepath.Join(outputDirectory, filename)
	if _, err := os.Lstat(destination); err == nil && !overwrite {
		return utils.MCPErrorResult(utils.NewParamError("overwrite", fmt.Sprintf("%s already exists", destination)))
	} else if err != nil && !os.IsNotExist(err) {
		return utils.MCPErrorResult(utils.NewInternalError(fmt.Errorf("inspect destination: %w", err)))
	}

	temporary, err := os.CreateTemp(outputDirectory, ".mcp-gitee-ent-download-*")
	if err != nil {
		return utils.MCPErrorResult(utils.NewInternalError(fmt.Errorf("create temporary download file: %w", err)))
	}
	temporaryPath := temporary.Name()
	keepTemporary := false
	defer func() {
		_ = temporary.Close()
		if !keepTemporary {
			_ = os.Remove(temporaryPath)
		}
	}()

	size, err := io.Copy(temporary, client.Response.Body)
	if err != nil {
		return utils.MCPErrorResult(utils.NewNetworkError(fmt.Errorf("download document: %w", err)))
	}
	if err := temporary.Close(); err != nil {
		return utils.MCPErrorResult(utils.NewInternalError(fmt.Errorf("close downloaded file: %w", err)))
	}
	if overwrite {
		if err := os.Rename(temporaryPath, destination); err != nil {
			return utils.MCPErrorResult(utils.NewInternalError(fmt.Errorf("replace downloaded file: %w", err)))
		}
	} else {
		if err := os.Link(temporaryPath, destination); err != nil {
			if os.IsExist(err) {
				return utils.MCPErrorResult(utils.NewParamError("overwrite", fmt.Sprintf("%s already exists", destination)))
			}
			return utils.MCPErrorResult(utils.NewInternalError(fmt.Errorf("save downloaded file: %w", err)))
		}
		if err := os.Remove(temporaryPath); err != nil {
			return utils.MCPErrorResult(utils.NewInternalError(fmt.Errorf("remove temporary download file: %w", err)))
		}
	}
	keepTemporary = true

	return utils.MCPJSONResult(types.DocumentDownloadResponse{
		DocumentID:  documentID,
		Format:      exportFormat,
		SavedPath:   destination,
		Filename:    filename,
		Size:        size,
		ContentType: client.Response.Header.Get("Content-Type"),
	})
}

func downloadFilename(response *http.Response, documentID int, exportFormat string) string {
	contentDisposition := response.Header.Get("Content-Disposition")
	if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
		if filename := safeFilename(params["filename"]); filename != "" {
			return filename
		}
	}

	extension := ".zip"
	if exportFormat == "pdf" {
		extension = ".pdf"
	}
	return fmt.Sprintf("document-%d%s", documentID, extension)
}

func safeFilename(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}
	filename = filepath.Base(filename)
	if filename == "." || filename == string(filepath.Separator) {
		return ""
	}
	return filename
}
