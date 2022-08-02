package constants

const (
	// Credentials Tab
	CREDENTIALS_TITLE = "Credentials"

	LUMINUS_TAB_TITLE            = "Luminus"
	LUMINUS_TAB_DESCRIPTION      = `Credentials are saved **locally**. It is used for logging into [Luminus](https://luminus.nus.edu.sg) **only**.`
	LUMINUS_USERNAME_TEXT        = "Username"
	LUMINUS_USERNAME_PLACEHOLDER = "Eg: nusstu\\e0123456"
	LUMINUS_PASSWORD_TEXT        = "Password"
	LUMINUS_PASSWORD_PLACEHOLDER = "Password"

	CANVAS_TAB_TITLE         = "Canvas"
	CANVAS_TAB_DESCRIPTION   = `Token is saved **locally**. It is used to access [Canvas](https://canvas.nus.edu.sg/) **only**.`
	CANVAS_TOKEN_TEXT        = "Canvas Token"
	CANVAS_TOKEN_PLACEHOLDER = "Account > Settings > New access token > Generate Token"

	SAVE_CREDENTIALS_TEXT           = "Save Credentials"
	VERIFYING_MESSAGE               = "Please wait while we verify your credentials..."
	VERIFICATION_SUCCESSFUL_MESSAGE = "Verification successful."
	VERIFICATION_FAILED_MESSAGE     = "Verification failed. Please check your credentials."

	// Preferences Tab
	PREFERENCES_TITLE = "Preferences"

	FILE_DIRECTORY_TAB_TITLE             = "File Directory"
	FILE_DIRECTORY_TAB_DESCRIPTION       = "Select the root folder to store all your modules' files."
	FILE_DIRECTORY_FOLDER_PATH_DEFAULT   = "Folder path not set"
	FILE_DIRECTORY_SELECT_DIRECTORY_TEXT = "Choose folder"

	SYNC_TAB_TITLE             = "Sync"
	SYNC_TAB_DESCRIPTION       = "Lominus helps to sync files and more **automatically** at the frequency specified below."
	SYNC_FREQUENCY_DISABLED    = "Disabled"
	SYNC_FREQUENCY_ONE_HOUR    = "1 hour"
	SYNC_FREQUENCY_TWO_HOUR    = "2 hour"
	SYNC_FREQUENCY_FOUR_HOUR   = "4 hour"
	SYNC_FREQUENCY_SIX_HOUR    = "6 hour"
	SYNC_FREQUENCY_TWELVE_HOUR = "12 hour"

	ADVANCED_TAB_TITLE                 = "Advanced"
	DEBUG_CHECKBOX_TITLE               = "Debug Mode"
	DEBUG_CHECKBOX_W_LINK_DESCRIPTION  = "Debug mode enables extensive logging to the [logfile](<file://%s>)."
	DEBUG_CHECKBOX_WO_LINK_DESCRIPTION = "Debug mode enables extensive logging to the logfile."
	DEBUG_TOGGLE_SUCCESSFUL_MESSAGE    = "Please restart Lominus for changes to take place."

	PREFERENCES_FAILED_MESSAGE = "An error has occurred :( Please try again"

	// Integrations Tab
	INTEGRATIONS_TITLE = "Integrations"

	TELEGRAM_TITLE                      = "Telegram"
	TELEGRAM_DESCRIPTION                = "Lominus can be linked to your Telegram bot to notify you when new grades are released.\n\nTo get started, you will need to create a bot via [BotFather](https://t.me/botfather) to retrieve a bot API token.\n\nAfterwards, you will need your Telegram User ID. One way is to use [userinfobot](https://t.me/userinfobot)."
	TELEGRAM_BOT_TOKEN_TEXT             = "Bot API Token"
	TELEGRAM_BOT_TOKEN_PLACEHOLDER      = "Your bot's API token"
	TELEGRAM_USER_ID_TEXT               = "User ID"
	TELEGRAM_USER_ID_PLACEHOLDER        = "Your User ID"
	TELEGRAM_DEFAULT_TEST_MESSAGE       = "Thank you for using Lominus! You have succesfully integrated Telegram with Lominus!\n\nBy integrating Telegram with Lominus, you will be notified of the following whenever Lominus polls for new update based on the intervals set:\n💥 new grades releases\n💥 new announcements (TBC)"
	TELEGRAM_TESTING_MESSAGE            = "Please wait while we send you a test message..."
	TELEGRAM_TESTING_SUCCESSFUL_MESSAGE = "Telegram integration successful!"
	TELEGRAM_TESTING_FAILED_MESSAGE     = "Telegram integration failed. Please ensure that you have chatted with your bot before."
	SAVE_TELEGRAM_DATA_TEXT             = "Save Telegram Info"

	// General
	NO_FOLDER_DIRECTORY_SELECTED = "Please select a folder to store your files: Preferences > Folder Directory"
	NO_FREQUENCY_SELECTED        = "Please choose a sync frequency: Preferences > Sync"
	CANCEL_TEXT                  = "Cancel"
	SYNC_TEXT                    = "Sync"
	QUIT_LOMINUS_TEXT            = "Quit Lominus"
)
