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
		VALUES($1, $2, $3, $4, $5)
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

func (d *Database) GetLast24Hours() ([]HistoricalData, error) {
	query := `
        SELECT 
            DATE_TRUNC('minute', created_at) AS time,
            AVG(temperature) AS temperature,
            AVG(humidity) AS humidity,
            AVG(soil_moisture) AS soil_moisture
        FROM sensor_data
        WHERE created_at >= NOW() - INTERVAL '24 hours'
        GROUP BY time
        ORDER BY time ASC;
    `

	rows, err := d.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query historical data: %w", err)
	}
	defer rows.Close()

	var data []HistoricalData

	for rows.Next() {
		var h HistoricalData
		err := rows.Scan(
			&h.Time,
			&h.Temperature,
			&h.Humidity,
			&h.SoilMoisture,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		data = append(data, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return data, nil
}

func (d *Database) GetStats() (*Stats, error) {
	query := `
		SELECT 
			COALESCE(AVG(temperature), 0) as avg_temp,
			COALESCE(AVG(humidity), 0) as avg_humidity,
			COALESCE(AVG(soil_moisture), 0) as avg_moisture,
			COALESCE(SUM(CASE WHEN water_pump = true THEN 1 ELSE 0 END), 0) as pump_activations
		FROM sensor_data
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`

	var stats Stats
	err := d.DB.QueryRow(query).Scan(
		&stats.AvgTemp,
		&stats.AvgHumidity,
		&stats.AvgMoisture,
		&stats.PumpActivations,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &stats, nil
}
