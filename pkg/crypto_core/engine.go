package crypto_core

/*
#cgo CFLAGS: -I${SRCDIR}
// Abaixo: Linkamos com nossa lib Rust, com a liboqs e com as dependencias de sistema (ssl, crypto, m)
#cgo LDFLAGS: -L${SRCDIR}/../../rust-core/target/release -lpqc_core -L/usr/local/lib -loqs -lcrypto -lm
#include "bridge.h"
*/
import "C"
import (
	"errors"
	"unsafe"
)

// PQCConfig segura os tamanhos dos buffers
type PQCConfig struct {
	PubLen    int
	PrivLen   int
	CipherLen int
	SharedLen int
}

// GetConfig retorna os tamanhos necessários para o Kyber-768
func GetConfig() PQCConfig {
	return PQCConfig{
		PubLen:    int(C.get_public_key_len()),
		PrivLen:   int(C.get_secret_key_len()),
		CipherLen: int(C.get_ciphertext_len()),
		SharedLen: int(C.get_shared_secret_len()),
	}
}

// GenerateKeyPair cria um par de chaves Kyber (Pública e Privada)
func GenerateKeyPair() ([]byte, []byte, error) {
	cfg := GetConfig()
	
	// Alocamos memória no Go
	pubKey := make([]byte, cfg.PubLen)
	privKey := make([]byte, cfg.PrivLen)

	// Passamos ponteiros para o Rust preencher
	// unsafe.Pointer(&pubKey[0]) pega o endereço do primeiro byte do slice
	res := C.pqc_keypair(
		(*C.uchar)(unsafe.Pointer(&pubKey[0])),
		(*C.uchar)(unsafe.Pointer(&privKey[0])),
	)

	if res != 0 {
		return nil, nil, errors.New("falha ao gerar par de chaves PQC")
	}

	return pubKey, privKey, nil
}

// Encapsulate (Lado Cliente): Gera segredo compartilhado e ciphertext
func Encapsulate(pubKey []byte) (ciphertext []byte, sharedSecret []byte, err error) {
	cfg := GetConfig()
	if len(pubKey) != cfg.PubLen {
		return nil, nil, errors.New("tamanho de chave pública inválido")
	}

	ct := make([]byte, cfg.CipherLen)
	ss := make([]byte, cfg.SharedLen)

	res := C.pqc_encaps(
		(*C.uchar)(unsafe.Pointer(&pubKey[0])),
		(*C.uchar)(unsafe.Pointer(&ct[0])),
		(*C.uchar)(unsafe.Pointer(&ss[0])),
	)

	if res != 0 {
		return nil, nil, errors.New("falha no encapsulamento PQC")
	}

	return ct, ss, nil
}

// Decapsulate (Lado Servidor): Abre o ciphertext e recupera o segredo
func Decapsulate(ciphertext []byte, privKey []byte) (sharedSecret []byte, err error) {
	cfg := GetConfig()
	if len(ciphertext) != cfg.CipherLen || len(privKey) != cfg.PrivLen {
		return nil, errors.New("tamanhos de input inválidos")
	}

	ss := make([]byte, cfg.SharedLen)

	res := C.pqc_decaps(
		(*C.uchar)(unsafe.Pointer(&ciphertext[0])),
		(*C.uchar)(unsafe.Pointer(&privKey[0])),
		(*C.uchar)(unsafe.Pointer(&ss[0])),
	)

	if res != 0 {
		return nil, errors.New("falha no desencapsulamento PQC (ataque detectado ou chave errada)")
	}

	return ss, nil
}