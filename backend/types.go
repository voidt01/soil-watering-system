package main

type SensorData struct {
	Temperature   float32 `json:"temperature"`
	Humidity      float32 `json:"humidity"`
	Soil_moisture int `json:"soil_moisture"`
}

type MQTTConfig struct {
	Broker   string
	ClientId string
	Topic    string
	Port     int
}
