package utils

import (
	"fmt"
	"strings"
)

// CaseStepFields lists the keys the Gitee API accepts for one test case step.
var CaseStepFields = []string{"description", "expected_result"}

// NormalizeCaseSteps validates the case_steps argument and returns the payload
// shape the Gitee API accepts. Each step must carry a non-empty description;
// expected_result is optional. Unknown keys (for example sort) are dropped.
func NormalizeCaseSteps(value any) ([]map[string]any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("case_steps must be an array of objects with a description")
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("case_steps must contain at least one step")
	}

	steps := make([]map[string]any, 0, len(items))
	for i, item := range items {
		step, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("case_steps[%d] must be an object with a description", i)
		}
		description, _ := step["description"].(string)
		if strings.TrimSpace(description) == "" {
			return nil, fmt.Errorf("case_steps[%d].description is required and must not be empty", i)
		}
		normalized := map[string]any{"description": strings.TrimSpace(description)}
		if expected, ok := step["expected_result"].(string); ok && strings.TrimSpace(expected) != "" {
			normalized["expected_result"] = strings.TrimSpace(expected)
		}
		steps = append(steps, normalized)
	}
	return steps, nil
}
