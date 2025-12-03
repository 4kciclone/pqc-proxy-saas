import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Link, useNavigate, Navigate } from 'react-router-dom';
import { SignedIn, SignedOut, SignIn, UserButton, useUser, ClerkProvider } from "@clerk/clerk-react";
import { 
  ShieldCheck, Activity, Server, Globe, 
  CreditCard, Lock, RefreshCw, Zap, ArrowRight, CheckCircle, Database 
} from 'lucide-react';

// URL da API
const API_URL = "https://pqc-api.onrender.com/api/v1";
const clerkPubKey = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;

if (!clerkPubKey) throw new Error("Missing Clerk Key");

// ==========================================
// 1. LANDING PAGE
// ==========================================
const LandingPage = () => {
  const navigate = useNavigate();
  const { isSignedIn } = useUser();

  return (
    <div className="min-h-screen bg-dark text-slate-200 font-sans selection:bg-primary selection:text-black">
      <nav className="border-b border-slate-800 bg-dark/80 backdrop-blur-md fixed w-full z-50">
        <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <ShieldCheck className="text-primary h-8 w-8" />
            <span className="font-bold text-xl tracking-tight text-white">PQC SHIELD</span>
          </div>
          <div className="flex items-center gap-6">
            <button 
              onClick={() => navigate(isSignedIn ? '/dashboard' : '/sign-in')}
              className="bg-white text-black px-5 py-2.5 rounded-lg font-bold text-sm hover:bg-slate-200 transition-all flex items-center gap-2"
            >
              {isSignedIn ? 'Go to Dashboard' : 'Client Login'} <ArrowRight size={16} />
            </button>
          </div>
        </div>
      </nav>

      <section className="pt-32 pb-20 px-6 text-center">
        <div className="max-w-5xl mx-auto">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-bold uppercase tracking-wider mb-6">
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
            </span>
            NIST Kyber-768 Standard Ready
          </div>
          <h1 className="text-5xl md:text-7xl font-extrabold text-white mb-8 leading-tight">
            Secure Your Legacy Infrastructure Against <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary to-accent">Quantum Threats</span>.
          </h1>
          <p className="text-xl text-slate-400 mb-10 max-w-2xl mx-auto leading-relaxed">
            The "Harvest Now, Decrypt Later" threat is real. PQC Shield provides a drop-in sidecar proxy that upgrades your existing systems to Post-Quantum Cryptography in minutes.
          </p>
          <div className="flex justify-center gap-4">
            <button onClick={() => navigate('/sign-in')} className="bg-primary text-black px-8 py-4 rounded-xl font-bold text-lg hover:bg-emerald-400 hover:scale-105 transition-all shadow-[0_0_30px_rgba(16,185,129,0.3)]">
              Start Free Trial
            </button>
          </div>
        </div>
      </section>

      <section className="py-20 bg-card/50 border-y border-slate-800">
        <div className="max-w-7xl mx-auto px-6 grid md:grid-cols-3 gap-12">
            <FeatureCard icon={<Lock className="text-primary h-8 w-8" />} title="Quantum-Resistant" desc="Uses ML-KEM (Kyber-768) and ML-DSA (Dilithium) to secure handshakes." />
            <FeatureCard icon={<Zap className="text-yellow-400 h-8 w-8" />} title="Wire-Speed Latency" desc="Hybrid architecture (Go + Rust) ensures <50µs overhead." />
            <FeatureCard icon={<Database className="text-accent h-8 w-8" />} title="Drop-in Solution" desc="No code refactoring required. Deploy Docker Sidecar instantly." />
        </div>
      </section>
      
      <footer className="py-12 border-t border-slate-800 text-center text-slate-500 text-sm">
        <p>© 2025 PQC Shield Inc. All rights reserved. NIST & FIPS Compliant.</p>
      </footer>
    </div>
  );
};

const FeatureCard = ({ icon, title, desc }) => (
  <div className="p-6 rounded-2xl bg-card border border-slate-800 hover:border-primary/50 transition-colors group">
    <div className="mb-4 p-3 bg-slate-900 rounded-lg w-fit group-hover:scale-110 transition-transform">{icon}</div>
    <h3 className="text-xl font-bold text-white mb-3">{title}</h3>
    <p className="text-slate-400 leading-relaxed">{desc}</p>
  </div>
);

