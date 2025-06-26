package shamir

import (
	"fmt"
	"sync"

	"github.com/awnumar/memguard"
	"github.com/hashicorp/vault/shamir"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/rs/zerolog"
)

// Collector securely collects and manages Shamir's Secret Sharing shares
// in memory using memguard for sensitive share storage.
type Collector struct {
	mu        sync.Mutex
	shares    []*memguard.LockedBuffer
	threshold int
	log       zerolog.Logger
}

// NewCollector creates a new Collector with the specified threshold.
// The collector will attempt reconstruction only after collecting `threshold` shares.
func NewCollector(log zerolog.Logger) *Collector {
	return &Collector{
		mu:        sync.Mutex{},
		shares:    make([]*memguard.LockedBuffer, 0, ThresholdShares),
		threshold: ThresholdShares,
		log:       log,
	}
}

// Collect adds a new share to the collector.
// If a duplicate share is provided, it is ignored.
// If the threshold is already met, returns ErrConflict.
func (c *Collector) Collect(share []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.shares) >= c.threshold {
		return fmt.Errorf("[%w] enough pieces already", e.ErrConflict)
	}

	// Deduplicate by content
	for _, s := range c.shares {
		if s.EqualTo(share) {
			c.log.Info().
				Int("threshold", c.threshold).
				Msg("shamir's share collected previously")

			return nil
		}
	}

	c.shares = append(c.shares, memguard.NewBufferFromBytes(share))

	return nil
}

// IsThresholdMet returns true if the threshold number of shares has been collected.
func (c *Collector) IsThresholdMet() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.shares) >= c.threshold
}

// Size returns the current number of collected shares.
func (c *Collector) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.shares)
}

// StatusMessage returns a human-readable status of how many shares are collected.
func (c *Collector) StatusMessage() string {
	return fmt.Sprintf("Collected %d out of %d root key pieces", c.Size(), c.threshold)
}

// Reconstruct attempts to reconstruct the original secret from collected shares.
// Returns ErrNotReady if the threshold has not been met.
func (c *Collector) Reconstruct() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.shares) < c.threshold {
		return nil, fmt.Errorf("[%w] not enough pieces", e.ErrNotReady)
	}

	rawShares := make([][]byte, len(c.shares))
	for i, buf := range c.shares {
		rawShares[i] = buf.Bytes()
	}

	rek, err := shamir.Combine(rawShares)
	if err != nil {
		c.log.Error().Err(err).
			Int("threshold", c.threshold).
			Msg("shamir's combine")

		return nil, e.InternalErr(err)
	}

	return rek, nil
}

// Reset securely wipes all collected shares from memory and resets the collector state.
func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, buf := range c.shares {
		buf.Destroy()
	}

	c.shares = nil
}
