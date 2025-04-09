package types

type ScrumVersion struct {
	ID                int    `json:"id"`
	Title             string `json:"title"`
	State             string `json:"state"`
	Description       string `json:"description"`
	Number            string `json:"number"`
	TotalIssuesCount  int    `json:"total_issues_count"`
	ClosedIssuesCount int    `json:"closed_issues_count"`
	PlanReleasedAt    string `json:"plan_released_at"`
	ReleasedAt        string `json:"released_at"`
}
