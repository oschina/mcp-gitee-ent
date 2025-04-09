package types

// BasicIssue 定义了 Issue 的基本结构
type BasicIssue struct {
	Id                 int               `json:"id"`
	RootId             int               `json:"root_id"`
	ParentId           int               `json:"parent_id"`
	ProjectId          int               `json:"project_id"`
	Ident              string            `json:"ident"`
	Title              string            `json:"title"`
	ProgramId          int               `json:"program_id"`
	Author             *BasicUser        `json:"author"`
	State              string            `json:"state"`
	PriorityHuman      string            `json:"priority_human"`
	Branch             string            `json:"branch"`
	CreatedAt          string            `json:"created_at"`
	UpdatedAt          string            `json:"updated_at"`
	PlanStartedAt      string            `json:"plan_started_at"`
	Deadline           string            `json:"deadline"`
	Labels             []BasicLabel      `json:"labels"`
	Assignee           *BasicUser        `json:"assignee"`
	Collaborators      []*BasicUser      `json:"collaborators"`
	IssueType          *BasicIssueType   `json:"issue_type"`
	ScrumSprint        *BasicScrumSprint `json:"scrum_sprint"`
	EstimatedDuration  float64           `json:"estimated_duration"`
	RegisteredDuration float64           `json:"registered_duration"`
}

// IssueComment defines the structure for an issue comment
type IssueComment struct {
	Id        int       `json:"id"`
	Content   string    `json:"content"`
	Author    BasicUser `json:"author"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type IssueDetail struct {
	BasicIssue
	Description string `json:"description"`
}
