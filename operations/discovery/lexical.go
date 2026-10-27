package discovery

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"gitee.com/oschina/mcp-gitee-ent/catalog"
)

const (
	scoreNameExact      = 1000
	scoreNamePrefix     = 300
	scoreNameContains   = 200
	scoreAliasExact     = 150
	scoreAliasContains  = 90
	scoreSummaryFull    = 80
	scoreNameToken      = 40
	scoreAliasToken     = 25
	scoreSummaryToken   = 10
	scoreTagToken       = 8
	scoreDomainToken    = 6
	defaultLexicalLimit = 8
	maxLexicalLimit     = 100
)

// LexicalSearcher is the dependency-free default searcher. It ranks tools by
// exact name, prefix, aliases, summary, tags and domain matches.
type LexicalSearcher struct {
	registry *catalog.Registry
}

// NewLexicalSearcher creates a lexical searcher over the registry.
func NewLexicalSearcher(registry *catalog.Registry) *LexicalSearcher {
	return &LexicalSearcher{registry: registry}
}

// Search implements Searcher.
func (s *LexicalSearcher) Search(ctx context.Context, query string, limit int) ([]Hit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = defaultLexicalLimit
	}
	if limit > maxLexicalLimit {
		limit = maxLexicalLimit
	}

	rawQuery := strings.TrimSpace(strings.ToLower(query))
	if rawQuery == "" {
		return nil, nil
	}
	compactQuery := strings.ReplaceAll(normalizeText(rawQuery), " ", "")
	tokens := strings.Fields(normalizeText(rawQuery))
	normalizedQuery := normalizeText(rawQuery)

	hits := make([]Hit, 0)
	for _, spec := range s.registry.List() {
		if !spec.Invokable {
			continue
		}
		score := scoreSpec(spec, normalizedQuery, compactQuery, tokens)
		if score <= 0 {
			continue
		}
		hits = append(hits, Hit{Name: spec.Tool.Name, Score: score, Source: "lexical"})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].Name < hits[j].Name
		}
		return hits[i].Score > hits[j].Score
	})

	if len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}

func scoreSpec(spec catalog.Spec, normalizedQuery, compactQuery string, tokens []string) float64 {
	name := strings.ToLower(spec.Tool.Name)
	compactName := strings.ReplaceAll(normalizeText(name), " ", "")
	summary := strings.ToLower(spec.Summary + " " + spec.Tool.Description)
	domain := strings.ToLower(spec.Domain)

	var score float64
	switch compactName {
	case compactQuery:
		score += scoreNameExact
	default:
		if compactQuery != "" && strings.HasPrefix(compactName, compactQuery) {
			score += scoreNamePrefix
		} else if compactQuery != "" && strings.Contains(compactName, compactQuery) {
			score += scoreNameContains
		}
	}

	bestAliasScore := 0.0
	for _, alias := range spec.Aliases {
		normalizedAlias := strings.ToLower(strings.TrimSpace(alias))
		compactAlias := strings.ReplaceAll(normalizeText(normalizedAlias), " ", "")
		candidate := 0.0
		switch {
		case compactAlias == compactQuery:
			candidate = scoreAliasExact
		case compactQuery != "" && (strings.Contains(compactAlias, compactQuery) || strings.Contains(compactQuery, compactAlias)):
			candidate = scoreAliasContains + float64(len([]rune(compactAlias)))
		}
		if candidate > bestAliasScore {
			bestAliasScore = candidate
		}
	}
	score += bestAliasScore

	if compactQuery != "" && strings.Contains(normalizeText(summary), normalizedQuery) {
		score += scoreSummaryFull
	}

	for _, token := range tokens {
		if len([]rune(token)) < 2 && !containsCJK(token) {
			continue
		}
		if strings.Contains(compactName, token) {
			score += scoreNameToken
		}
		if aliasContains(spec.Aliases, token) {
			score += scoreAliasToken
		}
		if strings.Contains(summary, token) {
			score += scoreSummaryToken
		}
		if tagContains(spec.Tags, token) {
			score += scoreTagToken
		}
		if strings.Contains(domain, token) {
			score += scoreDomainToken
		}
	}

	return score
}

func aliasContains(aliases []string, token string) bool {
	for _, alias := range aliases {
		if strings.Contains(strings.ToLower(alias), token) {
			return true
		}
	}
	return false
}

func tagContains(tags []string, token string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), token) {
			return true
		}
	}
	return false
}

func normalizeText(value string) string {
	var builder strings.Builder
	previousSpace := false
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) {
			builder.WriteRune(r)
			previousSpace = false
			continue
		}
		if !previousSpace {
			builder.WriteRune(' ')
			previousSpace = true
		}
	}
	return strings.TrimSpace(builder.String())
}

func containsCJK(value string) bool {
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
