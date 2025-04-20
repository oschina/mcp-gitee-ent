package types

type BasicScrumSprint struct {
	Id               int    `json:"id"`
	Title            string `json:"title"`
	EnterpriseId     int    `json:"enterprise_id"`
	ProgramId        int    `json:"program_id"`
	State            string `json:"state"`
	StartedAt        string `json:"started_at"`
	FinishedAt       string `json:"finished_at"`
	ActualFinishedAt string `json:"actual_finished_at"`
}

type ScrumSprint struct {
	BasicScrumSprint
	Description       string    `json:"description"`
	TimeScale         float64   `json:"time_scale"`
	OpenIssuesCount   int       `json:"open_issues_count"`
	ClosedIssuesCount int       `json:"closed_issues_count"`
	TotalIssuesCount  int       `json:"total_issues_count"`
	TotalFeatureCount int       `json:"total_feature_count"`
	TotalBugCount     int       `json:"total_bug_count"`
	TotalDuration     float64   `json:"total_duration"`
	Assignee          BasicUser `json:"assignee"`
}
