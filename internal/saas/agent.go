package saas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/4kciclone/pqc-proxy/internal/config" // Importe o novo pacote config
)

type AgentConfig struct {
	ProxyID    string
	LicenseKey string
	CloudURL   string
}

func StartAgent(cfg AgentConfig) {
	fmt.Println("📡 Iniciando Agente SaaS (Telemetry & Dynamic Config)...")
	
	startTime := time.Now()
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for range ticker.C {
			sendHeartbeat(cfg, startTime)
		}
	}()
}

func sendHeartbeat(cfg AgentConfig, startTime time.Time) {
	payload := map[string]interface{}{
		"proxy_id":           cfg.ProxyID,
		"license_key":        cfg.LicenseKey,
		"status":             "healthy",
		"active_connections": 1, // Mock por enquanto
		"uptime_seconds":     int64(time.Since(startTime).Seconds()),
	}

	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(cfg.CloudURL+"/api/v1/heartbeat", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("⚠️  [Agente] Falha ao contatar SaaS: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// Decodifica resposta para pegar config remota
		var cloudResp struct {
			Status  string `json:"status"`
			Command string `json:"command"`
			Target  string `json:"target"` // Campo que a nuvem manda
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err == nil {
			// Lógica de Atualização Dinâmica
			current := config.GetTarget()
			
			// Se a nuvem mandou um alvo válido e é diferente do atual
			if cloudResp.Target != "" && cloudResp.Target != current {
				log.Printf("🔄 [Agente] Reconfiguração Remota Recebida!")
				log.Printf("   ANTIGO: %s  --->  NOVO: %s", current, cloudResp.Target)
				
				// Atualiza a variável global (Thread-Safe)
				config.SetTarget(cloudResp.Target)
			}
		}
	} else if resp.StatusCode == 403 {
		log.Fatalf("🚨 [Agente] LICENÇA INVÁLIDA ou SUSPENSA! Desligando Proxy...")
	}
}