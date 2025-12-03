import { useState, useEffect } from 'react'
import './App.css'

// URL da sua API no Render
const API_URL = "https://pqc-api.onrender.com/api/v1/stats";

function App() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(new Date());

  const fetchData = async () => {
    try {
      const response = await fetch(API_URL);
      const json = await response.json();
      setData(json);
      setLastUpdate(new Date());
    } catch (error) {
      console.error("Erro ao buscar dados:", error);
    } finally {
      setLoading(false);
    }
  };

  // Atualiza a cada 5 segundos (Live polling)
  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  if (loading) return <div className="loading">INITIALIZING QUANTUM LINK...</div>;

  return (
    <div className="dashboard-container">
      <header className="header">
        <div className="logo">🛡️ PQC SHIELD <span className="tag">ENTERPRISE</span></div>
        <div className="status-indicator">
          SYSTEM STATUS: <span className="online">ONLINE</span>
          <div className="pulse"></div>
        </div>
      </header>

      <main className="main-content">
        <div className="stats-grid">
          <div className="card">
            <h3>ACTIVE PROXIES</h3>
            <div className="big-number">{data?.total_proxies || 0}</div>
          </div>
          <div className="card">
            <h3>TOTAL ENCRYPTED TRAFFIC</h3>
            <div className="big-number">-- TB</div>
            <div className="sub-text">Kyber-768 Protected</div>
          </div>
          <div className="card">
            <h3>THREATS MITIGATED</h3>
            <div className="big-number text-red">0</div>
            <div className="sub-text">Quantum Attacks</div>
          </div>
        </div>

        <div className="table-section">
          <div className="table-header">
            <h2>LIVE NETWORK NODES</h2>
            <span className="last-update">Updated: {lastUpdate.toLocaleTimeString()}</span>
          </div>
          
          <table className="nodes-table">
            <thead>
              <tr>
                <th>NODE ID</th>
                <th>STATUS</th>
                <th>CONNECTIONS</th>
                <th>UPTIME</th>
                <th>LAST SEEN</th>
              </tr>
            </thead>
            <tbody>
              {data?.proxies && data.proxies.length > 0 ? (
                data.proxies.map((proxy) => (
                  <tr key={proxy.id}>
                    <td className="font-mono">{proxy.id}</td>
                    <td><span className="badge success">{proxy.status}</span></td>
                    <td>{proxy.connections} active</td>
                    <td>{proxy.uptime}s</td>
                    <td>{new Date(proxy.last_seen).toLocaleTimeString()}</td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="5" className="text-center">NO NODES DETECTED via SATELLITE LINK</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </main>
    </div>
  )
}

export default App