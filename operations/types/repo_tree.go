package types

type RepoTree struct {
	Ref  string `json:"ref"`
	Path string `json:"path"`
	Tree tree   `json:"tree"`
}

type tree struct {
	Folders []fileObject `json:"folders"`
	Files   []fileObject `json:"files"`
}

type fileObject struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	RawPath string `json:"raw_path"`
}
