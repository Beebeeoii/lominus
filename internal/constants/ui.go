// Package constants provide constants to be used internally within Lominus (not exported as API)
// such as UI constants.
package constants

const (
	// Credentials Tab
	CREDENTIALS_TITLE = "Credentials"

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
	FILE_DIRECTORY_FOLDER_PATH_DEFAULT   = "Select the root folder to store all your modules' files."
	FILE_DIRECTORY_SELECT_DIRECTORY_TEXT = "Choose folder"
	FILE_DIRECTORY_CHOOSE_LOCATION_TEXT  = "Choose"

	SYNC_TAB_TITLE             = "Sync"
	SYNC_TAB_DESCRIPTION       = "Sync files and more **automatically** at the frequency specified below."
	SYNC_FREQUENCY_DISABLED    = "Disabled"
	SYNC_FREQUENCY_ONE_HOUR    = "1 hour"
	SYNC_FREQUENCY_TWO_HOUR    = "2 hour"
	SYNC_FREQUENCY_FOUR_HOUR   = "4 hour"
	SYNC_FREQUENCY_SIX_HOUR    = "6 hour"
	SYNC_FREQUENCY_TWELVE_HOUR = "12 hour"

	ADVANCED_TAB_TITLE                 = "Advanced"
	DEBUG_CHECKBOX_TITLE               = "Debug Mode"
	DEBUG_CHECKBOX_W_LINK_DESCRIPTION  = "Debug mode enables extensive logging to the [logfile](<%s>)."
	DEBUG_CHECKBOX_WO_LINK_DESCRIPTION = "Debug mode enables extensive logging to the logfile."
	DEBUG_TOGGLE_SUCCESSFUL_MESSAGE    = "Please restart Lominus for changes to take place."

	PREFERENCES_FAILED_MESSAGE = "An error has occurred :( Please try again"

	// Integrations Tab
	INTEGRATIONS_TITLE = "Integrations"

	TELEGRAM_TITLE                      = "Telegram"
	TELEGRAM_DESCRIPTION                = "Link Lominus to your Telegram bot to get notified of grades releases.\n\nYou will need:\n- your bot token via [BotFather](https://t.me/botfather) on Telegram\n- your Telegram user ID (one way is to use [userinfobot](https://t.me/userinfobot))"
	TELEGRAM_BOT_TOKEN_TEXT             = "Bot API Token"
	TELEGRAM_BOT_TOKEN_PLACEHOLDER      = "Your bot's API token"
	TELEGRAM_USER_ID_TEXT               = "User ID"
	TELEGRAM_USER_ID_PLACEHOLDER        = "Your user ID"
	TELEGRAM_DEFAULT_TEST_MESSAGE       = "Thank you for using Lominus! You have succesfully integrated Telegram with Lominus!\n\nBy integrating Telegram with Lominus, you will be notified of the following whenever Lominus polls for new update based on the intervals set:\n💥 new grades releases\n💥 new announcements (TBC)"
	TELEGRAM_TESTING_MESSAGE            = "Please wait while we send you a test message..."
	TELEGRAM_TESTING_SUCCESSFUL_MESSAGE = "Telegram integration successful!"
	TELEGRAM_TESTING_FAILED_MESSAGE     = "Telegram integration failed.\nPlease ensure that you have chatted with your bot before."
	SAVE_TELEGRAM_DATA_TEXT             = "Save Telegram Info"

	// General
	NO_FOLDER_DIRECTORY_SELECTED = "Please select a folder to store your files: Preferences > Folder Directory"
	NO_FREQUENCY_SELECTED        = "Please choose a sync frequency: Preferences > Sync"
	CANCEL_TEXT                  = "Cancel"
	SYNC_TEXT                    = "Sync"
	QUIT_LOMINUS_TEXT            = "Quit Lominus"

	DIALOG_PADDING = 30
)
