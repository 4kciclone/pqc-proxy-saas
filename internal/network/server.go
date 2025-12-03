package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/4kciclone/pqc-proxy/pkg/crypto_core"
)

// TARGET_SERVER define para onde o proxy encaminha o tráfego após decriptar.
// Em produção, isso seria seu Banco de Dados (ex: "localhost:5432") ou Mainframe.
// Para teste, usaremos o Google.
const TARGET_SERVER = "google.com:80"

// StartServer inicia o listener TCP PQC
func StartServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
	defer listener.Close()

	fmt.Printf("🛡️  PQC Proxy (Server) ouvindo em 0.0.0.0:%s\n", port)
	fmt.Printf("🎯 Target configurado: %s\n", TARGET_SERVER)
	fmt.Println("   Aguardando conexões...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Erro no accept: %v", err)
			continue
		}
		// Processa cada conexão em uma Goroutine
		go handlePQCConnection(conn)
	}
}

func handlePQCConnection(clientConn net.Conn) {
	defer clientConn.Close()
	remoteAddr := clientConn.RemoteAddr().String()
	log.Printf("[%s] Nova conexão PQC iniciada.", remoteAddr)

	// ====================================================
	// FASE 1: HANDSHAKE PQC (Kyber-768)
	// ====================================================
	
	// 1. Gerar Par de Chaves Epêmeras
	pk, sk, err := crypto_core.GenerateKeyPair()
	if err != nil {
		log.Printf("[%s] Erro KeyGen: %v", remoteAddr, err)
		return
	}

	// 2. Enviar Chave Pública para o Cliente
	if _, err := clientConn.Write(pk); err != nil {
		log.Printf("[%s] Erro ao enviar PK: %v", remoteAddr, err)
		return
	}

	// 3. Receber Ciphertext do Cliente
	cfg := crypto_core.GetConfig()
	ciphertext := make([]byte, cfg.CipherLen)
	if _, err := io.ReadFull(clientConn, ciphertext); err != nil {
		log.Printf("[%s] Erro ao ler Ciphertext: %v", remoteAddr, err)
		return
	}

	// 4. Decapsular para obter o Segredo Compartilhado
	start := time.Now()
	sharedSecret, err := crypto_core.Decapsulate(ciphertext, sk)
	if err != nil {
		log.Printf("[%s] 🚨 FALHA CRÍTICA: Ataque ou erro de chave detectado!", remoteAddr)
		return
	}
	
	// 5. Derivar chave AES-256 usando HKDF
	aesKey, err := crypto_core.DeriveKey(sharedSecret, "pqc-tunnel-v1")
	if err != nil {
		log.Printf("Erro ao derivar chave: %v", err)
		return
	}

	log.Printf("[%s] 🔐 Túnel Quântico estabelecido (%v). Iniciando AES stream...", remoteAddr, time.Since(start))

	// ====================================================
	// FASE 2: CONEXÃO COM O LEGADO (UPSTREAM)
	// ====================================================
	
	targetConn, err := net.Dial("tcp", TARGET_SERVER)
	if err != nil {
		log.Printf("[%s] Falha ao conectar no servidor alvo (%s): %v", remoteAddr, TARGET_SERVER, err)
		return
	}
	defer targetConn.Close()

	// ====================================================
	// FASE 3: PIPELINE DE DADOS (Bi-direcional)
	// ====================================================
	
	var wg sync.WaitGroup
	wg.Add(2)

	// Pipeline A: Cliente (Encriptado) -> Proxy (Decripta) -> Target (Texto Plano)
	go func() {
		defer wg.Done()
		err := crypto_core.DecryptStream(aesKey, clientConn, targetConn)
		if err != nil && err != io.EOF {
			log.Printf("[%s] Erro DecryptStream: %v", remoteAddr, err)
		}
		// Fecha a escrita do target para sinalizar fim do stream
		if tcpConn, ok := targetConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// Pipeline B: Target (Texto Plano) -> Proxy (Encripta) -> Cliente (Encriptado)
	go func() {
		defer wg.Done()
		err := crypto_core.EncryptStream(aesKey, targetConn, clientConn)
		if err != nil && err != io.EOF {
			log.Printf("[%s] Erro EncryptStream: %v", remoteAddr, err)
		}
		// Fecha a escrita do cliente
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
	log.Printf("[%s] Conexão encerrada.", remoteAddr)
}