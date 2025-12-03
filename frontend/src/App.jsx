import { useState, useEffect } from 'react';
import { SignedIn, SignedOut, SignIn, UserButton, useUser } from "@clerk/clerk-react";
import { 
  ShieldCheck, Activity, Server, Globe, 
  CreditCard, Lock, RefreshCw, Zap, LogOut 
} from 'lucide-react';

// URL da API (Render)
const API_URL = "https://pqc-api.onrender.com/api/v1";

// --- COMPONENTES VISUAIS ---

// 1. Tela de Login (Aparece se não estiver logado)
const LoginPage = () => (
  <div className="min-h-screen flex flex-col items-center justify-center bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-slate-900 via-[#0F172A] to-black">
    <div className="text-center mb-8">
      <div className="flex justify-center mb-4">
        <div className="p-4 bg-primary/10 rounded-full border border-primary/20 shadow-[0_0_30px_rgba(16,185,129,0.2)]">
          <ShieldCheck size={48} className="text-primary" />
        </div>
      </div>
      <h1 className="text-4xl font-bold text-white tracking-tight">PQC SHIELD</h1>
      <p className="text-slate-400 mt-2">Enterprise Post-Quantum Cryptography</p>
    </div>
    <div className="bg-card p-1 rounded-xl border border-slate-700 shadow-2xl">
      <SignIn />
    </div>
  </div>
);

