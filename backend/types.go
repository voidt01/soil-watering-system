package main

type SensorData struct {
	Temperature  float32 `json:"temperature"`
	Humidity     float32 `json:"humidity"`
	SoilMoisture int     `json:"soil_moisture"`
	WaterPump    bool    `json:"water_pump"`
	Timestamp    int64   `json:"-"`
}

type MQTTConfig struct {
	Broker   string
	ClientId string
	Topic    string
	Port     int
}