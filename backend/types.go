package main

import "time"

type SensorData struct {
	Temperature  float32   `json:"temperature"`
	Humidity     float32   `json:"humidity"`
	SoilMoisture int       `json:"soil_moisture"`
	WaterPump    bool      `json:"water_pump"`
	Timestamp    time.Time `json:"-"`
}

type HistoricalData struct {
	Temperature  float32   
	Humidity     float32   
	SoilMoisture int       
	WaterPump    bool      
	CreatedAt    time.Time 
}

type HistoricalDataResponse struct {
	Time         string  `json:"time"`
	Temperature  float32 `json:"temperature"`
	Humidity     float32 `json:"humidity"`
	SoilMoisture int     `json:"soil_moisture"`
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