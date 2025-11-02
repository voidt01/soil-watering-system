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
)

type HTTPServer struct {
	MQTTClient *MQTTClient
	Notif      *Notifications
	Database   *Database
}

func NewHTTPServer(ctx context.Context, mqcli *MQTTClient, notif *Notifications, database *Database, cfg *Config) error {
	httpServer := HTTPServer{
		MQTTClient: mqcli,
		Notif:      notif,
		Database:   database,
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: httpServer.routes(),
	}

	errChan := make(chan error, 1)

	go func() {
		log.Print("Waiting for context to be cancelled (http server goroutine)")
		<-ctx.Done()

		log.Print("Shutting down http server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		errChan <- server.Shutdown(shutdownCtx)
	}()

	log.Printf("Starting up http server on port: %d", cfg.HTTPPort)
	err := server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("Error occurred in server: %s", err)
		return err
	}

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
	router.HandleFunc("GET /analytics", hs.analyticsHandler)

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
			hs.Notif.CheckSoilMoisture(data.SoilMoisture, data.Temperature, data.Humidity)

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

func (hs *HTTPServer) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	historicalData, err := hs.Database.GetLast24Hours()
	if err != nil {
		log.Printf("Failed to get historical data: %v", err)
		http.Error(w, "Failed to fetch historical data", http.StatusInternalServerError)
		return
	}

	stats, err := hs.Database.GetStats()
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		http.Error(w, "Failed to fetch statistics", http.StatusInternalServerError)
		return
	}

	var transformedData []HistoricalDataResponse
	for _, data := range historicalData {
		transformedData = append(transformedData, HistoricalDataResponse{
			Time:         data.Time.Format("15:04"),
			Temperature:  data.Temperature,
			Humidity:     data.Humidity,
			SoilMoisture: data.SoilMoisture,
		})
	}

	response := AnalyticsResponse{
		HistoricalData: transformedData,
		Stats:          *stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
