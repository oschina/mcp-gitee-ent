package types

type Program struct {
	Assignee      *BasicUser `json:"assignee"`
	Category      string     `json:"category"`
	CreatedAt     string     `json:"created_at"`
	Description   string     `json:"description"`
	Id            int        `json:"id"`
	Ident         string     `json:"ident"`
	IsTopped      bool       `json:"is_topped"`
	IssuesCount   int        `json:"issues_count"`
	Name          string     `json:"name"`
	ProjectsCount int        `json:"projects_count"`
	Status        int        `json:"status"`
	Type          string     `json:"type"`
	UsersCount    int        `json:"users_count"`
}
