package types

type EnterpriseMember struct {
	BlockMessage   string          `json:"block_message"`
	BlockStatus    int             `json:"block_status"`
	CreatedAt      string          `json:"created_at"`
	Email          string          `json:"email"`
	EnterpriseRole *enterpriseRole `json:"enterprise_role"`
	Id             int             `json:"id"`
	IsBlock        bool            `json:"is_block"`
	Name           string          `json:"name"`
	Occupation     string          `json:"occupation"`
	Phone          string          `json:"phone"`
	Pinyin         string          `json:"pinyin"`
	Remark         string          `json:"remark"`
	User           *BasicUser      `json:"user"`
	Username       string          `json:"username"`
}

type enterpriseRole struct {
	Id    int    `json:"id"`
	Ident string `json:"ident"`
	Name  string `json:"name"`
}
