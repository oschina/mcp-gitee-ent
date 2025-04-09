package types

type IssueType struct {
	Id          int    `json:"id"`
	Ident       string `json:"ident"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Template    string `json:"template"`
	Category    string `json:"category"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
