package types

type BasicRepository struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	Public            int    `json:"public"`
	EnterpriseId      int    `json:"enterprise_id"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type Repository struct {
	BasicRepository
	Creator           *BasicUser       `json:"creator"`
	Description       string           `json:"description"`
	ForkedCount       int              `json:"forked_count"`
	GetDefaultBranch  string           `json:"get_default_branch"`
	IsFork            bool             `json:"is_forked"`
	LastPushAt        string           `json:"last_push_at"`
	MembersCount      int              `json:"members_count"`
	NameWithNamespace string           `json:"name_with_namespace"`
	Namespace         *Namespace       `json:"namespace"`
	Outsourced        bool             `json:"outsourced"`
	ParentProject     *BasicRepository `json:"parent_project"`
	RepoSize          int              `json:"repo_size"`
	StarsCount        int              `json:"stars_count"`
	StatusName        string           `json:"status_name"`
	WatchesCount      int              `json:"watches_count"`
}
