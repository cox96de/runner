package mock

import (
	"context"
	"sync"
	"time"

	"github.com/cox96de/runner/lib"
)

// Locker is a mock implementation of lib.Locker.
// It is used for testing.
type Locker struct {
	data map[string]*lockData
	lock sync.Mutex
}

type lockData struct {
	value string
	exp   time.Time
}

func (m *Locker) Lock(ctx context.Context, key, value string, expiresIn time.Duration) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if d, ok := m.data[key]; ok {
		if d.exp.After(time.Now()) {
			return false, nil
		}
	}
	m.data[key] = &lockData{value: value, exp: time.Now().Add(expiresIn)}
	return true, nil
}

func (m *Locker) Unlock(ctx context.Context, key string) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if d, ok := m.data[key]; ok {
		if d.exp.Before(time.Now()) {
			return false, nil
		}
		delete(m.data, key)
		return true, nil
	}
	return false, nil
}

func NewMockLocker() lib.Locker {
	return &Locker{data: map[string]*lockData{}}
}
