package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HTTPServer struct {
	MQTTClient *MQTTClient
	TeleBot *tgbotapi.BotAPI
}

func NewHTTPServer(ctx context.Context, mqcli *MQTTClient, bot *tgbotapi.BotAPI, port int) error {
	httpServer := HTTPServer{
		MQTTClient: mqcli,
		TeleBot: bot,
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: httpServer.routes(),
	}

	errChan := make(chan error, 1)

	// Goroutine to handle graceful shutdown
	go func() {
		log.Print("Waiting for context to be cancelled (http server goroutine)")
		<-ctx.Done()

		log.Print("Shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		errChan <- server.Shutdown(shutdownCtx)
	}()

	log.Printf("Starting up http server on port: %d", port)
	err := server.ListenAndServe()
	
	// If the server closed due to shutdown (not an actual error), continue
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("Error occurred in server: %s", err)
		return err
	}

	// Wait for the shutdown goroutine to complete
	shutdownErr := <-errChan
	if shutdownErr != nil {
		log.Printf("Error occurred when shutting down the server: %s", shutdownErr)
		return shutdownErr
	}

	log.Print("Successfully shut down the server")
	return nil
}

func (hs *HTTPServer) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/data-streams", hs.sseHandler)
	router.HandleFunc("POST /actuator", hs.actuatorHandler)

	return hs.corsMiddleware(router)
}

func (hs *HTTPServer) sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientGone := r.Context().Done()

	for {
		select {
		case <-clientGone:
			log.Print("Client disconnected")
			return
		case data := <-hs.MQTTClient.MessageChan():
			CheckSoilMoisture(hs.TeleBot, data.SoilMoisture)

			dataByte, err := json.Marshal(data)
			if err != nil {
				log.Printf("failed to marshal data to JSON: %v", err)
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", dataByte)
			w.(http.Flusher).Flush()
		}
	}
}

func (hs *HTTPServer) actuatorHandler(w http.ResponseWriter, r *http.Request) {
	dataByte, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ok := json.Valid(dataByte)
	if !ok {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = hs.MQTTClient.PublishCommand(dataByte)
	if err != nil {
		log.Printf("MQTT Publish error: %v", err)
		http.Error(w, "Failed to send command to device", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"message":"Command sent successfully"}`))
}

func (hs *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}