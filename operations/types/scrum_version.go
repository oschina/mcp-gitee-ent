package types

type ScrumVersion struct {
	ID                int    `json:"id"`
	Title             string `json:"title"`
	EnterpriseId      int    `json:"enterprise_id"`
	ProgramId         int    `json:"program_id"`
	State             string `json:"state"`
	Description       string `json:"description"`
	Number            string `json:"number"`
	TotalIssuesCount  int    `json:"total_issues_count"`
	ClosedIssuesCount int    `json:"closed_issues_count"`
	PlanReleasedAt    string `json:"plan_released_at"`
	ReleasedAt        string `json:"released_at"`
}
