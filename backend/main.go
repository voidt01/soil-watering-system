package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    
    client := newMQTTClient()

    token := client.Subscribe(topic, 1, nil)
    token.Wait()
    log.Printf("Subscribed to topic: %s\n", topic)

    // Process messages
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        processedMsgChan := processMsg(ctx, mqttmsgchan)
        for range processedMsgChan {
            // For now, just consume; later pass to DB or frontend
        }
    }()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan

	cancel()

    log.Println("\nShutting down gracefully...")
    client.Unsubscribe(topic)
    client.Disconnect(250)

    wg.Wait()
    log.Println("Shutdown complete")
}
