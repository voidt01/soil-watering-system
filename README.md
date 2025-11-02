# Soil Watering System

An IoT-based automatic watering system that monitors soil moisture and controls a water pump in real time. Sensor data is sent through MQTT, processed by a backend service, and displayed on a web dashboard for monitoring and device control.

---

## Description

This project is built to maintain optimal soil moisture using a soil sensor connected to an ESP32 microcontroller.  
The device publishes sensor data to an MQTT broker, the backend handles automation logic and data storage, and the frontend displays live and historical readings.  
Users can control the water pump manually or let the system run in automatic mode based on sensor thresholds.

---

## Features

- Real-time soil moisture monitoring  
- Automatic watering based on sensor readings  
- Manual pump control through the web dashboard  
- Device communication using Mosquitto MQTT Broker  
- Live and historical data visualization  
- Deployment via Docker Compose  

---

## Backend Technology (Summary)

| Component        | Technology Used |
|------------------|-----------------|
| Language         | Go (Golang) |
| Device Messaging | MQTT (Mosquitto) |
| Database         | PostgreSQL |
| Real-Time Stream | Server-Sent Events (SSE) |
| Notifications    | Telegram Bot |
| Deployment       | Docker Compose |

---

## System Flow (High Level)

1. ESP32 reads soil moisture, temperature, and humidity.
2. Sensor data is published to the MQTT broker.
3. Backend receives data, applies watering logic, and stores it in the database.
4. Dashboard receives live updates via SSE and displays charts.
5. Users can send pump control commands, which are sent back to the device via MQTT.

