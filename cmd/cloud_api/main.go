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

// Modelo de Dados
type ProxyStats struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	LicenseKey  string    `json:"-"`
	Connections int       `json:"connections"`
	Uptime      int64     `json:"uptime"`
	LastSeen    time.Time `json:"last_seen"`
	TargetAddr  string    `json:"target_addr"` // Campo Novo
}

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

	// Migração: Adiciona coluna target_addr se não existir
	// Nota: Em produção real usariamos ferramentas de migração, aqui fazemos manual
	queries := []string{
		`CREATE TABLE IF NOT EXISTS proxies (
			id TEXT PRIMARY KEY,
			license_key TEXT,
			status TEXT,
			connections INT,
			uptime BIGINT,
			last_seen TIMESTAMP
		);`,
		// Adiciona coluna se não existir (Postgres 9.6+)
		`ALTER TABLE proxies ADD COLUMN IF NOT EXISTS target_addr TEXT DEFAULT 'google.com:80';`,
	}
	
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Printf("Aviso na migração: %v", err)
		}
	}
}

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
	initDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/v1/heartbeat", enableCORS(handleHeartbeat))
	http.HandleFunc("/api/v1/stats", enableCORS(handleStats))
	http.HandleFunc("/api/v1/config", enableCORS(handleUpdateConfig)) // Novo Endpoint

	fmt.Printf("🚀 Cloud API rodando na porta %s...\n", port)
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

	if req.LicenseKey != "SAAS-ENTERPRISE-XYZ" && req.LicenseKey != "DOCKER-TEST-KEY" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 1. Atualiza Status
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
		log.Printf("DB Error: %v", err)
		http.Error(w, "Database Error", 500)
		return
	}

	// 2. Busca Configuração Atual para devolver ao Agente
	var targetAddr string
	err = db.QueryRow("SELECT target_addr FROM proxies WHERE id=$1", req.ProxyID).Scan(&targetAddr)
	if err != nil || targetAddr == "" {
		targetAddr = "google.com:80"
	}

	// 3. Responde com Comandos e Configs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "authorized",
		"command": "continue",
		"target":  targetAddr, // O Agente vai ler isso
	})
	
	fmt.Printf("💓 Heartbeat: %s -> Alvo: %s\n", req.ProxyID, targetAddr)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, status, connections, uptime, last_seen, target_addr FROM proxies")
	if err != nil {
		http.Error(w, "Database Error", 500)
		return
	}
	defer rows.Close()

	var proxies []ProxyStats
	for rows.Next() {
		var p ProxyStats
		// Scan pode falhar se target_addr for NULL no banco antigo, tratando com Default
		var target sql.NullString
		if err := rows.Scan(&p.ID, &p.Status, &p.Connections, &p.Uptime, &p.LastSeen, &target); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		if target.Valid {
			p.TargetAddr = target.String
		} else {
			p.TargetAddr = "google.com:80"
		}
		proxies = append(proxies, p)
	}

	response := map[string]interface{}{
		"proxies":       proxies,
		"total_proxies": len(proxies),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Novo Handler para o Dashboard mudar o alvo
func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		ProxyID   string `json:"proxy_id"`
		NewTarget string `json:"new_target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	_, err := db.Exec("UPDATE proxies SET target_addr=$1 WHERE id=$2", req.NewTarget, req.ProxyID)
	if err != nil {
		http.Error(w, "Database Error", 500)
		return
	}

	fmt.Printf("⚙️ CONFIG ALTERADA: %s agora aponta para %s\n", req.ProxyID, req.NewTarget)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"updated"}`))
}