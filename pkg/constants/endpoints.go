package constants

// Canvas Endpoints
const (
	CANVAS_USER_SELF_ENDPOINT      = "https://canvas.nus.edu.sg/api/v1/users/self"
	CANVAS_MODULES_ENDPOINT        = "https://canvas.nus.edu.sg/api/v1/courses?per_page=30"
	CANVAS_MODULE_FOLDERS_ENDPOINT = "https://canvas.nus.edu.sg/api/v1/courses/%s/folders?per_page=30"
	CANVAS_FOLDERS_ENDPOINT        = "https://canvas.nus.edu.sg/api/v1/folders/%s/folders?per_page=30"
	CANVAS_FILES_ENDPOINT          = "https://canvas.nus.edu.sg/api/v1/folders/%s/files?per_page=30"
	CANVAS_FILE_ENDPOINT           = "https://canvas.nus.edu.sg/api/v1/files/%s"
)

// Luminus Endpoints
const (
	LUMINUS_AUTH_CODE_ENDPOINT     = "https://vafs.nus.edu.sg/adfs/oauth2/authorize?response_type=code&client_id=E10493A3B1024F14BDC7D0D8B9F649E9-234390&state=V6E9kYSq3DDQ72fSZZYFzLNKFT9dz38vpoR93IL8&redirect_uri=https://luminus.nus.edu.sg/auth/callback&scope=&resource=sg_edu_nus_oauth&nonce=V6E9kYSq3DDQ72fSZZYFzLNKFT9dz38vpoR93IL8"
	LUMINUS_AUTH_JWT_ENDPOINT      = "https://luminus.nus.edu.sg/v2/api/login/adfstoken"
	LUMINUS_AUTH_REDIRECT_ENDPOINT = "https://luminus.nus.edu.sg/auth/callback"
)
