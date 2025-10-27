package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client mqtt.Client
	data   chan SensorData
}

func NewMQTTClient(ctx context.Context, cfg *Config) (*MQTTClient, error) {
	msgChan := make(chan SensorData, 100)

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.MQTTBroker, cfg.MQTTPort)).
		SetClientID(cfg.MQTTClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second)

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v\n", err)
	})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("MQTT connected")

		token := c.Subscribe(cfg.MQTTTopic, 1, createMsgHandler(msgChan))
		token.Wait()
		if token.Error() != nil {
			log.Printf("Subscribe error: %v\n", token.Error())
			return
		}

		log.Printf("Subscribed to: %s\n", cfg.MQTTTopic)
	})

	client := mqtt.NewClient(opts)

	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT connection failed: %v", token.Error())
		return nil, fmt.Errorf("connection failed: %w", token.Error())
	}

	go func() {
		log.Print("Waiting for context to be Cancelled(mqtt goroutine)")
		<-ctx.Done()
		log.Print("Cleaning up mqtt client")
		client.Disconnect(250)
		close(msgChan)
	}()

	return &MQTTClient{client: client, data: msgChan}, nil
}

func createMsgHandler(msgChan chan<- SensorData) mqtt.MessageHandler {
	return func(c mqtt.Client, msg mqtt.Message) {
		var data SensorData

		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Printf("Unmarshalling Error: %v | payload: %s", err, msg.Payload())
			return
		}

		data.Timestamp = time.Now().UnixMilli()

		select {
		case msgChan <- data:
			t := time.UnixMilli(data.Timestamp)
			log.Printf("[%s] temperature : %.2f, humidity: %.2f, Soil Moisture value: %d, Water Pump(ON/OFF): %v", t.Format("2006-01-02 15:05:45"), data.Temperature, data.Humidity, data.SoilMoisture, data.WaterPump)
		default:
			log.Print("Message channel full, dropping message")
		}
	}
}

func (m *MQTTClient) PublishCommand(payload []byte) error {
	if m.client == nil || !m.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	token := m.client.Publish("esp32/actuator", 1, false, payload)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("failed to acticvate pump: %w", token.Error())
	}	

	log.Printf("Published to esp32/actuator: %s", string(payload))
	return nil
}

func (m *MQTTClient) MessageChan() <-chan SensorData {
	return m.data
}
