import React, { useState, useEffect } from 'react';
import { Droplet, Wind, Thermometer, Zap } from 'lucide-react';

export default function SoilWateringSystem() {
  const [sensorData, setSensorData] = useState({
    temperature: 0,
    humidity: 0,
    soil_moisture: 0,
    water_pump: false,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [publishError, setPublishError] = useState(null);
  const [isPublishing, setIsPublishing] = useState(false);

  useEffect(() => {
    connectSSE();
  }, []);

  const connectSSE = () => {
    const eventSource = new EventSource('http://localhost:4000/data-streams');

    eventSource.onopen = () => {
      setError(null);
      setLoading(false);
      console.log('Connected to SSE stream');
    };

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setSensorData(data);
      } catch (err) {
        console.error('Failed to parse SSE data:', err);
      }
    };

    eventSource.onerror = () => {
      setError('Connection lost. Reconnecting...');
      eventSource.close();
      setTimeout(connectSSE, 3000);
    };

    return () => eventSource.close();
  };

  const toggleWaterPump = async () => {
    setIsPublishing(true);
    setPublishError(null);

    try {
      const payload = {
        water_pump: !sensorData.water_pump,
      };

      const response = await fetch('http://localhost:4000/actuator', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const result = await response.json();
      console.log('Command sent:', result);
      // Note: The actual state will update when the ESP32 publishes the new value via MQTT
    } catch (err) {
      setPublishError(`Failed to send command: ${err.message}`);
      console.error('Error publishing command:', err);
    } finally {
      setIsPublishing(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 to-slate-800 p-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-12">
          <h1 className="text-4xl font-bold text-white mb-2">🌱 Soil Watering System</h1>
          <p className="text-slate-400">Real-time monitoring and control</p>
        </div>

        {/* Status Indicator */}
        <div className="mb-8">
          {loading && (
            <div className="bg-blue-900/30 border border-blue-500 rounded-lg p-4">
              <p className="text-blue-300">Connecting to sensor stream...</p>
            </div>
          )}
          {error && (
            <div className="bg-yellow-900/30 border border-yellow-500 rounded-lg p-4">
              <p className="text-yellow-300">{error}</p>
            </div>
          )}
        </div>

        {/* Main Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
          {/* Temperature Card */}
          <div className="bg-slate-700/50 backdrop-blur border border-slate-600 rounded-lg p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm mb-2">Temperature</p>
                <p className="text-4xl font-bold text-white">
                  {sensorData.temperature.toFixed(1)}°C
                </p>
              </div>
              <Thermometer className="w-12 h-12 text-red-400" />
            </div>
          </div>

          {/* Humidity Card */}
          <div className="bg-slate-700/50 backdrop-blur border border-slate-600 rounded-lg p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm mb-2">Humidity</p>
                <p className="text-4xl font-bold text-white">
                  {sensorData.humidity.toFixed(1)}%
                </p>
              </div>
              <Wind className="w-12 h-12 text-blue-400" />
            </div>
          </div>

          {/* Soil Moisture Card */}
          <div className="bg-slate-700/50 backdrop-blur border border-slate-600 rounded-lg p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm mb-2">Soil Moisture</p>
                <p className="text-4xl font-bold text-white">
                  {sensorData.soil_moisture}%
                </p>
              </div>
              <Droplet className="w-12 h-12 text-cyan-400" />
            </div>
          </div>

          {/* Water Pump Status Card */}
          <div className="bg-slate-700/50 backdrop-blur border border-slate-600 rounded-lg p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-400 text-sm mb-2">Water Pump</p>
                <p className={`text-2xl font-bold ${sensorData.water_pump ? 'text-green-400' : 'text-slate-400'}`}>
                  {sensorData.water_pump ? '🟢 ON' : '⚪ OFF'}
                </p>
              </div>
              <Zap className={`w-12 h-12 ${sensorData.water_pump ? 'text-green-400' : 'text-slate-500'}`} />
            </div>
          </div>
        </div>

        {/* Control Section */}
        <div className="bg-slate-700/50 backdrop-blur border border-slate-600 rounded-lg p-8">
          <h2 className="text-xl font-bold text-white mb-6">Water Pump Control</h2>
          
          {publishError && (
            <div className="bg-red-900/30 border border-red-500 rounded-lg p-4 mb-4">
              <p className="text-red-300">{publishError}</p>
            </div>
          )}

          <button
            onClick={toggleWaterPump}
            disabled={isPublishing}
            className={`w-full py-4 px-6 rounded-lg font-bold text-lg transition-all duration-200 ${
              sensorData.water_pump
                ? 'bg-red-600 hover:bg-red-700 text-white'
                : 'bg-green-600 hover:bg-green-700 text-white'
            } disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {isPublishing ? 'Sending...' : sensorData.water_pump ? 'Turn OFF' : 'Turn ON'}
          </button>

          <p className="text-slate-400 text-sm mt-4">
            Current pump state: <span className="font-semibold text-white">{sensorData.water_pump ? 'Running' : 'Idle'}</span>
          </p>
        </div>
      </div>
    </div>
  );
}