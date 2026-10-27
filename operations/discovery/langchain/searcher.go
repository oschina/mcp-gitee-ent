package langchain

import (
	"context"
	"fmt"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
	"gitee.com/oschina/mcp-gitee-ent/operations/discovery"

	"github.com/tmc/langchaingo/schema"
)

type Searcher struct {
	Retriever schema.Retriever
}

func (s Searcher) Search(ctx context.Context, query string, limit int) ([]discovery.Hit, error) {
	if s.Retriever == nil {
		return nil, fmt.Errorf("langchain retriever is nil")
	}
	if limit <= 0 {
		limit = 8
	}
	documents, err := s.Retriever.GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, err
	}
	hits := make([]discovery.Hit, 0, limit)
	for _, document := range documents {
		name, _ := document.Metadata["tool_name"].(string)
		if name == "" {
			continue
		}
		hits = append(hits, discovery.Hit{
			Name:   name,
			Score:  float64(document.Score),
			Source: "embedding",
		})
		if len(hits) >= limit {
			break
		}
	}
	return hits, nil
}

func BuildDocuments(registry *catalog.Registry) []schema.Document {
	catalogDocuments := registry.Documents()
	documents := make([]schema.Document, 0, len(catalogDocuments))
	for _, document := range catalogDocuments {
		documents = append(documents, schema.Document{
			PageContent: document.SearchText,
			Metadata: map[string]any{
				"tool_name": document.Name,
				"domain":    document.Domain,
			},
		})
	}
	return documents
}
