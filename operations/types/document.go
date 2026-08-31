package types

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
