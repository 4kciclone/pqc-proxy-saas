package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" // Driver PostgreSQL
)

var db *sql.DB

// Modelo de Dados (Espelho da Tabela SQL)
type ProxyStats struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	LicenseKey  string    `json:"-"` // Não enviar pro frontend
	Connections int       `json:"connections"`
	Uptime      int64     `json:"uptime"`
	LastSeen    time.Time `json:"last_seen"`
}

func initDB() {
	// Pega a URL do Banco das variáveis de ambiente do Render
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Fallback para teste local (se você tiver postgres local rodando)
		// Se não tiver, o código vai falhar ao tentar conectar, o que é esperado em prod sem config.
		log.Println("⚠️ DATABASE_URL não definida. Tentando conectar em localhost...")
		connStr = "postgres://user:password@localhost/pqc_db?sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Tenta conectar
	if err = db.Ping(); err != nil {
		log.Printf("Erro ao conectar no DB: %v", err)
	} else {
		log.Println("✅ Conectado ao PostgreSQL com sucesso!")
	}

	// Criar Tabela se não existir (Migração Automática)
	query := `
	CREATE TABLE IF NOT EXISTS proxies (
		id TEXT PRIMARY KEY,
		license_key TEXT,
		status TEXT,
		connections INT,
		uptime BIGINT,
		last_seen TIMESTAMP
	);`
	
	if _, err := db.Exec(query); err != nil {
		log.Fatal("Erro ao criar tabelas:", err)
	}
}

// Middleware CORS (Para o Vercel acessar)
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

func main() {
	// Inicializa conexão com o Banco
	initDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/v1/heartbeat", enableCORS(handleHeartbeat))
	http.HandleFunc("/api/v1/stats", enableCORS(handleStats))

	fmt.Printf("🚀 Cloud API (PostgreSQL) rodando na porta %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

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

	// Validação de Licença Real
	if req.LicenseKey != "SAAS-ENTERPRISE-XYZ" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// SQL: Upsert (Inserir ou Atualizar se já existe)
	// Sintaxe compatível com PostgreSQL
	query := `
		INSERT INTO proxies (id, license_key, status, connections, uptime, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) 
		DO UPDATE SET 
			status = EXCLUDED.status,
			connections = EXCLUDED.connections,
			uptime = EXCLUDED.uptime,
			last_seen = EXCLUDED.last_seen;
	`
	
	_, err := db.Exec(query, req.ProxyID, req.LicenseKey, req.Status, req.Connections, req.Uptime, time.Now())
	if err != nil {
		log.Printf("Erro no Banco de Dados: %v", err)
		http.Error(w, "Database Error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "authorized", "command": "continue"}`))
	fmt.Printf("💾 Persistido no DB: %s\n", req.ProxyID)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	// Busca dados reais do banco
	rows, err := db.Query("SELECT id, status, connections, uptime, last_seen FROM proxies")
	if err != nil {
		http.Error(w, "Database Error", 500)
		return
	}
	defer rows.Close()

	var proxies []ProxyStats
	for rows.Next() {
		var p ProxyStats
		if err := rows.Scan(&p.ID, &p.Status, &p.Connections, &p.Uptime, &p.LastSeen); err != nil {
			continue
		}
		proxies = append(proxies, p)
	}

	// Formata JSON de resposta
	response := map[string]interface{}{
		"proxies":       proxies,
		"total_proxies": len(proxies),
		"db_status":     "connected",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}