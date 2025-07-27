package keystore

// Keystore defines the interface for secret key storage (e.g. REK).
type Keystore interface {
	Load(secret []byte) error
	Get() ([]byte, error)
	Wipe()
	IsLoaded() bool
}
