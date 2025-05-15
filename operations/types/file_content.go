package types

type FileContent struct {
	FileType   string `json:"file_type"`
	Size       int    `json:"size"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	RawContent string `json:"raw_content"`
}
