package main

import (
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendSoilAlert(message string) error {
	// Get token from environment variable
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	// Your Telegram user/chat ID (get from @userinfobot on Telegram)
	chatID := int64(8138154689) // Replace with your chat ID

	msg := tgbotapi.NewMessage(chatID, message)
	_, err = bot.Send(msg)
	return err
}

// Usage in your soil monitoring system
func CheckSoilMoisture(moisture int) {
	if moisture < 2700 {
		alert := fmt.Sprintf("ALERT: Soil is dry! needs water! Current: %d%%\nWatering pump activated.", moisture)
		if err := SendSoilAlert(alert); err != nil {
			log.Printf("Failed to send Telegram alert: %v", err)
		}
	}
}
