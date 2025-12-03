package config

import "sync"

var (
	mutex         sync.RWMutex
	// Valor padrão inicial (Google)
	CurrentTarget = "google.com:80" 
)

// SetTarget atualiza o destino de forma segura (Thread-Safe)
func SetTarget(target string) {
	mutex.Lock()
	defer mutex.Unlock()
	CurrentTarget = target
}

// GetTarget lê o destino atual de forma segura (usado por milhares de conexões)
func GetTarget() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return CurrentTarget
}