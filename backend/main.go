package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env instead")
	}

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %s", err)
	}

	db, err := OpenDB("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s", err)
	}
	log.Println("Connected to database successfully")
	defer db.Close()

	database := NewDatabase(db)

	mqClient, err := NewMQTTClient(ctx, database, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %s", err)
	}

	notif, err := newNotification(cfg)
	if err != nil {
		log.Fatalf("Failed to init bot: %s", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = NewHTTPServer(ctx, mqClient, notif, cfg.HTTPPort)
		if err != nil {
			log.Fatalf("Failed to init HTTP Server: %s", err)
		}
	}()

	waitForShutdown(cancel)

	wg.Wait()
	log.Print("Application shutdown")
}

func waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal from OS: %s", sig)

	cancel()
}

func OpenDB(driver, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}