import React, { useState, useEffect } from 'react';
import { Droplet, Wind, Thermometer, Zap, Home, BarChart3, Activity, Wifi, WifiOff } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, AreaChart, Area } from 'recharts';

export default function IoTDashboard() {
  const [currentPage, setCurrentPage] = useState('home');
  const [sensorData, setSensorData] = useState({
    temperature: 0,
    humidity: 0,
    soil_moisture: 0,
    water_pump: false,
  });
  const [mqttConnected, setMqttConnected] = useState(false);
  const [historicalData, setHistoricalData] = useState([]);
  const [stats, setStats] = useState({
    avgTemp: 0,
    avgHumidity: 0,
    avgMoisture: 0,
    pumpActivations: 0,
  });
  const [isPublishing, setIsPublishing] = useState(false);
  const [isLoadingAnalytics, setIsLoadingAnalytics] = useState(false);

  useEffect(() => {
    let eventSource = null;
    let reconnectTimeout = null;
    
    const connectSSE = () => {
      if (eventSource) {
        eventSource.close();
      }
      
      const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000';
      eventSource = new EventSource(`${API_URL}/data-streams`);

      eventSource.onopen = () => {
        setMqttConnected(true);
        console.log('Connected to SSE stream');
      };

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          // Validate data
          if (typeof data.temperature === 'number' && 
              typeof data.humidity === 'number' && 
              typeof data.soil_moisture === 'number') {
            setSensorData(data);
          }
        } catch (err) {
          console.error('Failed to parse SSE data:', err);
        }
      };

      eventSource.onerror = (error) => {
        console.error('SSE error:', error);
        setMqttConnected(false);
        eventSource.close();
        
        reconnectTimeout = setTimeout(connectSSE, 3000);
      };
    };

    connectSSE();

    return () => {
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (eventSource) {
        eventSource.close();
      }
    };
  }, []); 

  useEffect(() => {
    if (currentPage === 'analytics') {
      fetchAnalytics();
    }
  }, [currentPage]);

  const connectSSE = () => {
    const eventSource = new EventSource('http://localhost:4000/data-streams');

    eventSource.onopen = () => {
      setMqttConnected(true);
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
      setMqttConnected(false);
      eventSource.close();
      setTimeout(connectSSE, 3000);
    };

    return () => eventSource.close();
  };

  const fetchAnalytics = async () => {
    setIsLoadingAnalytics(true);
    try {
      const response = await fetch('http://localhost:4000/analytics');
      if (!response.ok) throw new Error('Failed to fetch analytics');
      
      const data = await response.json();
      setHistoricalData(data.historical_data || []); 

      const newStats = data.stats
        ? {
            avgTemp: data.stats.avg_temp,
            avgHumidity: data.stats.avg_humidity,
            avgMoisture: data.stats.avg_moisture,
            pumpActivations: data.stats.pump_activations,
          }
        : {
            avgTemp: 0,
            avgHumidity: 0,
            avgMoisture: 0,
            pumpActivations: 0,
          };

      setStats(newStats);
    } catch (err) {
      console.error('Error fetching analytics:', err);
    } finally {
      setIsLoadingAnalytics(false);
    }
  };

  const toggleWaterPump = async () => {
    setIsPublishing(true);
    try {
      const payload = { water_pump: !sensorData.water_pump };
      const response = await fetch('http://localhost:4000/actuator', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    } catch (err) {
      console.error('Error:', err);
    } finally {
      setIsPublishing(false);
    }
  };

  // Home Page
  const HomePage = () => (
    <div className="min-h-screen bg-gradient-to-br from-amber-50 via-white to-stone-100">
      <div className="max-w-6xl mx-auto px-8 py-16">
        <div className="text-center mb-16">
          <div className="inline-block mb-6">
            <div className="w-24 h-24 bg-gradient-to-br from-amber-600 to-amber-800 rounded-3xl flex items-center justify-center shadow-xl">
              <Droplet className="w-12 h-12 text-white" />
            </div>
          </div>
          <h1 className="text-6xl font-bold text-stone-800 mb-4">
            Smart Soil Monitor
          </h1>
          <p className="text-xl text-stone-600 mb-8">
            Intelligent soil moisture management for optimal plant growth
          </p>
          <button
            onClick={() => setCurrentPage('dashboard')}
            className="bg-gradient-to-r from-amber-600 to-amber-700 hover:from-amber-700 hover:to-amber-800 text-white px-8 py-4 rounded-xl font-semibold text-lg shadow-lg transition-all duration-300 transform hover:scale-105"
          >
            Open Dashboard
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-20">
          <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200">
            <div className="w-14 h-14 bg-amber-100 rounded-xl flex items-center justify-center mb-4">
              <Activity className="w-7 h-7 text-amber-700" />
            </div>
            <h3 className="text-xl font-bold text-stone-800 mb-2">Real-time Monitoring</h3>
            <p className="text-stone-600">Track temperature, humidity, and soil moisture in real-time</p>
          </div>
          
          <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200">
            <div className="w-14 h-14 bg-amber-100 rounded-xl flex items-center justify-center mb-4">
              <Zap className="w-7 h-7 text-amber-700" />
            </div>
            <h3 className="text-xl font-bold text-stone-800 mb-2">Automated Control</h3>
            <p className="text-stone-600">Control water pump remotely with instant response</p>
          </div>
          
          <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200">
            <div className="w-14 h-14 bg-amber-100 rounded-xl flex items-center justify-center mb-4">
              <BarChart3 className="w-7 h-7 text-amber-700" />
            </div>
            <h3 className="text-xl font-bold text-stone-800 mb-2">Data Analytics</h3>
            <p className="text-stone-600">Analyze trends and patterns for better decisions</p>
          </div>
        </div>
      </div>
    </div>
  );

  // Dashboard Page
  const DashboardPage = () => (
    <div className="min-h-screen bg-gradient-to-br from-amber-50 via-white to-stone-100 p-8">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-4xl font-bold text-stone-800 mb-2">Dashboard</h1>
            <p className="text-stone-600">Real-time sensor monitoring</p>
          </div>
          <div className={`flex items-center gap-2 px-4 py-2 rounded-lg ${mqttConnected ? 'bg-green-100' : 'bg-red-100'}`}>
            {mqttConnected ? <Wifi className="w-5 h-5 text-green-700" /> : <WifiOff className="w-5 h-5 text-red-700" />}
            <span className={`font-semibold ${mqttConnected ? 'text-green-700' : 'text-red-700'}`}>
              {mqttConnected ? 'MQTT Connected' : 'MQTT Disconnected'}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
            <div className="flex items-center justify-between mb-4">
              <p className="text-stone-600 font-medium">Temperature</p>
              <Thermometer className="w-8 h-8 text-amber-600" />
            </div>
            <p className="text-4xl font-bold text-stone-800">
              {sensorData.temperature.toFixed(1)}°C
            </p>
            <p className="text-sm text-stone-500 mt-2">Optimal range: 20-30°C</p>
          </div>

          <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
            <div className="flex items-center justify-between mb-4">
              <p className="text-stone-600 font-medium">Humidity</p>
              <Wind className="w-8 h-8 text-blue-600" />
            </div>
            <p className="text-4xl font-bold text-stone-800">
              {sensorData.humidity.toFixed(1)}%
            </p>
            <p className="text-sm text-stone-500 mt-2">Optimal range: 40-60%</p>
          </div>

          <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
            <div className="flex items-center justify-between mb-4">
              <p className="text-stone-600 font-medium">Soil Moisture</p>
              <Droplet className="w-8 h-8 text-cyan-600" />
            </div>
            <p className="text-4xl font-bold text-stone-800">
              {sensorData.soil_moisture}
            </p>
            <p className="text-sm text-stone-500 mt-2">Lower = More moisture</p>
          </div>

          <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
            <div className="flex items-center justify-between mb-4">
              <p className="text-stone-600 font-medium">Water Pump</p>
              <Zap className={`w-8 h-8 ${sensorData.water_pump ? 'text-green-600' : 'text-stone-400'}`} />
            </div>
            <p className={`text-3xl font-bold ${sensorData.water_pump ? 'text-green-600' : 'text-stone-500'}`}>
              {sensorData.water_pump ? '● RUNNING' : '○ IDLE'}
            </p>
            <p className="text-sm text-stone-500 mt-2">Status: {sensorData.water_pump ? 'Active' : 'Standby'}</p>
          </div>
        </div>

        <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200">
          <h2 className="text-2xl font-bold text-stone-800 mb-6">Water Pump Control</h2>
          
          <button
            onClick={toggleWaterPump}
            disabled={isPublishing || !mqttConnected}
            className={`w-full py-4 px-6 rounded-xl font-bold text-lg transition-all duration-300 ${
              sensorData.water_pump
                ? 'bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 text-white shadow-lg'
                : 'bg-gradient-to-r from-amber-600 to-amber-700 hover:from-amber-700 hover:to-amber-800 text-white shadow-lg'
            } disabled:opacity-50 disabled:cursor-not-allowed transform hover:scale-105`}
          >
            {isPublishing ? 'Sending Command...' : sensorData.water_pump ? 'Turn OFF Pump' : 'Turn ON Pump'}
          </button>

          <div className="mt-4 text-center">
            <p className="text-stone-600">
              Current state: <span className="font-semibold text-stone-800">{sensorData.water_pump ? 'Pump is running' : 'Pump is idle'}</span>
            </p>
          </div>
        </div>
      </div>
    </div>
  );

  // Analytics Page
  const AnalyticsPage = () => (
    <div className="min-h-screen bg-gradient-to-br from-amber-50 via-white to-stone-100 p-8">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-4xl font-bold text-stone-800 mb-2">Analytics</h1>
            <p className="text-stone-600">Historical data and trend analysis (Last 24 hours)</p>
          </div>
          <button
            onClick={fetchAnalytics}
            disabled={isLoadingAnalytics}
            className="bg-gradient-to-r from-amber-600 to-amber-700 hover:from-amber-700 hover:to-amber-800 text-white px-6 py-3 rounded-lg font-semibold shadow-lg transition-all duration-300 transform hover:scale-105 disabled:opacity-50"
          >
            {isLoadingAnalytics ? 'Loading...' : 'Refresh Data'}
          </button>
        </div>

        {isLoadingAnalytics ? (
          <div className="flex items-center justify-center h-64">
            <div className="text-stone-600 text-xl">Loading analytics data...</div>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
                <p className="text-stone-600 font-medium mb-2">Avg Temperature</p>
                <p className="text-3xl font-bold text-stone-800">{stats.avgTemp.toFixed(1)}°C</p>
              </div>
              <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
                <p className="text-stone-600 font-medium mb-2">Avg Humidity</p>
                <p className="text-3xl font-bold text-stone-800">{stats.avgHumidity.toFixed(1)}%</p>
              </div>
              <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
                <p className="text-stone-600 font-medium mb-2">Avg Soil Moisture</p>
                <p className="text-3xl font-bold text-stone-800">{stats.avgMoisture.toFixed(0)}</p>
              </div>
              <div className="bg-white rounded-2xl p-6 shadow-lg border border-stone-200">
                <p className="text-stone-600 font-medium mb-2">Pump Activations</p>
                <p className="text-3xl font-bold text-stone-800">{stats.pumpActivations}</p>
              </div>

            </div>

            {historicalData.length === 0 ? (
              <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200 text-center">
                <p className="text-stone-600 text-lg">No historical data available yet. Start collecting data to see analytics.</p>
              </div>
            ) : (
              <>
                <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200 mb-6">
                  <h2 className="text-2xl font-bold text-stone-800 mb-6">Temperature & Humidity Trends</h2>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={historicalData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e7e5e4" />
                      <XAxis dataKey="time" stroke="#78716c" />
                      <YAxis stroke="#78716c" />
                      <Tooltip contentStyle={{ backgroundColor: '#fff', border: '1px solid #d6d3d1', borderRadius: '8px' }} />
                      <Legend />
                      <Line type="monotone" dataKey="temperature" stroke="#d97706" strokeWidth={2} name="Temperature (°C)" />
                      <Line type="monotone" dataKey="humidity" stroke="#0284c7" strokeWidth={2} name="Humidity (%)" />
                    </LineChart>
                  </ResponsiveContainer>
                </div>

                <div className="bg-white rounded-2xl p-8 shadow-lg border border-stone-200">
                  <h2 className="text-2xl font-bold text-stone-800 mb-6">Soil Moisture Level</h2>
                  <ResponsiveContainer width="100%" height={300}>
                    <AreaChart data={historicalData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#e7e5e4" />
                      <XAxis dataKey="time" stroke="#78716c" />
                      <YAxis stroke="#78716c" />
                      <Tooltip contentStyle={{ backgroundColor: '#fff', border: '1px solid #d6d3d1', borderRadius: '8px' }} />
                      <Area type="monotone" dataKey="soil_moisture" stroke="#0891b2" fill="#67e8f9" fillOpacity={0.6} name="Soil Moisture" />
                    </AreaChart>
                  </ResponsiveContainer>
                  <p className="text-sm text-stone-600 mt-4">* Lower values indicate higher moisture content in the soil</p>
                </div>
              </>
            )}
          </>
        )}
      </div>
    </div>
  );

  // Navigation
  const Navigation = () => (
    <nav className="bg-white border-b border-stone-200 shadow-sm">
      <div className="max-w-7xl mx-auto px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-amber-600 to-amber-800 rounded-lg flex items-center justify-center">
              <Droplet className="w-6 h-6 text-white" />
            </div>
            <span className="text-xl font-bold text-stone-800">Smart Soil</span>
          </div>
          
          <div className="flex gap-2">
            <button
              onClick={() => setCurrentPage('home')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all ${
                currentPage === 'home' 
                  ? 'bg-amber-100 text-amber-700' 
                  : 'text-stone-600 hover:bg-stone-100'
              }`}
            >
              <Home className="w-5 h-5" />
              Home
            </button>
            <button
              onClick={() => setCurrentPage('dashboard')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all ${
                currentPage === 'dashboard' 
                  ? 'bg-amber-100 text-amber-700' 
                  : 'text-stone-600 hover:bg-stone-100'
              }`}
            >
              <Activity className="w-5 h-5" />
              Dashboard
            </button>
            <button
              onClick={() => setCurrentPage('analytics')}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-all ${
                currentPage === 'analytics' 
                  ? 'bg-amber-100 text-amber-700' 
                  : 'text-stone-600 hover:bg-stone-100'
              }`}
            >
              <BarChart3 className="w-5 h-5" />
              Analytics
            </button>
          </div>
        </div>
      </div>
    </nav>
  );

  return (
    <div className="min-h-screen">
      <Navigation />
      {currentPage === 'home' && <HomePage />}
      {currentPage === 'dashboard' && <DashboardPage />}
      {currentPage === 'analytics' && <AnalyticsPage />}
    </div>
  );
}