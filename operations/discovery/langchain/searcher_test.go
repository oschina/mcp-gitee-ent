package langchain

import (
	"context"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/utils"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tmc/langchaingo/schema"
)

type fakeRetriever struct {
	documents []schema.Document
	err       error
}

func (f fakeRetriever) GetRelevantDocuments(context.Context, string) ([]schema.Document, error) {
	return f.documents, f.err
}

func TestSearcherMapsToolNameMetadata(t *testing.T) {
	searcher := Searcher{Retriever: fakeRetriever{documents: []schema.Document{
		{PageContent: "milestones", Metadata: map[string]any{"tool_name": "list_milestones"}, Score: 0.9},
		{PageContent: "missing metadata"},
		{PageContent: "cases", Metadata: map[string]any{"tool_name": "list_test_cases"}, Score: 0.8},
	}}}
	hits, err := searcher.Search(context.Background(), "里程碑", 10)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) != 2 || hits[0].Name != "list_milestones" || hits[1].Name != "list_test_cases" {
		t.Fatalf("hits = %#v", hits)
	}
	if hits[0].Source != "embedding" {
		t.Fatalf("source = %q", hits[0].Source)
	}
}

func TestBuildDocumentsUsesCatalogMetadata(t *testing.T) {
	registry := catalog.NewRegistry()
	if err := registry.Register(catalog.Spec{
		Tool: mcp.NewTool("list_milestones", mcp.WithDescription("List milestones")),
		Handler: func(context.Context, mcp.CallToolRequest, ...utils.Option) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		},
		Domain:    "milestones",
		Aliases:   []string{"里程碑"},
		Invokable: true,
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	documents := BuildDocuments(registry)
	if len(documents) != 1 {
		t.Fatalf("documents = %d, want 1", len(documents))
	}
	if documents[0].Metadata["tool_name"] != "list_milestones" || documents[0].Metadata["domain"] != "milestones" {
		t.Fatalf("metadata = %#v", documents[0].Metadata)
	}
}
