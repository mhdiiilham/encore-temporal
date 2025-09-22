package temporalclient

import (
	"sync"

	"go.temporal.io/sdk/client"
)

var (
	once sync.Once
	err  error
	tc   client.Client
)

// GetTemporalClient returns a singleton Temporal client instance.
//
// It initializes the client once using sync.Once to ensure that only a single
// connection to the Temporal server is created, even if called multiple times
// from different services. Subsequent calls return the same client and error
// values from the first initialization attempt.
func GetTemporalClient(opts client.Options) (client.Client, error) {
	once.Do(func() {
		tc, err = client.Dial(opts)
	})

	return tc, err
}
