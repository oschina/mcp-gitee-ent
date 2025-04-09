package types

type PagedResponse[T any] struct {
	TotalCount int `json:"total_count"`
	Data       []T `json:"data"`
}

type BasicUser struct {
	Id                 int    `json:"id"`
	Username           string `json:"username"`
	Name               string `json:"name"`
	AvatarUrl          string `json:"avatar_url"`
	IsEnterpriseMember bool   `json:"is_enterprise_member"`
	Remark             string `json:"remark"`
}

type BasicEnterprise struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type BasicProgram struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Assignee    BasicUser `json:"assignee"`
	Author      BasicUser `json:"author"`
}

type BasicLabel struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type BasicIssueType struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type BasicIssueState struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Namespace struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}

type UserInfo struct {
	Id        int    `json:"id"`
	Login     string `json:"login"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}
