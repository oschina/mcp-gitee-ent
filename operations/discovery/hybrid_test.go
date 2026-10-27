package discovery

import (
	"context"
	"errors"
	"testing"
)

type fakeSearcher struct {
	hits []Hit
	err  error
}

func (f fakeSearcher) Search(context.Context, string, int) ([]Hit, error) {
	return f.hits, f.err
}

func TestHybridSearcherRRF(t *testing.T) {
	hybrid := &HybridSearcher{
		Lexical: fakeSearcher{hits: []Hit{{Name: "a"}, {Name: "b"}}},
		Vector:  fakeSearcher{hits: []Hit{{Name: "b"}, {Name: "c"}}},
	}
	hits, err := hybrid.Search(context.Background(), "query", 3)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) != 3 || hits[0].Name != "b" {
		t.Fatalf("hits = %#v, want b first", hits)
	}
	if hits[0].Source != "hybrid" {
		t.Fatalf("source = %q, want hybrid", hits[0].Source)
	}
}

func TestHybridSearcherFallsBackToLexical(t *testing.T) {
	hybrid := &HybridSearcher{
		Lexical: fakeSearcher{hits: []Hit{{Name: "a"}}},
		Vector:  fakeSearcher{err: errors.New("vector down")},
	}
	hits, err := hybrid.Search(context.Background(), "query", 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) != 1 || hits[0].Name != "a" {
		t.Fatalf("hits = %#v, want lexical fallback", hits)
	}
}

func TestHybridSearcherBoostsExactName(t *testing.T) {
	hybrid := &HybridSearcher{
		Lexical: fakeSearcher{hits: []Hit{{Name: "other"}, {Name: "list_milestones"}}},
		Vector:  fakeSearcher{hits: []Hit{{Name: "other"}}},
	}
	hits, err := hybrid.Search(context.Background(), "list_milestones", 3)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(hits) == 0 || hits[0].Name != "list_milestones" {
		t.Fatalf("hits = %#v, want exact tool first", hits)
	}
}
