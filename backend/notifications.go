package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	lastAlertTime time.Time
	alertMutex    sync.Mutex
	alertCooldown = 30 * time.Minute
)

func SendSoilAlert(bot *tgbotapi.BotAPI, message string) error {
	// Your Telegram user/chat ID (get from @userinfobot on Telegram)
	chatID := int64(8138154689) // Replace with your chat ID

	msg := tgbotapi.NewMessage(chatID, message)
	_, err := bot.Send(msg)
	return err
}

// Usage in your soil monitoring system
func CheckSoilMoisture(bot *tgbotapi.BotAPI, moisture int) {
	if moisture < 2700 {
		alertMutex.Lock()
		defer alertMutex.Unlock()

		if time.Since(lastAlertTime) < alertCooldown {
			return
		}

		alert := fmt.Sprintf("ALERT: Soil is dry! needs water! Current: %d%%\nWatering pump activated.", moisture)
		if err := SendSoilAlert(bot, alert); err != nil {
			log.Printf("Failed to send Telegram alert: %v", err)
		} else {
			lastAlertTime = time.Now()
			log.Printf("Soil moisture alert sent: %d", moisture)
		}
	}
}
