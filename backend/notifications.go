package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	alertMutex              sync.Mutex
	lastDryAlertTime        time.Time
	lastWetAlertTime        time.Time
	lastPreventiveAlertTime time.Time
	alertCooldown           = 30 * time.Minute
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

func (nf *Notifications) CheckSoilMoisture(moisture int, temperature float32, humidity float32) {

	if moisture > 3000 {
		alertMutex.Lock()
		defer alertMutex.Unlock()

		if time.Since(lastDryAlertTime) < alertCooldown {
			return
		}

		alert := fmt.Sprintf(
			"ALERT: Soil is DRY!\nCurrent moisture: %d\nWater pump is activated automatically.",
			moisture,
		)

		if err := nf.SendSoilAlert(alert); err != nil {
			log.Printf("Failed to send dry soil alert: %v", err)
		} else {
			lastDryAlertTime = time.Now()
			log.Printf("Dry soil alert sent: %d", moisture)
		}

		return
	}

	if moisture < 2800 {
		alertMutex.Lock()
		defer alertMutex.Unlock()

		if time.Since(lastWetAlertTime) < alertCooldown {
			return
		}

		alert := fmt.Sprintf(
			"Soil moisture is normal.\nCurrent moisture: %d\nNo watering needed.",
			moisture,
		)

		if err := nf.SendSoilAlert(alert); err != nil {
			log.Printf("Failed to send wet soil notice: %v", err)
		} else {
			lastWetAlertTime = time.Now()
			log.Printf("Wet/optimal soil notice sent: %d", moisture)
		}

		return
	}

	if moisture > 2800 && moisture < 3000 && temperature >= 32.0 && humidity < 40.0 {
		alertMutex.Lock()
		defer alertMutex.Unlock()

		if time.Since(lastPreventiveAlertTime) < alertCooldown {
			return
		}

		alert := fmt.Sprintf(
			"Soil moisture is between dry and wet + the environmental situation is a bit hot.\nCurrent moisture: %d\nwatering needed.",
			moisture,
		)

		if err := nf.SendSoilAlert(alert); err != nil {
			log.Printf("Failed to send wet soil notice: %v", err)
		} else {
			lastWetAlertTime = time.Now()
			log.Printf("Wet/optimal soil notice sent: %d", moisture)
		}

		return
	}
}
