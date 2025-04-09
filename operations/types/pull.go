package types

type BasicPull struct {
	Id           int          `json:"id"`
	Iid          int          `json:"iid"`
	ProjectId    int          `json:"project_id"`
	Title        string       `json:"title"`
	State        string       `json:"state"`
	Draft        bool         `json:"draft"`
	CheckState   int          `json:"check_state"`
	TestState    int          `json:"test_state"`
	Labels       []BasicLabel `json:"labels"`
	Author       BasicUser    `json:"author"`
	TargetBranch Reference    `json:"target_branch"`
	SourceBranch Reference    `json:"source_branch"`
	Conflict     bool         `json:"conflict"`
	CanMerge     bool         `json:"can_merge"`
	PrAssignNum  int          `json:"pr_assign_num"`
	PrTestNum    int          `json:"pr_test_num"`
	Assignees    []Assignee   `json:"assignees"`
	Testers      []Assignee   `json:"testers"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
	MergedAt     string       `json:"merged_at"`
}

type Reference struct {
	Branch string `json:"branch"`
	Name   string `json:"name"`
}

type Assignee struct {
	BasicUser
	State int `json:"state"`
}

type PullDiff struct {
}

type PullDetail struct {
	BasicPull
	Body string `json:"body"`
}

type PullComment struct {
	Id        int            `json:"id"`
	Author    BasicUser      `json:"author"`
	Content   string         `json:"content"`
	CreatedAt string         `json:"created_at"`
	Parent    *ParentComment `json:"parent"`
}

type ParentComment struct {
	Id     int       `json:"id"`
	Author BasicUser `json:"author"`
}
