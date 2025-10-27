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

type Notifications struct {
	bot *tgbotapi.BotAPI
	cfg *Config
}

func newNotification(cfg *Config) (*Notifications, error) {
	notif := &Notifications{
		cfg: cfg,
	}
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to init Bot")
	}
	notif.bot = bot
	
	return notif, nil
}

func (nf *Notifications) SendSoilAlert(message string) error {
	chatID := int64(nf.cfg.TelegramChatID) 

	msg := tgbotapi.NewMessage(chatID, message)
	_, err := nf.bot.Send(msg)
	return err
}

// Usage in your soil monitoring system
func (nf *Notifications) CheckSoilMoisture(moisture int) {
	if moisture < 2700 {
		alertMutex.Lock()
		defer alertMutex.Unlock()

		if time.Since(lastAlertTime) < alertCooldown {
			return
		}

		alert := fmt.Sprintf("ALERT: Soil is dry! needs water! Current: %d%%\nWatering pump activated.", moisture)
		if err := nf.SendSoilAlert(alert); err != nil {
			log.Printf("Failed to send Telegram alert: %v", err)
		} else {
			lastAlertTime = time.Now()
			log.Printf("Soil moisture alert sent: %d", moisture)
		}
	}
}
