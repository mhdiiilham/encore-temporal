package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

// IDProvider defines methods for generating unique identifiers.
type IDProvider interface {
	GenerateBillingID(prefix string) string
	GenerateIdempotencyKey(prefix string, payload []byte) string
}

// IDGenerator provides methods for generating unique IDs.
type IDGenerator struct {
	r       *rand.Rand
	counter uint64
}

// NewIDGenerator creates a new generator.
// If seed == 0, it uses time-based seed (good for production).
func NewIDGenerator(seed int64) IDProvider {
	var src rand.Source
	if seed == 0 {
		src = rand.NewSource(time.Now().UnixNano())
	} else {
		src = rand.NewSource(seed)
	}
	return &IDGenerator{r: rand.New(src)}
}

// GenerateBillingID creates a unique booking ID.
// Format: PREFIX-YYYYMMDD-COUNTER-RANDOM
func (g *IDGenerator) GenerateBillingID(prefix string) string {
	date := time.Now().Format("20060102")
	cnt := atomic.AddUint64(&g.counter, 1)
	randomPart := g.r.Intn(1_000_000)
	return fmt.Sprintf("%s-%s-%06d-%06d", prefix, date, cnt, randomPart)
}

// GenerateIdempotencyKey creates a stable idempotency key from request data.
// Combines a prefix, timestamp, and a SHA-256 hash of the payload.
func (g *IDGenerator) GenerateIdempotencyKey(prefix string, payload []byte) string {
	hash := sha256.Sum256(payload)
	ts := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d-%s", prefix, ts, hex.EncodeToString(hash[:8]))
}