// 2. O Dashboard Real (Aparece apenas quando logado)
const Dashboard = () => {
  const { user } = useUser();
  const [data, setData] = useState(null);
  const [activeTab, setActiveTab] = useState('overview'); // overview, nodes, billing
  const [selectedProxy, setSelectedProxy] = useState("");
  const [newTarget, setNewTarget] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  // Busca dados
  const fetchData = async () => {
    try {
      const response = await fetch(`${API_URL}/stats`);
      const json = await response.json();
      setData(json);
    } catch (error) { console.error(error); }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleConfigUpdate = async (e) => {
    e.preventDefault();
    setIsUpdating(true);
    try {
      await fetch(`${API_URL}/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ proxy_id: selectedProxy, new_target: newTarget })
      });
      alert("Rota atualizada com sucesso!");
      setNewTarget("");
    } catch (e) { alert("Erro ao atualizar."); }
    setIsUpdating(false);
  };

  return (
    <div className="min-h-screen bg-dark flex">
      {/* SIDEBAR */}
      <aside className="w-64 bg-card border-r border-slate-800 flex flex-col">
        <div className="p-6 flex items-center gap-3 border-b border-slate-800">
          <ShieldCheck className="text-primary" />
          <span className="font-bold text-white tracking-wide">PQC SHIELD</span>
        </div>
        
        <nav className="flex-1 p-4 space-y-2">
          <SidebarItem icon={<Activity />} label="Overview" active={activeTab === 'overview'} onClick={() => setActiveTab('overview')} />
          <SidebarItem icon={<Server />} label="Nodes & Routing" active={activeTab === 'nodes'} onClick={() => setActiveTab('nodes')} />
          <SidebarItem icon={<CreditCard />} label="Billing & Plan" active={activeTab === 'billing'} onClick={() => setActiveTab('billing')} />
        </nav>

        <div className="p-4 border-t border-slate-800">
          <div className="flex items-center gap-3 px-4 py-3 rounded-lg bg-slate-900/50">
            <UserButton />
            <div className="flex flex-col overflow-hidden">
              <span className="text-sm font-medium text-white truncate">{user.fullName}</span>
              <span className="text-xs text-slate-500 truncate">{user.primaryEmailAddress.toString()}</span>
            </div>
          </div>
        </div>
      </aside>

      {/* MAIN CONTENT */}
      <main className="flex-1 overflow-y-auto">
        <header className="h-16 border-b border-slate-800 flex items-center justify-between px-8 bg-dark/50 backdrop-blur-sm sticky top-0 z-10">
          <h2 className="text-lg font-semibold text-white capitalize">{activeTab}</h2>
          <div className="flex items-center gap-2 text-xs font-mono text-primary bg-primary/10 px-3 py-1 rounded-full border border-primary/20">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
            </span>
            SYSTEM OPERATIONAL
          </div>
        </header>

        <div className="p-8">
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Stats Cards */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <StatCard title="Active Nodes" value={data?.total_proxies || 0} icon={<Server className="text-accent" />} />
                <StatCard title="Security Level" value="NIST L3" sub="Kyber-768 Enforced" icon={<Lock className="text-primary" />} />
                <StatCard title="Target Latency" value="< 50µs" sub="Hardware Accelerated" icon={<Zap className="text-yellow-400" />} />
              </div>

              {/* Tabela Rápida */}
              <div className="bg-card rounded-xl border border-slate-800 overflow-hidden">
                <div className="p-4 border-b border-slate-800 flex justify-between items-center">
                  <h3 className="font-semibold text-slate-200">Live Telemetry</h3>
                  <RefreshCw size={16} className="text-slate-500" />
                </div>
                <Table proxies={data?.proxies} />
              </div>
            </div>
          )}

          {activeTab === 'nodes' && (
            <div className="max-w-2xl mx-auto space-y-8">
              <div className="bg-card p-6 rounded-xl border border-slate-800 shadow-lg">
                <h3 className="text-xl font-bold text-white mb-1">Command Center</h3>
                <p className="text-slate-400 text-sm mb-6">Reconfigure o roteamento dos seus proxies em tempo real.</p>
                
                <form onSubmit={handleConfigUpdate} className="space-y-4">
                  <div>
                    <label className="block text-xs font-bold text-slate-500 uppercase mb-2">Select Node</label>
                    <select 
                      className="w-full bg-dark border border-slate-700 rounded-lg p-3 text-white focus:border-primary outline-none"
                      onChange={(e) => setSelectedProxy(e.target.value)}
                    >
                      <option value="">Choose a proxy...</option>
                      {data?.proxies?.map(p => <option key={p.id} value={p.id}>{p.id}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs font-bold text-slate-500 uppercase mb-2">New Target Destination</label>
                    <input 
                      type="text" 
                      placeholder="Ex: db-prod.internal:5432"
                      className="w-full bg-dark border border-slate-700 rounded-lg p-3 text-white focus:border-primary outline-none font-mono"
                      value={newTarget}
                      onChange={(e) => setNewTarget(e.target.value)}
                    />
                  </div>
                  <button 
                    disabled={isUpdating || !selectedProxy}
                    className="w-full bg-primary hover:bg-emerald-400 text-dark font-bold py-3 rounded-lg transition-all flex items-center justify-center gap-2 disabled:opacity-50"
                  >
                    {isUpdating ? <RefreshCw className="animate-spin" /> : <Globe size={18} />}
                    UPDATE ROUTING
                  </button>
                </form>
              </div>
            </div>
          )}

          {activeTab === 'billing' && (
             <div className="max-w-4xl mx-auto">
               <div className="bg-gradient-to-br from-card to-slate-900 border border-slate-700 rounded-2xl p-8 text-center">
                 <div className="inline-flex p-4 bg-primary/10 rounded-full mb-6">
                   <CreditCard size={32} className="text-primary" />
                 </div>
                 <h2 className="text-3xl font-bold text-white mb-2">Enterprise Plan</h2>
                 <p className="text-slate-400 mb-8">Proteção Pós-Quântica ilimitada para sua infraestrutura crítica.</p>
                 
                 <div className="flex justify-center gap-8 mb-10">
                    <div className="text-left">
                      <div className="text-2xl font-bold text-white">$499<span className="text-sm text-slate-500">/mo</span></div>
                      <div className="text-sm text-slate-400">per node</div>
                    </div>
                 </div>

                 <button className="bg-white text-black px-8 py-3 rounded-lg font-bold hover:bg-slate-200 transition-colors">
                   Manage Subscription (Stripe)
                 </button>
                 
                 <div className="mt-8 pt-8 border-t border-slate-800 text-left">
                   <h4 className="text-sm font-bold text-slate-300 mb-4">YOUR LICENSE KEY</h4>
                   <div className="bg-black p-4 rounded border border-slate-700 font-mono text-primary flex justify-between items-center">
                     <span>SAAS-ENTERPRISE-XYZ</span>
                     <span className="text-xs text-slate-500 uppercase border border-slate-700 px-2 py-1 rounded">Active</span>
                   </div>
                   <p className="text-xs text-slate-500 mt-2">Use esta chave ao iniciar seus containers Docker.</p>
                 </div>
               </div>
             </div>
          )}
        </div>
      </main>
    </div>
  );
};

// --- SUB-COMPONENTES AUXILIARES ---

const SidebarItem = ({ icon, label, active, onClick }) => (
  <button 
    onClick={onClick}
    className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-colors text-sm font-medium ${active ? 'bg-primary/10 text-primary border border-primary/20' : 'text-slate-400 hover:text-white hover:bg-white/5'}`}
  >
    {icon} {label}
  </button>
);

const StatCard = ({ title, value, sub, icon }) => (
  <div className="bg-card p-6 rounded-xl border border-slate-800">
    <div className="flex justify-between items-start mb-4">
      <div>
        <h3 className="text-slate-400 text-xs font-bold uppercase tracking-wider">{title}</h3>
        <div className="text-3xl font-bold text-white mt-1">{value}</div>
      </div>
      <div className="p-2 bg-slate-800 rounded-lg">{icon}</div>
    </div>
    {sub && <div className="text-xs text-slate-500 font-mono bg-slate-900 inline-block px-2 py-1 rounded">{sub}</div>}
  </div>
);

const Table = ({ proxies }) => (
  <table className="w-full text-sm text-left">
    <thead className="text-xs text-slate-500 uppercase bg-slate-900/50">
      <tr>
        <th className="px-6 py-4">Node ID</th>
        <th className="px-6 py-4">Status</th>
        <th className="px-6 py-4">Target</th>
        <th className="px-6 py-4">Uptime</th>
      </tr>
    </thead>
    <tbody className="divide-y divide-slate-800">
      {proxies?.map(p => (
        <tr key={p.id} className="hover:bg-white/5 transition-colors">
          <td className="px-6 py-4 font-mono text-accent">{p.id}</td>
          <td className="px-6 py-4"><span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20"><span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>{p.status}</span></td>
          <td className="px-6 py-4 text-slate-300 font-mono">{p.target_addr}</td>
          <td className="px-6 py-4 text-slate-400">{p.uptime}s</td>
        </tr>
      )) || <tr><td colSpan="4" className="px-6 py-8 text-center text-slate-500">No signals detected via satellite uplink.</td></tr>}
    </tbody>
  </table>
);

// --- APP ROOT (AUTH WRAPPER) ---
import { ClerkProvider } from '@clerk/clerk-react';

const clerkPubKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;

if (!clerkPubKey) {
  throw new Error("Missing Publishable Key. Add VITE_CLERK_PUBLISHABLE_KEY to .env.local");
}

function App() {
  return (
    <ClerkProvider publishableKey={clerkPubKey}>
      <SignedOut>
        <LoginPage />
      </SignedOut>
      <SignedIn>
        <Dashboard />
      </SignedIn>
    </ClerkProvider>
  );
}

export default App;