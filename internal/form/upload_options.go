package form

type UploadOptions struct {
	Albums []string `json:"albums"`
	Folder string   `json:"folder"`
}
