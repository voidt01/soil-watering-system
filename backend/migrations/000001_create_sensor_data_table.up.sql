CREATE TABLE sensor_data(
    id INT AUTO_INCREMENT PRIMARY KEY,
    temperature float NOT NULL,
    humidity float NOT NULL,
    soil_moisture INT NOT NULL,
    water_pump BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE INDEX idx_time ON sensor_data(created_at);