package saas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type AgentConfig struct {
	ProxyID    string
	LicenseKey string
	CloudURL   string
}

// StartAgent roda em background enviando métricas
func StartAgent(cfg AgentConfig) {
	fmt.Println("📡 Iniciando Agente SaaS (Telemetry & Licensing)...")
	
	startTime := time.Now()
	ticker := time.NewTicker(5 * time.Second) // Reporta a cada 5s

	// Loop infinito em background
	go func() {
		for range ticker.C {
			sendHeartbeat(cfg, startTime)
		}
	}()
}

func sendHeartbeat(cfg AgentConfig, startTime time.Time) {
	// Dados que enviamos para a nuvem
	payload := map[string]interface{}{
		"proxy_id":           cfg.ProxyID,
		"license_key":        cfg.LicenseKey,
		"status":             "healthy",
		"active_connections": 1, // Mock: No futuro pegaremos do contador real
		"uptime_seconds":     int64(time.Since(startTime).Seconds()),
	}

	jsonPayload, _ := json.Marshal(payload)

	// Envia POST para a nuvem
	resp, err := http.Post(cfg.CloudURL+"/api/v1/heartbeat", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("⚠️  [Agente] Falha ao contatar SaaS: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// Silencioso se der certo, para não poluir o log
	} else if resp.StatusCode == 403 {
		log.Fatalf("🚨 [Agente] LICENÇA INVÁLIDA! O SaaS bloqueou este proxy. Desligando...")
	} else {
		log.Printf("⚠️  [Agente] Nuvem respondeu com erro: %d", resp.StatusCode)
	}
}