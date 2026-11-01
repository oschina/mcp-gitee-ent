package toolspecs

import (
	"fmt"
	"strings"

	"gitee.com/oschina/mcp-gitee-ent/utils"
)

var canonicalParamRules = map[string][]paramRename{}

func addCanonicalRules(tool string, rules ...paramRename) {
	canonicalParamRules[tool] = append(canonicalParamRules[tool], rules...)
}

func init() {
	for _, tool := range []string{
		"comment_enterprise_pull",
		"create_enterprise_issue",
		"create_enterprise_repo_pull",
		"create_enterprise_repo_release",
		"get_check_run_annotations",
		"get_enterprise_pull_detail",
		"get_enterprise_pull_diff",
		"get_enterprise_repo_tree",
		"get_enterprise_repository_file_content",
		"list_commit_check_runs",
		"list_enterprise_issues",
		"list_enterprise_pull_comments",
		"list_enterprise_pulls",
		"list_enterprise_repo_releases",
		"list_milestones",
		"list_project_check_runs",
		"list_webhooks",
		"merge_enterprise_pull",
		"update_enterprise_issue",
		"update_enterprise_pull",
	} {
		addCanonicalRules(tool, paramRename{Legacy: "project_id", Canonical: "repository_id"})
	}
	for _, tool := range []string{"create_milestone", "search_members", "update_milestone"} {
		addCanonicalRules(tool, paramRename{Legacy: "project_ids", Canonical: "repository_ids"})
	}
	addCanonicalRules("list_enterprise_issues",
		paramRename{Legacy: "program_id", Canonical: "program_ids", Description: "Program IDs (Gitee 项目 IDs), comma-separated."},
		paramRename{Legacy: "priority", Canonical: "priorities", Description: "Comma-separated priorities (0-4)."},
	)
	addCanonicalRules("list_issue_related_issues",
		paramRename{Legacy: "priority", Canonical: "priorities", Description: "Comma-separated priorities."},
	)
	addCanonicalRules("create_test_report",
		paramRename{Legacy: "state", Canonical: "passed", Description: "Whether the test report is marked as passed."},
	)
	addCanonicalRules("update_test_report",
		paramRename{Legacy: "state", Canonical: "passed", Description: "Whether the test report is marked as passed."},
	)
	addCanonicalRules("update_test_case",
		paramRename{Legacy: "state", Canonical: "state_id", Description: "Case state ID: 0 draft, 1 reviewing, 2 passed, 3 failed, 4 deprecated."},
	)
	addCanonicalRules("update_test_plan_state",
		paramRename{Legacy: "state", Canonical: "state_id", Description: "Target state ID: 0 not started, 1 in progress, 2 completed, 3 suspended. Transitions: 0->1|3, 1->2|3, 3->0|1."},
	)
	addCanonicalRules("list_programs",
		paramRename{Legacy: "status", Canonical: "statuses", Description: "Program statuses (0: start, 1: pause, 2: close), comma-separated, e.g. 0,1."},
	)
	addCanonicalRules("create_enterprise_repo_pull",
		paramRename{Legacy: "assignee_id", Canonical: "assignee_ids", Description: "Assignee user IDs (comma-separated)."},
	)
	addCanonicalRules("list_enterprise_pulls",
		paramRename{Legacy: "assignee_id", Canonical: "assignee_ids", Description: "Filter by PR reviewer user IDs; separate multiple reviewers with commas."},
	)
	addCanonicalRules("list_test_plan_cases",
		paramRename{Legacy: "issue_id", Canonical: "linked_issue_id", Description: "Only return cases not yet linked to this issue."},
	)
	addCanonicalRules("list_test_reports",
		paramRename{
			Legacy: "test_plan_ids", Canonical: "test_plan_ids",
			Description: "Test plan IDs to filter by. Pass an array of numeric IDs.",
			NewType:     "array",
			NewItems:    map[string]any{"type": "number"},
			ToLegacy:    encodeIDList,
		},
	)
}

func encodeIDList(value any) (any, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("Parameter %q must be an array of numeric test plan IDs.", "test_plan_ids")
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		id, err := utils.SafelyConvertToInt(item)
		if err != nil {
			return nil, fmt.Errorf("Parameter %q must be an array of numeric test plan IDs.", "test_plan_ids")
		}
		parts = append(parts, fmt.Sprintf("%d", id))
	}
	return strings.Join(parts, ","), nil
}
