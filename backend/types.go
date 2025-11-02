package main

import "time"

type HistoricalData struct {
	Time         time.Time
	Temperature  float32
	Humidity     float32
	SoilMoisture float32
}

type SensorData struct {
	Temperature  float32   `json:"temperature"`
	Humidity     float32   `json:"humidity"`
	SoilMoisture int       `json:"soil_moisture"`
	WaterPump    bool      `json:"water_pump"`
	Timestamp    time.Time `json:"-"`
}

type HistoricalDataResponse struct {
	Time         string  `json:"time"`
	Temperature  float32 `json:"temperature"`
	Humidity     float32 `json:"humidity"`
	SoilMoisture float32 `json:"soil_moisture"`
}

type Stats struct {
	AvgTemp         float64 `json:"avg_temp"`
	AvgHumidity     float64 `json:"avg_humidity"`
	AvgMoisture     float64 `json:"avg_moisture"`
	PumpActivations int     `json:"pump_activations"`
}

type AnalyticsResponse struct {
	HistoricalData []HistoricalDataResponse `json:"historical_data"`
	Stats          Stats                    `json:"stats"`
}
