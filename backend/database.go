package main

import (
	"database/sql"
	"fmt"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(DB *sql.DB) *Database {
	return &Database{DB: DB}
}

func (d *Database) InsertSensorData(data SensorData) error {
	query := `
		INSERT INTO sensor_data(temperature, humidity, soil_moisture, water_pump, created_at)
		VALUES(?, ?, ?, ?, ?)
	`

	_, err := d.DB.Exec(query, 
		data.Temperature,
		data.Humidity,
		data.SoilMoisture,
		data.WaterPump,
		data.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert to sensor data: %w", err)
	}

	return nil
}