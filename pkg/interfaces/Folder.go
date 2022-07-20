package interfaces

// TODO Documentation
type CanvasFolderObject struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HiddenForUser bool   `json:"hidden_for_user"`
	FilesCount    int    `json:"files_count"`
	FoldersCount  int    `json:"folders_count"`
}
