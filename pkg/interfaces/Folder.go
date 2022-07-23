package interfaces

type FolderObject interface {
	CanvasFolderObject | LuminusFolderObject
}

// TODO Documentation
type CanvasFolderObject struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HiddenForUser bool   `json:"hidden_for_user"`
	FilesCount    int    `json:"files_count"`
	FoldersCount  int    `json:"folders_count"`
}

type LuminusFolderObject struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	IsActive     bool   `json:"isActive"`
	AllowUpload  bool   `json:"alowUpload"`
	FilesCount   int    `json:"totalFilesCount"`
	FoldersCount int    `json:"subFolderCount"`
}

func (i CanvasFolderObject) GetId() int {
	return i.Id
}

func (i LuminusFolderObject) GetId() string {
	return i.Id
}
