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

func NewMQTTClient(ctx context.Context, cfg MQTTConfig) (*MQTTClient, error) {
	msgChan := make(chan SensorData, 100)

	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Broker, cfg.Port)).
		SetClientID(cfg.ClientId).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second)

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v\n", err)
	})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("MQTT connected")

		token := c.Subscribe(cfg.Topic, 1, createMsgHandler(msgChan))
		token.Wait()
		if token.Error() != nil {
			log.Printf("Subscribe error: %v\n", token.Error())
			return
		}

		log.Printf("Subscribed to: %s\n", cfg.Topic)
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
		}

		select {
		case msgChan <- data:
			log.Printf("Payload successfully sent | temperature : %.2f, humidity: %.2f, Soil Moisture value: %d", data.Temperature, data.Humidity, data.Soil_moisture)
		default:
			log.Print("Message channel full, dropping message")
		}
	}
}

func (m *MQTTClient) MessageChan() <-chan SensorData {
	return m.data
}
