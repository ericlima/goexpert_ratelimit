package limiter

import (
	"context"
	"sync"
	"time"
)

// MemoryClient estrutura para gerenciar limites em memória.
type MemoryClient struct {
	mu       sync.Mutex
	counters map[string]*memoryEntry
}

// memoryEntry representa os dados associados a uma chave (IP ou token).
type memoryEntry struct {
	count     int
	expireAt  time.Time
	blockedAt *time.Time
}

// NewMemoryClient cria uma nova instância de MemoryClient.
func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		counters: make(map[string]*memoryEntry),
	}
}

// IncrementKey incrementa o contador para uma chave específica.
func (m *MemoryClient) IncrementKey(ctx context.Context, key string, expiration time.Duration) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.counters[key]
	if !exists || time.Now().After(entry.expireAt) {
		m.counters[key] = &memoryEntry{
			count:    1,
			expireAt: time.Now().Add(expiration),
		}
		return 1, nil
	}

	entry.count++
	return entry.count, nil
}

// BlockKey bloqueia uma chave por um tempo específico.
func (m *MemoryClient) BlockKey(ctx context.Context, key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	blockedAt := now.Add(duration)
	m.counters[key] = &memoryEntry{
		count:     0,
		expireAt:  now,
		blockedAt: &blockedAt,
	}
	return nil
}

// IsBlocked verifica se uma chave está bloqueada.
func (m *MemoryClient) IsBlocked(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.counters[key]
	if !exists || entry.blockedAt == nil {
		return false, nil
	}

	return time.Now().Before(*entry.blockedAt), nil
}
