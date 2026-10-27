package discovery

import (
	"context"
	"log"
	"sort"
	"strings"
)

const rrfK = 60.0

type HybridSearcher struct {
	Lexical Searcher
	Vector  Searcher
}

func (h *HybridSearcher) Search(ctx context.Context, query string, limit int) ([]Hit, error) {
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	if h.Vector == nil {
		return h.Lexical.Search(ctx, query, limit)
	}

	candidateLimit := limit * 2
	lexicalHits, lexicalErr := h.Lexical.Search(ctx, query, candidateLimit)
	vectorHits, vectorErr := h.Vector.Search(ctx, query, candidateLimit)

	if vectorErr != nil {
		if lexicalErr != nil {
			return nil, lexicalErr
		}
		log.Printf("tool vector search failed, falling back to lexical: %v", vectorErr)
		return trimHits(lexicalHits, limit), nil
	}
	if lexicalErr != nil {
		return trimHits(vectorHits, limit), nil
	}

	scores := make(map[string]float64)
	names := make(map[string]bool)
	for rank, hit := range lexicalHits {
		scores[hit.Name] += 1.0 / (rrfK + float64(rank+1))
		names[hit.Name] = true
	}
	for rank, hit := range vectorHits {
		scores[hit.Name] += 1.0 / (rrfK + float64(rank+1))
		names[hit.Name] = true
	}

	hits := make([]Hit, 0, len(names))
	for name := range names {
		score := scores[name]
		if strings.EqualFold(name, query) {
			score += 10
		}
		hits = append(hits, Hit{Name: name, Score: score, Source: "hybrid"})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Name < hits[j].Name
		}
		return hits[i].Score > hits[j].Score
	})
	return trimHits(hits, limit), nil
}

func trimHits(hits []Hit, limit int) []Hit {
	if len(hits) > limit {
		return hits[:limit]
	}
	return hits
}
