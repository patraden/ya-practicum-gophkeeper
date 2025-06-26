package keystore

import (
	"fmt"
	"sync"

	"github.com/awnumar/memguard"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
)

// Keystore defines the interface for secret key storage (e.g. REK).
type Keystore interface {
	Load(secret []byte) error
	Get() ([]byte, error)
	Wipe()
	IsLoaded() bool
}

// InMemoryKeystore is a secure in-memory REK store.
type InMemoryKeystore struct {
	mu     sync.RWMutex
	rek    *memguard.LockedBuffer
	loaded bool
}

// NewInMemoryKeystore creates new empty instance of InMemoryKeystore.
func NewInMemoryKeystore() *InMemoryKeystore {
	return &InMemoryKeystore{
		mu:     sync.RWMutex{},
		rek:    nil,
		loaded: false,
	}
}

// Load sets the REK once securely.
func (ks *InMemoryKeystore) Load(secret []byte) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.loaded && ks.rek != nil {
		return fmt.Errorf("[%w] key store", e.ErrConflict)
	}

	buf := memguard.NewBufferFromBytes(secret)
	ks.rek = buf
	ks.loaded = true

	return nil
}

// Get returns a copy of the REK bytes safely.
func (ks *InMemoryKeystore) Get() ([]byte, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	if !ks.loaded || ks.rek == nil {
		return nil, fmt.Errorf("[%w] key store", e.ErrNotReady)
	}

	buf := ks.rek.Bytes()
	cpy := make([]byte, len(buf))

	copy(cpy, buf)

	return cpy, nil
}

// IsLoaded returns true if a REK is loaded.
// KeyStore usage should guarantee that loaded key is valid.
// For example, during unsealing process prior to storing the key
// it will be validated against expected key hash in pg.
//
// Method should be simple and performant as it will be heavily used
// by gRPC interceptor on every request to the server.
func (ks *InMemoryKeystore) IsLoaded() bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	return ks.loaded && ks.rek != nil
}

// Wipe securely zeroes out and destroys the REK buffer.
func (ks *InMemoryKeystore) Wipe() {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.rek != nil {
		ks.rek.Destroy()
		ks.rek = nil
	}

	ks.loaded = false
}
