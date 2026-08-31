package documents

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/operations/types"
	"gitee.com/oschina/mcp-gitee-ent/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

const maxDocumentImportFileSize = 6 * 1024 * 1024

var UploadDocumentTool = mcp.NewTool(
	UploadDocument,
	mcp.WithDescription("Upload and import local ZIP, HTML, TXT, or Markdown files into an existing enterprise knowledge-base document. This adds pages to an existing document; it does not create a new document. Each file must be no larger than 6 MiB."),
	mcp.WithReadOnlyHintAnnotation(false),
	mcp.WithDestructiveHintAnnotation(false),
	mcp.WithIdempotentHintAnnotation(false),
	mcp.WithOpenWorldHintAnnotation(true),
	mcp.WithNumber("enterprise_id", mcp.Description("Enterprise ID returned by list_enterprises."), mcp.Required(), mcp.Min(1)),
	mcp.WithNumber("document_id", mcp.Description("Target document ID returned by list_enterprise_documents."), mcp.Required(), mcp.Min(1)),
	mcp.WithArray(
		"file_paths",
		mcp.Description("Absolute paths on the machine running the MCP server. Supported extensions: .zip, .html, .txt, and .md. Imported content may contain at most 20 pages."),
		mcp.Required(),
		mcp.MinItems(1),
		mcp.MaxItems(20),
		mcp.Items(map[string]interface{}{"type": "string"}),
	),
)

func UploadDocumentHandleFunc(ctx context.Context, request mcp.CallToolRequest, opts ...utils.Option) (*mcp.CallToolResult, error) {
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
	filePaths, err := stringArray(arguments, "file_paths")
	if err != nil {
		return utils.MCPErrorResult(err)
	}
	if len(filePaths) == 0 || len(filePaths) > 20 {
		return utils.MCPErrorResult(utils.NewParamError("file_paths", "provide between 1 and 20 files"))
	}

	resolvedPaths := make([]string, 0, len(filePaths))
	for _, filePath := range filePaths {
		resolved, err := utils.ResolveReadableFile(filePath)
		if err != nil {
			return utils.MCPErrorResult(err)
		}
		if err := validateImportFile(resolved); err != nil {
			return utils.MCPErrorResult(err)
		}
		resolvedPaths = append(resolvedPaths, resolved)
	}

	body, contentType := newMultipartFileBody(resolvedPaths)
	defer body.Close()
	path := fmt.Sprintf("/%d/docs/%d/upload", enterpriseID, documentID)
	client := utils.NewGiteeClient(http.MethodPost, path, append(opts, utils.WithBody(body, contentType))...)
	_, err = client.DoContext(ctx)
	if err != nil {
		_ = body.CloseWithError(err)
		return utils.MCPErrorResult(err)
	}
	defer client.Response.Body.Close()

	response := types.DocumentUploadResponse{DocumentID: documentID, Imported: true}
	pages, refreshErr := fetchDocumentPages(ctx, enterpriseID, documentID, opts...)
	if refreshErr != nil {
		response.Warning = "Files were imported, but the updated page list could not be loaded: " + refreshErr.Error()
	} else {
		response.Pages = pages.Pages
	}
	return utils.MCPJSONResult(response)
}

func validateImportFile(path string) error {
	extension := strings.ToLower(filepath.Ext(path))
	switch extension {
	case ".zip", ".html", ".txt", ".md":
	default:
		return utils.NewParamError("file_paths", fmt.Sprintf("unsupported file extension %q for %s", extension, filepath.Base(path)))
	}

	info, err := os.Stat(path)
	if err != nil {
		return utils.NewParamError("file_paths", fmt.Sprintf("cannot inspect %q: %v", path, err))
	}
	if info.Size() > maxDocumentImportFileSize {
		return utils.NewParamError("file_paths", fmt.Sprintf("%s is larger than 6 MiB", filepath.Base(path)))
	}
	return nil
}

func newMultipartFileBody(paths []string) (*io.PipeReader, string) {
	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	contentType := multipartWriter.FormDataContentType()

	go func() {
		var writeErr error
		defer func() {
			if writeErr != nil {
				_ = writer.CloseWithError(writeErr)
				return
			}
			if closeErr := multipartWriter.Close(); closeErr != nil {
				_ = writer.CloseWithError(closeErr)
				return
			}
			_ = writer.Close()
		}()

		for _, path := range paths {
			file, err := os.Open(path)
			if err != nil {
				writeErr = err
				return
			}

			part, err := multipartWriter.CreateFormFile("files[]", importFilename(path))
			if err == nil {
				_, err = io.Copy(part, file)
			}
			closeErr := file.Close()
			if err != nil {
				writeErr = err
				return
			}
			if closeErr != nil {
				writeErr = closeErr
				return
			}
		}
	}()

	return reader, contentType
}

func importFilename(path string) string {
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)
	return strings.TrimSuffix(filename, extension) + strings.ToLower(extension)
}
