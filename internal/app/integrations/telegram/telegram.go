// Package intTelegram provides path retrievers for Lominus Telegram integration files.
package intTelegram

import (
	"github.com/beebeeoii/lominus/internal/app"
	"github.com/boltdb/bolt"
)

type TelegramIds struct {
	UserId string
	BotId  string
}

func GetTelegramIds() (TelegramIds, error) {
	dbInstance := app.GetDBInstance()
	var telegramIds TelegramIds

	err := dbInstance.View(func(tx *bolt.Tx) error {
		intBucket := tx.Bucket([]byte("Integrations"))
		telegramUserId := string(intBucket.Get([]byte("telegramUserId")))
		telegramBotId := string(intBucket.Get([]byte("telegramBotId")))

		telegramIds.UserId = telegramUserId
		telegramIds.BotId = telegramBotId

		return nil
	})

	if err != nil {
		return TelegramIds{}, err
	}

	return telegramIds, nil
}

// SaveTelegramCredentials saves the user's Telegram userId and botId locally.
func SaveTelegramCredentials(userId string, botId string) error {
	dbInstance := app.GetDBInstance()

	updateErr := dbInstance.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("Integrations")).Put([]byte("telegramUserId"), []byte(userId))
		err1 := tx.Bucket([]byte("Integrations")).Put([]byte("telegramBotId"), []byte(botId))

		if err != nil {
			return err
		}

		return err1
	})

	return updateErr
}
