package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/4kciclone/pqc-proxy/internal/config" // Importe o config
	"github.com/4kciclone/pqc-proxy/pkg/crypto_core"
)

// StartServer inicia o listener TCP PQC
func StartServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
	defer listener.Close()

	fmt.Printf("🛡️  PQC Proxy (Server) ouvindo em 0.0.0.0:%s\n", port)
	// Mostra o target inicial
	fmt.Printf("🎯 Target Inicial: %s (Gerenciado via SaaS)\n", config.GetTarget())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Erro no accept: %v", err)
			continue
		}
		go handlePQCConnection(conn)
	}
}

func handlePQCConnection(clientConn net.Conn) {
	defer clientConn.Close()
	remoteAddr := clientConn.RemoteAddr().String()
	
	// FASE 1: HANDSHAKE PQC (Kyber-768)
	pk, sk, err := crypto_core.GenerateKeyPair()
	if err != nil {
		log.Println("KeyGen Error", err)
		return
	}
	if _, err := clientConn.Write(pk); err != nil {
		return
	}

	cfg := crypto_core.GetConfig()
	ciphertext := make([]byte, cfg.CipherLen)
	if _, err := io.ReadFull(clientConn, ciphertext); err != nil {
		return
	}

	sharedSecret, err := crypto_core.Decapsulate(ciphertext, sk)
	if err != nil {
		log.Printf("[%s] 🚨 Ataque detectado!", remoteAddr)
		return
	}
	
	aesKey, _ := crypto_core.DeriveKey(sharedSecret, "pqc-tunnel-v1")
	log.Printf("[%s] 🔐 Túnel PQC OK. Buscando destino...", remoteAddr)

	// ====================================================
	// FASE 2: ROTEAMENTO DINÂMICO (Critical Update)
	// ====================================================
	
	// Pega o destino atual da memória (atualizado pelo Agente SaaS)
	currentTarget := config.GetTarget()
	
	targetConn, err := net.Dial("tcp", currentTarget)
	if err != nil {
		log.Printf("[%s] ❌ Falha ao conectar no alvo (%s): %v", remoteAddr, currentTarget, err)
		return
	}
	defer targetConn.Close()
	
	log.Printf("[%s] ➡️ Encaminhando para: %s", remoteAddr, currentTarget)

	// FASE 3: PIPELINE DE DADOS (AES-256)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		crypto_core.DecryptStream(aesKey, clientConn, targetConn)
		if tcpConn, ok := targetConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		crypto_core.EncryptStream(aesKey, targetConn, clientConn)
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
}