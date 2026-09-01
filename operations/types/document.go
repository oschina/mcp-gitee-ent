package types

import (
	"regexp"
	"strings"
)

// searchHighlightTagRegexp matches the <em class="gitee-highlight-keyword">
// highlight tags that Gitee injects into keyword-search results.
var searchHighlightTagRegexp = regexp.MustCompile(`</?em[^>]*>`)

// StripSearchHighlight removes highlight markup and trims the value.
func StripSearchHighlight(value string) string {
	return strings.TrimSpace(searchHighlightTagRegexp.ReplaceAllString(value, ""))
}

const searchContentPreviewRunes = 300

// SearchContentPreview strips highlight markup, collapses whitespace, and
// truncates the value so a full page body is never returned as a preview.
func SearchContentPreview(value string) string {
	cleaned := strings.Join(strings.Fields(StripSearchHighlight(value)), " ")
	if runes := []rune(cleaned); len(runes) > searchContentPreviewRunes {
		return string(runes[:searchContentPreviewRunes]) + "…"
	}
	return cleaned
}

type DocumentNode struct {
	ID         int           `json:"id"`
	ParentID   int           `json:"parent_id"`
	Name       string        `json:"name"`
	FilePath   string        `json:"file_path"`
	FileType   string        `json:"file_type"`
	FileID     int           `json:"file_id"`
	Program    *BasicProgram `json:"program,omitempty"`
	Creator    *BasicUser    `json:"creator,omitempty"`
	EditorType string        `json:"editor_type"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

type Document struct {
	DocumentID     int           `json:"document_id"`
	DocumentNodeID int           `json:"document_node_id"`
	ParentID       int           `json:"parent_id"`
	Name           string        `json:"name"`
	FilePath       string        `json:"file_path"`
	Program        *BasicProgram `json:"program,omitempty"`
	Creator        *BasicUser    `json:"creator,omitempty"`
	EditorType     string        `json:"editor_type"`
	CreatedAt      string        `json:"created_at"`
	UpdatedAt      string        `json:"updated_at"`
}

func NewDocument(node DocumentNode) Document {
	return Document{
		DocumentID:     node.FileID,
		DocumentNodeID: node.ID,
		ParentID:       node.ParentID,
		Name:           node.Name,
		FilePath:       node.FilePath,
		Program:        node.Program,
		Creator:        node.Creator,
		EditorType:     node.EditorType,
		CreatedAt:      node.CreatedAt,
		UpdatedAt:      node.UpdatedAt,
	}
}

type DocumentPageNode struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	ParentID     int                `json:"parent_id"`
	FileType     int                `json:"file_type"`
	Version      string             `json:"version"`
	LastEditTime string             `json:"last_edit_time"`
	ContentType  string             `json:"content_type"`
	CreatedAt    string             `json:"created_at"`
	UpdatedAt    string             `json:"updated_at"`
	Children     []DocumentPageNode `json:"children,omitempty"`
}

type DocumentPage struct {
	PageID       int    `json:"page_id"`
	ParentPageID int    `json:"parent_page_id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ContentType  string `json:"content_type,omitempty"`
	Version      string `json:"version,omitempty"`
	LastEditTime string `json:"last_edit_time,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

type DocumentPageDetailNode struct {
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	ParentID     int        `json:"parent_id"`
	FileType     int        `json:"file_type"`
	Version      string     `json:"version"`
	LastEditTime string     `json:"last_edit_time"`
	ContentType  string     `json:"content_type"`
	Content      string     `json:"content"`
	ContentHTML  string     `json:"content_html,omitempty"`
	Creator      *BasicUser `json:"creator,omitempty"`
	Editor       *BasicUser `json:"editor,omitempty"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

type DocumentPageDetail struct {
	PageID       int        `json:"page_id"`
	Name         string     `json:"name"`
	ParentPageID int        `json:"parent_page_id"`
	Type         string     `json:"type"`
	Version      string     `json:"version"`
	LastEditTime string     `json:"last_edit_time"`
	ContentType  string     `json:"content_type"`
	Content      string     `json:"content"`
	ContentHTML  string     `json:"content_html,omitempty"`
	Creator      *BasicUser `json:"creator,omitempty"`
	Editor       *BasicUser `json:"editor,omitempty"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

func NewDocumentPageDetail(node DocumentPageDetailNode) DocumentPageDetail {
	pageType := "page"
	if node.FileType == 2 {
		pageType = "folder"
	}

	return DocumentPageDetail{
		PageID:       node.ID,
		Name:         node.Name,
		ParentPageID: node.ParentID,
		Type:         pageType,
		Version:      node.Version,
		LastEditTime: node.LastEditTime,
		ContentType:  node.ContentType,
		Content:      node.Content,
		ContentHTML:  node.ContentHTML,
		Creator:      node.Creator,
		Editor:       node.Editor,
		CreatedAt:    node.CreatedAt,
		UpdatedAt:    node.UpdatedAt,
	}
}

type DocumentListResponse struct {
	TotalCount int        `json:"total_count"`
	Documents  []Document `json:"documents"`
}

type DocumentPagesResponse struct {
	DocumentID int            `json:"document_id"`
	TotalCount int            `json:"total_count"`
	Pages      []DocumentPage `json:"pages"`
}

type DocumentPageResponse struct {
	DocumentID int                `json:"document_id"`
	Page       DocumentPageDetail `json:"page"`
}

type DocumentUploadResponse struct {
	DocumentID int            `json:"document_id"`
	Imported   bool           `json:"imported"`
	Pages      []DocumentPage `json:"pages,omitempty"`
	Warning    string         `json:"warning,omitempty"`
}

type DocumentDownloadResponse struct {
	DocumentID  int    `json:"document_id"`
	Format      string `json:"format"`
	SavedPath   string `json:"saved_path"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
}

// DocumentSearchNode is one page-level hit returned by
// GET /{enterprise_id}/doc_nodes/search. Name and Content may carry highlight
// markup; name is absent when the server falls back to database search.
type DocumentSearchNode struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Content       string     `json:"content"`
	InfoID        int        `json:"info_id"`
	InfoName      string     `json:"info_name"`
	DocPath       string     `json:"doc_path"`
	DocID         int        `json:"doc_id"`
	Creator       *BasicUser `json:"creator"`
	LastUpdatedAt string     `json:"last_updated_at"`
}

type DocumentSearchHit struct {
	DocumentID     int        `json:"document_id"`
	DocumentNodeID int        `json:"document_node_id"`
	DocumentName   string     `json:"document_name"`
	DocumentPath   string     `json:"document_path"`
	PageID         int        `json:"page_id"`
	PageName       string     `json:"page_name,omitempty"`
	ContentPreview string     `json:"content_preview,omitempty"`
	Creator        *BasicUser `json:"creator,omitempty"`
	UpdatedAt      string     `json:"updated_at,omitempty"`
}

type DocumentSearchResponse struct {
	Search     string              `json:"search"`
	TotalCount int                 `json:"total_count"`
	Hits       []DocumentSearchHit `json:"hits"`
}

func NewDocumentSearchHit(node DocumentSearchNode) DocumentSearchHit {
	return DocumentSearchHit{
		DocumentID:     node.InfoID,
		DocumentNodeID: node.DocID,
		DocumentName:   StripSearchHighlight(node.InfoName),
		DocumentPath:   node.DocPath,
		PageID:         node.ID,
		PageName:       StripSearchHighlight(node.Name),
		ContentPreview: SearchContentPreview(node.Content),
		Creator:        node.Creator,
		UpdatedAt:      node.LastUpdatedAt,
	}
}
