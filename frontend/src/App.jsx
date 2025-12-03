import { useState, useEffect } from 'react'
import './App.css'

// URL da sua API no Render
const API_URL = "https://pqc-api.onrender.com/api/v1";

function App() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(new Date());
  
  // Estado para o formulário de Configuração
  const [selectedProxy, setSelectedProxy] = useState(null);
  const [newTarget, setNewTarget] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  const fetchData = async () => {
    try {
      const response = await fetch(`${API_URL}/stats`);
      const json = await response.json();
      setData(json);
      setLastUpdate(new Date());
    } catch (error) {
      console.error("Erro ao buscar dados:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleConfigUpdate = async (e) => {
    e.preventDefault();
    if (!selectedProxy || !newTarget) return;

    setIsUpdating(true);
    try {
      await fetch(`${API_URL}/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          proxy_id: selectedProxy,
          new_target: newTarget
        })
      });
      alert(`Comando enviado! O Proxy ${selectedProxy} atualizará em breve.`);
      setNewTarget("");
    } catch (error) {
      alert("Erro ao enviar comando.");
    } finally {
      setIsUpdating(false);
    }
  };

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
        {/* STATS CARDS */}
        <div className="stats-grid">
          <div className="card">
            <h3>ACTIVE NODES</h3>
            <div className="big-number">{data?.total_proxies || 0}</div>
          </div>
          <div className="card">
            <h3>GLOBAL TARGET</h3>
            <div className="sub-text highlight">
               {data?.proxies?.[0]?.target_addr || "Loading..."}
            </div>
          </div>
          <div className="card">
            <h3>SECURITY LEVEL</h3>
            <div className="big-number text-green">MAX</div>
            <div className="sub-text">Kyber-768 Enabled</div>
          </div>
        </div>

        {/* CONTROL PANEL (NOVO) */}
        <div className="control-panel">
          <h2>🎮 COMMAND CENTER</h2>
          <form className="config-form" onSubmit={handleConfigUpdate}>
            <div className="form-group">
              <label>SELECT NODE</label>
              <select 
                onChange={(e) => setSelectedProxy(e.target.value)}
                defaultValue=""
              >
                <option value="" disabled>Select a Proxy...</option>
                {data?.proxies?.map(p => (
                  <option key={p.id} value={p.id}>{p.id} ({p.status})</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>NEW TARGET DESTINATION (e.g. google.com:80)</label>
              <input 
                type="text" 
                placeholder="domain.com:port" 
                value={newTarget}
                onChange={(e) => setNewTarget(e.target.value)}
              />
            </div>
            <button type="submit" disabled={isUpdating || !selectedProxy}>
              {isUpdating ? "SENDING..." : "UPDATE ROUTING"}
            </button>
          </form>
        </div>

        {/* TABLE SECTION */}
        <div className="table-section">
          <div className="table-header">
            <h2>LIVE TELEMETRY</h2>
            <span className="last-update">Updated: {lastUpdate.toLocaleTimeString()}</span>
          </div>
          
          <table className="nodes-table">
            <thead>
              <tr>
                <th>NODE ID</th>
                <th>TARGET</th>
                <th>STATUS</th>
                <th>CONNECTIONS</th>
                <th>LAST SEEN</th>
              </tr>
            </thead>
            <tbody>
              {data?.proxies && data.proxies.length > 0 ? (
                data.proxies.map((proxy) => (
                  <tr key={proxy.id}>
                    <td className="font-mono">{proxy.id}</td>
                    <td className="text-blue">{proxy.target_addr}</td>
                    <td><span className="badge success">{proxy.status}</span></td>
                    <td>{proxy.connections}</td>
                    <td>{new Date(proxy.last_seen).toLocaleTimeString()}</td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="5" className="text-center">NO DATA</td>
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