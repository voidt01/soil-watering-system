package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client   mqtt.Client
	data     chan SensorData
	database *Database
}

func NewMQTTClient(ctx context.Context, database *Database, cfg *Config) (*MQTTClient, error) {
	msgChan := make(chan SensorData, 100)

	brokerURL := fmt.Sprintf("tcp://%s:%d", cfg.MQTTBroker, cfg.MQTTPort)
	if cfg.MQTTUseTLS {
		brokerURL = fmt.Sprintf("ssl://%s:%d", cfg.MQTTBroker, cfg.MQTTPort)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(cfg.MQTTClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second)

	// Configure TLS if enabled
	if cfg.MQTTUseTLS {
		tlsConfig, err := newTLSConfig(cfg.MQTTCAFile, cfg.MQTTCertFile, cfg.MQTTKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
		log.Println("TLS enabled for MQTT connection")
	}

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v\n", err)
	})

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("MQTT connected")

		token := c.Subscribe(cfg.MQTTTopic, 1, createMsgHandler(msgChan, database))
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
		log.Print("Waiting for context to be Cancelled (mqtt goroutine)")
		<-ctx.Done()
		log.Print("Cleaning up mqtt client")
		client.Disconnect(250)
		close(msgChan)
	}()

	return &MQTTClient{client: client, database: database, data: msgChan}, nil
}

// newTLSConfig creates a TLS configuration for mutual TLS authentication
func newTLSConfig(caFile, certFile, keyFile string) (*tls.Config, error) {
	// Load CA certificate
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

func createMsgHandler(msgChan chan<- SensorData, database *Database) mqtt.MessageHandler {
	return func(c mqtt.Client, msg mqtt.Message) {
		var data SensorData

		err := json.Unmarshal(msg.Payload(), &data)
		if err != nil {
			log.Printf("Unmarshalling Error: %v | payload: %s", err, msg.Payload())
			return
		}

		data.Timestamp = time.Now()

		if err := database.InsertSensorData(data); err != nil {
			log.Printf("Failed to save sensor data to DB: %s", err)
		} else {
			log.Printf("Data saved to DB successfully")
		}

		select {
		case msgChan <- data:
			log.Printf("temperature: %.2f, humidity: %.2f, Soil Moisture: %d, Water Pump: %v",
				data.Temperature, data.Humidity, data.SoilMoisture, data.WaterPump)
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
		return fmt.Errorf("failed to activate pump: %w", token.Error())
	}

	log.Printf("Published to esp32/actuator: %s", string(payload))
	return nil
}

func (m *MQTTClient) MessageChan() <-chan SensorData {
	return m.data
}