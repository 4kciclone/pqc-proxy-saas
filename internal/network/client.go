package network

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/4kciclone/pqc-proxy/pkg/crypto_core"
)

// StartClient inicia um servidor local que encapsula o tráfego e manda pro proxy remoto
func StartClient(proxyAddress string) {
	localPort := ":8080"
	listener, err := net.Listen("tcp", localPort)
	if err != nil {
		log.Fatalf("Erro ao ouvir localmente: %v", err)
	}
	
	fmt.Printf("💻 Cliente PQC (Sidecar) rodando em localhost%s\n", localPort)
	fmt.Printf("🚀 Todo tráfego enviado aqui será encriptado e mandado para %s\n", proxyAddress)

	for {
		// Aceita conexões da aplicação local (ex: Curl, Browser, App Legado)
		appConn, err := listener.Accept()
		if err != nil {
			log.Printf("Erro no accept local: %v", err)
			continue
		}
		// Para cada conexão local, abrimos um túnel PQC novo
		go handleAppConnection(appConn, proxyAddress)
	}
}

func handleAppConnection(appConn net.Conn, proxyAddress string) {
	defer appConn.Close()

	// 1. Conectar ao Proxy PQC Remoto (Servidor)
	proxyConn, err := net.Dial("tcp", proxyAddress)
	if err != nil {
		log.Printf("Falha ao conectar no Proxy PQC (%s): %v", proxyAddress, err)
		return
	}
	defer proxyConn.Close()

	// 2. Handshake PQC (Cliente recebe PK -> Envia Ciphertext)
	cfg := crypto_core.GetConfig()
	
	// Ler PK do servidor
	pk := make([]byte, cfg.PubLen)
	if _, err := io.ReadFull(proxyConn, pk); err != nil {
		log.Printf("Erro ao receber PK: %v", err)
		return
	}
	
	// Gerar Segredo e Ciphertext
	ciphertext, sharedSecret, err := crypto_core.Encapsulate(pk)
	if err != nil {
		log.Printf("Erro Encapsulate: %v", err)
		return
	}

	// Enviar Ciphertext pro servidor
	if _, err := proxyConn.Write(ciphertext); err != nil {
		log.Printf("Erro ao enviar Ciphertext: %v", err)
		return
	}

	// Derivar a mesma chave AES
	aesKey, err := crypto_core.DeriveKey(sharedSecret, "pqc-tunnel-v1")
	if err != nil {
		log.Printf("Erro DeriveKey: %v", err)
		return
	}

	// 3. Transferência de Dados (Bi-direcional)
	var wg sync.WaitGroup
	wg.Add(2)

	// App Local -> Encripta (AES) -> Proxy Remoto
	go func() {
		defer wg.Done()
		err := crypto_core.EncryptStream(aesKey, appConn, proxyConn)
		if err != nil && err != io.EOF {
			log.Printf("Erro Cliente->Proxy: %v", err)
		}
		if tcpConn, ok := proxyConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	// Proxy Remoto -> Decripta (AES) -> App Local
	go func() {
		defer wg.Done()
		err := crypto_core.DecryptStream(aesKey, proxyConn, appConn)
		if err != nil && err != io.EOF {
			log.Printf("Erro Proxy->Cliente: %v", err)
		}
		if tcpConn, ok := appConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
}