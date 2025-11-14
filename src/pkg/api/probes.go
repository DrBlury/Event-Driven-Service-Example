package api

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// PingFunc represents a health check that returns an error when the resource is unavailable.
type PingFunc func(ctx context.Context) error

// NewPingProbe wraps a PingFunc with standardised error handling suitable for InfoHandler probes.
func NewPingProbe(name string, fn PingFunc) ProbeFunc {
	return func(ctx context.Context) error {
		if fn == nil {
			return fmt.Errorf("%s probe: ping function is nil", name)
		}
		if ctx == nil {
			ctx = context.Background()
		}

		if err := fn(ctx); err != nil {
			return fmt.Errorf("%s probe failed: %w", name, err)
		}
		return nil
	}
}

// MongoPinger captures the subset of the MongoDB client used for readiness checks.
type MongoPinger interface {
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

// NewMongoPingProbe creates a ProbeFunc that pings MongoDB using the provided client.
// If readPref is nil it defaults to readpref.Primary.
func NewMongoPingProbe(client MongoPinger, readPref *readpref.ReadPref) ProbeFunc {
	return func(ctx context.Context) error {
		if client == nil {
			return errors.New("mongo probe: client is nil")
		}

		if ctx == nil {
			ctx = context.Background()
		}

		rp := readPref
		if rp == nil {
			rp = readpref.Primary()
		}

		if err := client.Ping(ctx, rp); err != nil {
			return fmt.Errorf("mongo probe failed: %w", err)
		}
		return nil
	}
}
