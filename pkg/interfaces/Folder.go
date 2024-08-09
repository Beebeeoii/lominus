// Package interfaces provide the fundamental blueprint for how each object
// looks like.
package interfaces

// CanvasFolderObject depicts the actual object return from Canvas.
// There are more fields being returned by Canvas, but these are just the
// relevant ones as of now.
type CanvasFolderObject struct {
	Id             int    `json:"id"`
	Name           string `json:"name"`
	FullName       string `json:"full_name"`
	HiddenForUser  bool   `json:"hidden_for_user"`
	FilesCount     int    `json:"files_count"`
	FoldersCount   int    `json:"folders_count"`
	ParentFolderId int    `json:"parent_folder_id"`
}
