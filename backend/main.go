package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	mqttCfg := MQTTConfig{
		Broker:   "localhost",
		ClientId: "go-backend-service",
		Topic:    "esp32/sensors",
		Port:     1883,
	}

	client, err := NewMQTTClient(ctx, mqttCfg)
	if err != nil {
		log.Fatalf("Failed to init MQTT: %s", err)
	}

	go func() {
		err = NewHTTPServer(ctx, client, 4000)
    	if err != nil {
        	log.Fatalf("Failed to init HTTP Server: %s", err)
    	}
	}()

	waitForShutdown(cancel)
}

func waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal from OS: %s", sig)

	cancel()
}
