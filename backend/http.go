package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type HTTPServer struct {
	MQTTClient *MQTTClient
	// database (next)
}

func NewHTTPServer(ctx context.Context, mqcli *MQTTClient, port int) error {
	httpServer := HTTPServer{
		MQTTClient: mqcli,
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: httpServer.routes(),
	}

	ErrShutdownChan := make(chan error)

	go func() {
		log.Print("Waiting for context to be Cancelled(http server goroutine)")
		<-ctx.Done()

		log.Print("Shutting down http server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ErrShutdownChan <- server.Shutdown(ctx)
	}()

	log.Printf("Starting up http server on port: %d", port)
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("Error occured in server: %s", err)
		return err
	}

	err = <-ErrShutdownChan
	if err != nil {
		log.Printf("Error occured when shutting down the server: %s", err)
		return err
	}

	log.Print("Successfully shutting down the server")

	return nil
}

func (hs *HTTPServer) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/data-streams", hs.sseHandler)

	return router
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
			// writing to db 

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
