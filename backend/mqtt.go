package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker = "tcp://localhost:1883"
	clientId = "go-backend-service"
	topic = "esp32/sensors"
)

type Message struct {
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	SoilValue   int     `json:"soil_moisture"`
}


func newMQTTClient() mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Error connecting to MQTT broker: %s", token.Error())
	}

	return client
}

var mqttmsgchan = make(chan mqtt.Message)

var messagePubHandler mqtt.MessageHandler = func(c mqtt.Client, m mqtt.Message) {
	mqttmsgchan <- m
}

var connectHandler mqtt.OnConnectHandler = func(c mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(c mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err)
}

func processMsg(ctx context.Context, input <-chan mqtt.Message) chan Message {
	out := make(chan Message)

	go func() {
		defer close(out)
		for {
			select {
			case msg, ok := <-input:
				if !ok {
					return
				}
				fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

				var IotMsg Message
				err := json.Unmarshal(msg.Payload(), &IotMsg)
				if err != nil {
					log.Print("Error unmarshalling message\n")
				}
				
				out <- IotMsg
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

