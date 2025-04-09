package types

type Release struct {
	Id             int       `json:"id"`
	Title          string    `json:"title"`
	IsPrerelease   bool      `json:"is_prerelease"`
	Description    string    `json:"description"`
	Author         BasicUser `json:"author"`
	ZipDownloadUrl string    `json:"zip_download_url"`
	TarDownloadUrl string    `json:"tar_download_url"`
	CreatedAt      string    `json:"created_at"`
}

type Tag struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Commit  commit `json:"commit"`
}

type ReleaseDetail struct {
	Release Release `json:"release"`
	Tag     Tag     `json:"tag"`
}

type commit struct {
	Id            string    `json:"id"`
	ShortId       string    `json:"short_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Message       string    `json:"message"`
	CompleteTitle string    `json:"complete_title"`
	Author        BasicUser `json:"author"`
	Committer     BasicUser `json:"committer"`
}