// ==========================================
// 2. DASHBOARD
// ==========================================
const Dashboard = () => {
  const { user } = useUser();
  const [data, setData] = useState(null);
  const [userPlan, setUserPlan] = useState({ plan: 'free', license_key: 'Loading...' });
  const [activeTab, setActiveTab] = useState('overview'); 
  const [selectedProxy, setSelectedProxy] = useState("");
  const [newTarget, setNewTarget] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  // Fetch Stats & User Plan
  useEffect(() => {
    const fetchData = async () => {
        try {
            const [statsRes, planRes] = await Promise.all([
                fetch(`${API_URL}/stats`),
                fetch(`${API_URL}/me?user_id=${user.id}`)
            ]);
            setData(await statsRes.json());
            setUserPlan(await planRes.json());
        } catch (e) { console.error(e); }
    };
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [user]);

  const handleConfigUpdate = async (e) => {
    e.preventDefault();
    setIsUpdating(true);
    try {
      await fetch(`${API_URL}/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ proxy_id: selectedProxy, new_target: newTarget })
      });
      alert("Comando enviado!");
      setNewTarget("");
    } catch (e) { alert("Erro."); }
    setIsUpdating(false);
  };

  const handleCheckout = async (planType) => {
    try {
        const res = await fetch(`${API_URL}/billing/checkout`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user_id: user.id, email: user.primaryEmailAddress.toString(), plan: planType })
        });
        const { url } = await res.json();
        window.location.href = url;
    } catch (e) { alert("Erro no checkout"); }
  };

  return (
    <div className="min-h-screen bg-dark flex font-sans text-slate-200">
      <aside className="w-64 bg-card border-r border-slate-800 flex flex-col fixed h-full z-20">
        <Link to="/" className="p-6 flex items-center gap-3 border-b border-slate-800 hover:bg-slate-800/50">
          <ShieldCheck className="text-primary" />
          <span className="font-bold text-white tracking-wide">PQC SHIELD</span>
        </Link>
        <nav className="flex-1 p-4 space-y-2">
          <SidebarItem icon={<Activity />} label="Overview" active={activeTab === 'overview'} onClick={() => setActiveTab('overview')} />
          <SidebarItem icon={<Server />} label="Nodes & Routing" active={activeTab === 'nodes'} onClick={() => setActiveTab('nodes')} />
          <SidebarItem icon={<CreditCard />} label="Billing" active={activeTab === 'billing'} onClick={() => setActiveTab('billing')} />
        </nav>
        <div className="p-4 border-t border-slate-800">
            <div className="flex items-center gap-3 px-4 py-3 rounded-lg bg-slate-900/50">
                <UserButton afterSignOutUrl="/" />
                <div className="flex flex-col overflow-hidden">
                    <span className="text-sm font-medium text-white truncate">{user.fullName}</span>
                    <span className="text-xs text-slate-500 uppercase">{userPlan.plan} Plan</span>
                </div>
            </div>
        </div>
      </aside>

      <main className="flex-1 ml-64 min-h-screen">
        <header className="h-16 border-b border-slate-800 flex items-center justify-between px-8 bg-dark/50 backdrop-blur-sm sticky top-0 z-10">
          <h2 className="text-lg font-semibold text-white capitalize">{activeTab}</h2>
          <div className="flex items-center gap-2 text-xs font-mono text-primary bg-primary/10 px-3 py-1 rounded-full border border-primary/20">
            <div className="h-2 w-2 rounded-full bg-primary animate-pulse"></div> SYSTEM OPERATIONAL
          </div>
        </header>

        <div className="p-8">
          {activeTab === 'overview' && (
            <div className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <StatCard title="Active Nodes" value={data?.total_proxies || 0} icon={<Server className="text-accent" />} />
                <StatCard title="Protection" value="Kyber-768" sub="Post-Quantum L3" icon={<Lock className="text-primary" />} />
                <StatCard title="Latency" value="< 50µs" sub="Wire Speed" icon={<Zap className="text-yellow-400" />} />
              </div>
              <div className="bg-card rounded-xl border border-slate-800 overflow-hidden">
                <div className="p-4 border-b border-slate-800 flex justify-between items-center"><h3 className="font-semibold text-slate-200">Live Telemetry</h3><RefreshCw size={16} className="text-slate-500" /></div>
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
                    <label className="block text-xs font-bold text-slate-500 uppercase mb-2">Select Proxy Node</label>
                    <select className="w-full bg-dark border border-slate-700 rounded-lg p-3 text-white focus:border-primary outline-none" onChange={(e) => setSelectedProxy(e.target.value)}>
                      <option value="">Choose a proxy...</option>
                      {data?.proxies?.map(p => <option key={p.id} value={p.id}>{p.id}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-xs font-bold text-slate-500 uppercase mb-2">New Target (e.g., ip:port)</label>
                    <input type="text" placeholder="domain.internal:port" className="w-full bg-dark border border-slate-700 rounded-lg p-3 text-white focus:border-primary outline-none font-mono" value={newTarget} onChange={(e) => setNewTarget(e.target.value)} />
                  </div>
                  <button disabled={isUpdating || !selectedProxy} className="w-full bg-primary hover:bg-emerald-400 text-dark font-bold py-3 rounded-lg flex justify-center gap-2 disabled:opacity-50">
                    {isUpdating ? <RefreshCw className="animate-spin" /> : <Globe size={18} />} UPDATE ROUTING
                  </button>
                </form>
              </div>
            </div>
          )}

          {activeTab === 'billing' && (
             <div className="max-w-6xl mx-auto space-y-8">
               <div className="text-center mb-8"><h2 className="text-3xl font-bold text-white">Choose Your Shield Level</h2></div>
               <div className="grid md:grid-cols-3 gap-8">
                 <PricingCard title="Developer" price="$0" features={['1 Node', 'Community Support']} current={userPlan.plan === 'free'} onClick={() => {}} />
                 <PricingCard title="Pro" price="$49" period="/mo" highlight features={['10 Nodes', 'Kyber-768']} current={userPlan.plan === 'pro'} onClick={() => handleCheckout('pro')} />
                 <PricingCard title="Enterprise" price="$499" period="/mo" features={['Unlimited', 'SLA Support']} current={userPlan.plan === 'enterprise'} onClick={() => handleCheckout('enterprise')} />
               </div>
               <div className="bg-card border border-slate-700 rounded-2xl p-8 flex flex-col md:flex-row justify-between items-center gap-6">
                 <div><h3 className="text-lg font-bold text-white">YOUR LICENSE KEY</h3><p className="text-sm text-slate-400">Use this key in Docker.</p></div>
                 <div className="bg-black p-4 rounded border border-slate-700 font-mono text-primary flex items-center gap-4">
                   <span>{userPlan.license_key}</span>
                   <button onClick={() => navigator.clipboard.writeText(userPlan.license_key)} className="text-xs bg-slate-800 px-3 py-1 rounded">COPY</button>
                 </div>
               </div>
             </div>
          )}
        </div>
      </main>
    </div>
  );
};

// Componentes UI Auxiliares
const SidebarItem = ({ icon, label, active, onClick }) => (
  <button onClick={onClick} className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-colors ${active ? 'bg-primary/10 text-primary border border-primary/20' : 'text-slate-400 hover:text-white hover:bg-white/5'}`}>{icon} {label}</button>
);
const StatCard = ({ title, value, sub, icon }) => (
  <div className="bg-card p-6 rounded-xl border border-slate-800 flex justify-between items-start">
    <div><h3 className="text-slate-400 text-xs font-bold uppercase">{title}</h3><div className="text-3xl font-bold text-white mt-1">{value}</div>{sub && <div className="text-xs text-slate-500 bg-slate-900 inline-block px-2 py-1 rounded mt-2">{sub}</div>}</div>
    <div className="p-2 bg-slate-800 rounded-lg">{icon}</div>
  </div>
);
const PricingCard = ({ title, price, period, features, highlight, current, onClick }) => (
  <div className={`p-8 rounded-2xl border flex flex-col ${highlight ? 'bg-slate-900 border-primary shadow-[0_0_30px_rgba(16,185,129,0.1)]' : 'bg-card border-slate-800'}`}>
    <h3 className="text-xl font-bold text-white mb-2">{title}</h3>
    <div className="text-4xl font-bold text-white mb-6">{price}<span className="text-sm text-slate-500 font-normal">{period}</span></div>
    <ul className="space-y-4 mb-8 flex-1">{features.map((f, i) => <li key={i} className="flex items-center gap-3 text-sm text-slate-300"><CheckCircle size={16} className={highlight ? "text-primary" : "text-slate-500"} /> {f}</li>)}</ul>
    <button onClick={onClick} disabled={current} className={`w-full py-3 rounded-lg font-bold transition-all ${current ? 'bg-slate-700 text-slate-400 cursor-default' : highlight ? 'bg-primary hover:bg-emerald-400 text-black' : 'bg-white hover:bg-slate-200 text-black'}`}>{current ? 'CURRENT PLAN' : 'SUBSCRIBE'}</button>
  </div>
);
const Table = ({ proxies }) => (
  <table className="w-full text-sm text-left"><thead className="text-xs text-slate-500 uppercase bg-slate-900/50"><tr><th className="px-6 py-4">Node ID</th><th className="px-6 py-4">Status</th><th className="px-6 py-4">Target</th><th className="px-6 py-4">Uptime</th></tr></thead><tbody className="divide-y divide-slate-800">{proxies?.map(p => (<tr key={p.id} className="hover:bg-white/5"><td className="px-6 py-4 font-mono text-accent">{p.id}</td><td className="px-6 py-4"><span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-500 border border-emerald-500/20"><span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>{p.status}</span></td><td className="px-6 py-4 text-slate-300 font-mono">{p.target_addr}</td><td className="px-6 py-4 text-slate-400">{p.uptime}s</td></tr>)) || <tr><td colSpan="4" className="px-6 py-8 text-center text-slate-500">No nodes detected.</td></tr>}</tbody></table>
);

// TELA DE LOGIN
const LoginPage = () => (
    <div className="min-h-screen flex flex-col items-center justify-center bg-dark">
      <Link to="/" className="mb-8 hover:opacity-80"><div className="p-4 bg-primary/10 rounded-full border border-primary/20"><ShieldCheck size={48} className="text-primary" /></div></Link>
      <h1 className="text-3xl font-bold text-white mb-8">Access Control</h1>
      <SignIn routing="path" path="/sign-in" signUpUrl="/sign-in" forceRedirectUrl="/dashboard" />
    </div>
);

// APP ROUTER
function App() {
  return (
    <ClerkProvider publishableKey={clerkPubKey}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/sign-in/*" element={<SignedOut><LoginPage /></SignedOut>} />
          <Route path="/dashboard" element={<><SignedIn><Dashboard /></SignedIn><SignedOut><Navigate to="/sign-in" replace /></SignedOut></>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </ClerkProvider>
  );
}

export default App;