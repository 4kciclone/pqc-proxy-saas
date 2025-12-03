package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

var db *sql.DB

// --- CONFIGURAÇÃO STRIPE (SUBSTITUA PELOS SEUS IDs REAIS) ---
const (
	PRICE_PRO        = "prod_TXMdxzOEpiKewT" // Cole o ID do plano Pro aqui
	PRICE_ENTERPRISE = "prod_TXMehthkHNkiw0" // Cole o ID do plano Enterprise aqui
	FRONTEND_URL     = "https://pqc-proxy-saas.vercel.app" // Sua URL Vercel
)

// --- MODELOS DE DADOS ---

type ProxyStats struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	LicenseKey  string    `json:"-"`
	Connections int       `json:"connections"`
	Uptime      int64     `json:"uptime"`
	LastSeen    time.Time `json:"last_seen"`
	TargetAddr  string    `json:"target_addr"`
}

type Tenant struct {
	UserID     string `json:"user_id"`
	Plan       string `json:"plan"`
	LicenseKey string `json:"license_key"`
}

// --- INICIALIZAÇÃO DO BANCO ---

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Println("⚠️ DATABASE_URL não definida. Tentando localhost...")
		connStr = "postgres://user:password@localhost/pqc_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Printf("Erro ao conectar no DB: %v", err)
	} else {
		log.Println("✅ Conectado ao PostgreSQL!")
	}

	// Criação das Tabelas
	queries := []string{
		`CREATE TABLE IF NOT EXISTS proxies (
			id TEXT PRIMARY KEY,
			license_key TEXT,
			status TEXT,
			connections INT,
			uptime BIGINT,
			last_seen TIMESTAMP,
			target_addr TEXT DEFAULT 'google.com:80'
		);`,
		`CREATE TABLE IF NOT EXISTS tenants (
            user_id TEXT PRIMARY KEY,
            email TEXT,
            stripe_customer_id TEXT,
            plan_tier TEXT DEFAULT 'free',
            license_key TEXT UNIQUE,
            status TEXT DEFAULT 'active'
        );`,
	}
	
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Migração DB: %v", err)
		}
	}
}

// --- MIDDLEWARES ---

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// --- FUNÇÕES AUXILIARES ---

func generateLicenseKey() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("PQC-%X", b)
}

// --- MAIN ---

func main() {
	initDB()
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Rotas Públicas/API
	http.HandleFunc("/api/v1/heartbeat", enableCORS(handleHeartbeat))
	http.HandleFunc("/api/v1/stats", enableCORS(handleStats))
	http.HandleFunc("/api/v1/config", enableCORS(handleUpdateConfig))
	
	// Rotas de Pagamento & Usuário
	http.HandleFunc("/api/v1/billing/checkout", enableCORS(handleCreateCheckoutSession))
	http.HandleFunc("/api/v1/webhook", handleStripeWebhook)
	http.HandleFunc("/api/v1/me", enableCORS(handleGetMyPlan))

	fmt.Printf("🚀 Cloud API rodando na porta %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// --- HANDLERS (TELEMETRIA & CONFIG) ---

func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		ProxyID     string `json:"proxy_id"`
		LicenseKey  string `json:"license_key"`
		Status      string `json:"status"`
		Connections int    `json:"active_connections"`
		Uptime      int64  `json:"uptime_seconds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	// Verifica licença (Hardcoded para testes + DB Check futuro)
	// Em prod real, faríamos um SELECT na tabela tenants para validar a LicenseKey
	if req.LicenseKey != "SAAS-ENTERPRISE-XYZ" && req.LicenseKey != "DOCKER-TEST-KEY" {
		// Verifica se a licença existe na tabela de tenants pagos
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE license_key=$1)", req.LicenseKey).Scan(&exists)
		if err != nil || !exists {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// Atualiza Proxy
	query := `
		INSERT INTO proxies (id, license_key, status, connections, uptime, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET 
			status = EXCLUDED.status, connections = EXCLUDED.connections, 
			uptime = EXCLUDED.uptime, last_seen = EXCLUDED.last_seen;
	`
	db.Exec(query, req.ProxyID, req.LicenseKey, req.Status, req.Connections, req.Uptime, time.Now())

	// Pega Config
	var targetAddr string
	db.QueryRow("SELECT target_addr FROM proxies WHERE id=$1", req.ProxyID).Scan(&targetAddr)
	if targetAddr == "" { targetAddr = "google.com:80" }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "authorized", "command": "continue", "target": targetAddr,
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, status, connections, uptime, last_seen, target_addr FROM proxies")
	if err != nil {
		http.Error(w, "DB Error", 500)
		return
	}
	defer rows.Close()

	var proxies []ProxyStats
	for rows.Next() {
		var p ProxyStats
		var target sql.NullString
		if err := rows.Scan(&p.ID, &p.Status, &p.Connections, &p.Uptime, &p.LastSeen, &target); err == nil {
			if target.Valid { p.TargetAddr = target.String } else { p.TargetAddr = "google.com:80" }
			proxies = append(proxies, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"proxies": proxies, "total_proxies": len(proxies),
	})
}

func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var req struct { ProxyID, NewTarget string }
	json.NewDecoder(r.Body).Decode(&req)

	_, err := db.Exec("UPDATE proxies SET target_addr=$1 WHERE id=$2", req.NewTarget, req.ProxyID)
	if err != nil {
		http.Error(w, "DB Error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"updated"}`))
}

// --- HANDLERS (PAGAMENTO) ---

func handleCreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w, "405", 405); return }

	var req struct { UserID, Email, Plan string }
	json.NewDecoder(r.Body).Decode(&req)

	priceId := ""
	if req.Plan == "pro" { priceId = PRICE_PRO } else { priceId = PRICE_ENTERPRISE }

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(priceId), Quantity: stripe.Int64(1)},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(FRONTEND_URL + "/dashboard?success=true"),
		CancelURL:  stripe.String(FRONTEND_URL + "/dashboard?canceled=true"),
		ClientReferenceID: stripe.String(req.UserID),
		CustomerEmail: stripe.String(req.Email),
	}

	s, err := session.New(params)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"url": s.URL})
}

func handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil { w.WriteHeader(http.StatusBadRequest); return }

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil { w.WriteHeader(http.StatusBadRequest); return }

	if event.Type == "checkout.session.completed" {
		var s stripe.CheckoutSession
		json.Unmarshal(event.Data.Raw, &s)

		plan := "pro"
		if s.AmountTotal > 10000 { plan = "enterprise" } // Lógica simples baseada no valor
		
		newKey := generateLicenseKey()
		
		query := `INSERT INTO tenants (user_id, email, stripe_customer_id, plan_tier, license_key) 
				  VALUES ($1, $2, $3, $4, $5) 
				  ON CONFLICT (user_id) DO UPDATE SET plan_tier=$4, license_key=$5`
		db.Exec(query, s.ClientReferenceID, s.CustomerEmail, s.Customer.ID, plan, newKey)
		fmt.Printf("💰 Venda: %s - %s\n", s.CustomerEmail, plan)
	}
	w.WriteHeader(200)
}

func handleGetMyPlan(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	var t Tenant
	err := db.QueryRow("SELECT plan_tier, license_key FROM tenants WHERE user_id=$1", userID).Scan(&t.Plan, &t.LicenseKey)
	if err != nil {
		t.Plan = "free"; t.LicenseKey = "FREE-TRIAL"
	}
	json.NewEncoder(w).Encode(t)
}