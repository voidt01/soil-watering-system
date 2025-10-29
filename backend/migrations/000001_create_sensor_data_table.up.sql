CREATE TABLE sensor_data(
    id SERIAL PRIMARY KEY,
    temperature REAL NOT NULL,
    humidity REAL NOT NULL,
    soil_moisture INTEGER NOT NULL,
    water_pump BOOLEAN NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_time ON sensor_data(created_at);