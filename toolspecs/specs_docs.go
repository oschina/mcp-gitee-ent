package toolspecs

import (
	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/operations/common_views"
	"gitee.com/oschina/mcp-gitee-ent/operations/documents"
	"gitee.com/oschina/mcp-gitee-ent/operations/kanbans"
)

func docSpecs() []catalog.Spec {
	return []catalog.Spec{
		visibleSpec(documents.CreateDocumentTool, documents.CreateDocumentHandleFunc, "documents", catalog.RiskWrite),
		visibleSpec(documents.CreateDocumentNodeTool, documents.CreateDocumentNodeHandleFunc, "documents", catalog.RiskWrite),
		visibleSpec(documents.RenameDocumentFolderTool, documents.RenameDocumentFolderHandleFunc, "documents", catalog.RiskWrite),
		visibleSpec(documents.ListDocumentsTool, documents.ListDocumentsHandleFunc, "documents", catalog.RiskRead),
		visibleSpec(documents.ListDocumentPagesTool, documents.ListDocumentPagesHandleFunc, "documents", catalog.RiskRead),
		visibleSpec(documents.GetDocumentPageTool, documents.GetDocumentPageHandleFunc, "documents", catalog.RiskRead),
		visibleSpec(documents.UploadDocumentTool, documents.UploadDocumentHandleFunc, "documents", catalog.RiskWrite),
		visibleSpec(documents.DownloadDocumentTool, documents.DownloadDocumentHandleFunc, "documents", catalog.RiskRead),
		visibleSpec(documents.UpdateDocumentPageTool, documents.UpdateDocumentPageHandleFunc, "documents", catalog.RiskWrite),
		hiddenSpec(documents.SearchDocumentsTool, documents.SearchDocumentsHandleFunc, "documents", catalog.RiskRead,
			[]string{"文档搜索", "知识库搜索", "search documents", "wiki search"},
			"Search documents and knowledge-base nodes.", map[string]any{"enterprise_id": 123, "search": "onboarding"}),
		hiddenSpec(documents.ListRecentDocumentsTool, documents.ListRecentDocumentsHandleFunc, "documents", catalog.RiskRead,
			[]string{"最近文档", "最近编辑文档", "recent documents"},
			"List recently edited documents.", map[string]any{"enterprise_id": 123, "program_id": 456}),
		hiddenSpec(documents.ListSharedDocumentsTool, documents.ListSharedDocumentsHandleFunc, "documents", catalog.RiskRead,
			[]string{"共享文档", "分享给我的文档", "shared documents"},
			"List documents shared with the current user.", map[string]any{"enterprise_id": 123}),
		hiddenSpec(documents.GetDocumentNodeTool, documents.GetDocumentNodeHandleFunc, "documents", catalog.RiskRead,
			[]string{"文档节点", "文档详情", "document node", "doc node detail"},
			"Get one document node by ID.", map[string]any{"enterprise_id": 123, "doc_node_id": 456}),
		hiddenSpec(documents.ListDocumentVersionsTool, documents.ListDocumentVersionsHandleFunc, "documents", catalog.RiskRead,
			[]string{"文档版本", "历史版本", "document versions"},
			"List document attachment versions.", map[string]any{"enterprise_id": 123, "doc_node_id": 456}),
		hiddenSpec(common_views.ListCommonViewsTool, common_views.ListCommonViewsHandleFunc, "views", catalog.RiskRead,
			[]string{"公共视图", "视图列表", "common views", "saved views"},
			"List saved common views.", map[string]any{"enterprise_id": 123, "type": "program_view"}),
		hiddenSpec(kanbans.ListKanbansTool, kanbans.ListKanbansHandleFunc, "kanbans", catalog.RiskRead,
			[]string{"看板列表", "kanbans", "boards"},
			"List kanban boards.", map[string]any{"enterprise_id": 123}),
		hiddenSpec(kanbans.ListKanbanColumnsTool, kanbans.ListKanbanColumnsHandleFunc, "kanbans", catalog.RiskRead,
			[]string{"看板栏", "看板列", "kanban columns"},
			"List columns of a kanban board.", map[string]any{"enterprise_id": 123, "program_id": 456, "kanban_id": 789}),
		hiddenSpec(documents.DeleteDocumentTool, documents.DeleteDocumentHandleFunc, "documents", catalog.RiskDestructive,
			[]string{"删除文档", "移除文档", "delete document"},
			"Delete an enterprise knowledge-base document or folder.",
			map[string]any{"enterprise_id": 123, "doc_node_id": 456}),
		hiddenSpec(documents.DeleteDocumentPageTool, documents.DeleteDocumentPageHandleFunc, "documents", catalog.RiskDestructive,
			[]string{"删除文档页面", "删除文件夹", "delete document page"},
			"Delete one page or folder inside a document.",
			map[string]any{"enterprise_id": 123, "document_id": 456, "page_id": 789}),
	}
}
