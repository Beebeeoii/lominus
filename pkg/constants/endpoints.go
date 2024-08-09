// Package constants provides constants such as web endpoints.
package constants

// Canvas Endpoints
const (
	CANVAS_USER_SELF_ENDPOINT      = "https://canvas.nus.edu.sg/api/v1/users/self"
	CANVAS_MODULES_ENDPOINT        = "https://canvas.nus.edu.sg/api/v1/dashboard/dashboard_cards"
	CANVAS_MODULE_FOLDER_ENDPOINT  = "https://canvas.nus.edu.sg/api/v1/courses/%s/folders/by_path/"
	CANVAS_MODULE_FOLDERS_ENDPOINT = "https://canvas.nus.edu.sg/api/v1/courses/%s/folders"
	CANVAS_FOLDERS_ENDPOINT        = "https://canvas.nus.edu.sg/api/v1/folders/%s/folders"
	CANVAS_FILES_ENDPOINT          = "https://canvas.nus.edu.sg/api/v1/folders/%s/files"
	CANVAS_FILE_ENDPOINT           = "https://canvas.nus.edu.sg/api/v1/files/%s"
)

// Telegram Endpoints
const (
	TELEGRAM_SEND_MESSAGE_ENDPOINT = "https://api.telegram.org/bot%s/sendMessage"
)
