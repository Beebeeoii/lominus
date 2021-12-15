package telegram

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/beebeeoii/lominus/internal/file"
)

type TelegramInfo struct {
	BotApi string
	UserId string
}

type TelegramError struct {
	Description string
}

const SEND_MSG_URL = "https://api.telegram.org/bot%s/sendMessage"
const CONTENT_TYPE = "application/x-www-form-urlencoded"
const POST = "POST"

func SendMessage(botApi string, userId string, message string) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	reqBody := url.Values{}
	reqBody.Set("chat_id", userId)
	reqBody.Set("text", message)
	sendMsgReq, sendMsgErr := http.NewRequest(POST, fmt.Sprintf(SEND_MSG_URL, botApi), strings.NewReader(reqBody.Encode()))

	if sendMsgErr != nil {
		return sendMsgErr
	}

	sendMsgReq.Header.Add("Content-Type", CONTENT_TYPE)
	sendMsgRes, sendMsgResErr := client.Do(sendMsgReq)

	if sendMsgResErr != nil {
		return sendMsgResErr
	}

	if sendMsgRes.StatusCode != 200 {
		bodyBytes, err := io.ReadAll(sendMsgRes.Body)
		if err != nil {
			return err
		}
		bodyString := string(bodyBytes)
		return &TelegramError{Description: bodyString}
	}

	return nil
}

func SaveTelegramData(telegramDataPath string, telegramInfo TelegramInfo) error {
	return file.EncodeStructToFile(telegramDataPath, telegramInfo)
}

func LoadTelegramData(telegramDataPath string) (TelegramInfo, error) {
	telegramInfo := TelegramInfo{}
	if !file.Exists(telegramDataPath) {
		return telegramInfo, &file.FileNotFoundError{FileName: telegramDataPath}
	}
	err := file.DecodeStructFromFile(telegramDataPath, &telegramInfo)

	return telegramInfo, err
}

func (e *TelegramError) Error() string {
	return fmt.Sprintf("TelegramError: %s", e.Description)
}
