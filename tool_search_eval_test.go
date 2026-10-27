package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"gitee.com/oschina/mcp-gitee-ent/operations/discovery"
)

type toolSearchEvalCase struct {
	Query    string   `json:"query"`
	Expected []string `json:"expected"`
	Kind     string   `json:"kind"`
}

func TestToolSearchEvalSet(t *testing.T) {
	resetToolsetFlags(t)
	registry, err := buildToolCatalog()
	if err != nil {
		t.Fatalf("buildToolCatalog() error = %v", err)
	}
	raw, err := os.ReadFile("testdata/tool_search_queries.json")
	if err != nil {
		t.Fatalf("read eval set: %v", err)
	}
	var cases []toolSearchEvalCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode eval set: %v", err)
	}
	if len(cases) < 50 {
		t.Fatalf("eval set has %d cases, want at least 50", len(cases))
	}

	searcher := discovery.NewLexicalSearcher(registry)
	total, passed, exactTotal, exactPassed := 0, 0, 0, 0
	for _, testCase := range cases {
		total++
		hits, err := searcher.Search(context.Background(), testCase.Query, 5)
		if err != nil {
			t.Fatalf("search %q: %v", testCase.Query, err)
		}
		names := make([]string, 0, len(hits))
		for _, hit := range hits {
			names = append(names, hit.Name)
		}
		match := false
		if testCase.Kind == "negative" {
			match = len(names) == 0
		} else {
			for _, name := range names {
				for _, expected := range testCase.Expected {
					if name == expected {
						match = true
					}
				}
			}
		}
		if match {
			passed++
		} else {
			t.Logf("MISS kind=%s query=%q got=%v expected=%v", testCase.Kind, testCase.Query, hits, testCase.Expected)
		}
		if testCase.Kind == "exact" {
			exactTotal++
			if len(names) > 0 && names[0] == testCase.Expected[0] {
				exactPassed++
			} else {
				t.Logf("EXACT_MISS query=%q got=%v expected=%s", testCase.Query, names, testCase.Expected[0])
			}
		}
	}

	rate := float64(passed) / float64(total)
	t.Logf("search eval: top5=%d/%d (%.1f%%), exact_top1=%d/%d", passed, total, rate*100, exactPassed, exactTotal)
	if exactPassed != exactTotal {
		t.Errorf("exact-name top1 = %d/%d, want 100%%", exactPassed, exactTotal)
	}
	if rate < 0.85 {
		t.Errorf("top5 pass rate = %.1f%%, want >= 85%%", rate*100)
	}
}
