package crypto_core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

// DeriveKey transforma o segredo compartilhado (Kyber) em uma chave AES-256 segura
func DeriveKey(sharedSecret []byte, context string) ([]byte, error) {
	// HKDF garante que a chave tenha entropia bem distribuída
	hkdf := hkdf.New(sha256.New, sharedSecret, nil, []byte(context))
	key := make([]byte, 32) // AES-256 precisa de 32 bytes
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}
	return key, nil
}

// SecureCopy lê de 'src', encripta e escreve em 'dst'
func EncryptStream(key []byte, src io.Reader, dst io.Writer) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	buffer := make([]byte, 32*1024) // Buffer de 32KB
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			// 1. Gera Nonce (Número usado uma vez)
			nonce := make([]byte, aesgcm.NonceSize())
			if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
				return err
			}

			// 2. Encripta o pedaço
			ciphertext := aesgcm.Seal(nil, nonce, buffer[:n], nil)

			// 3. Escreve o cabeçalho (Tamanho do pacote total)
			// Pacote = [Nonce (12)] + [Ciphertext (N + Tag)]
			packetLen := uint32(len(nonce) + len(ciphertext))
			if err := binary.Write(dst, binary.BigEndian, packetLen); err != nil {
				return err
			}

			// 4. Escreve o pacote
			if _, err := dst.Write(nonce); err != nil {
				return err
			}
			if _, err := dst.Write(ciphertext); err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// DecryptStream lê pacotes encriptados de 'src', decripta e escreve em 'dst'
func DecryptStream(key []byte, src io.Reader, dst io.Writer) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	for {
		// 1. Ler o tamanho do próximo pacote (4 bytes)
		var packetLen uint32
		err := binary.Read(src, binary.BigEndian, &packetLen)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		// 2. Ler o pacote inteiro (Nonce + Ciphertext)
		packet := make([]byte, packetLen)
		if _, err := io.ReadFull(src, packet); err != nil {
			return err
		}

		// 3. Separar Nonce e Ciphertext
		nonceSize := aesgcm.NonceSize()
		if len(packet) < nonceSize {
			return errors.New("pacote malformado")
		}
		nonce, ciphertext := packet[:nonceSize], packet[nonceSize:]

		// 4. Decriptar
		plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return errors.New("falha na decriptação: dados corrompidos ou ataque")
		}

		// 5. Escrever dados limpos
		if _, err := dst.Write(plaintext); err != nil {
			return err
		}
	}
}