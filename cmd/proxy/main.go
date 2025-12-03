package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/4kciclone/pqc-proxy/internal/network"
	"github.com/4kciclone/pqc-proxy/internal/saas" // <--- NOVO IMPORT
)

func main() {
	mode := flag.String("mode", "server", "Modo: 'server' ou 'client'")
	port := flag.String("port", "4433", "Porta PQC")
	addr := flag.String("addr", "localhost", "Endereço do servidor")
	
	// Flags de SaaS
	license := flag.String("license", "SAAS-ENTERPRISE-XYZ", "Chave de Licença")
	cloud := flag.String("cloud", "http://localhost:8080", "URL da API SaaS")
	
	flag.Parse()

	if *mode == "server" {
		// 1. Iniciar o Agente SaaS em background
		agentCfg := saas.AgentConfig{
			ProxyID:    "proxy-node-01",
			LicenseKey: *license,
			CloudURL:   *cloud,
		}
		saas.StartAgent(agentCfg)

		// 2. Iniciar o Servidor TCP (Bloqueia a thread principal)
		network.StartServer(*port)

	} else if *mode == "client" {
		fullAddr := fmt.Sprintf("%s:%s", *addr, *port)
		network.StartClient(fullAddr)
	} else {
		fmt.Println("Modo inválido.")
		os.Exit(1)
	}
}