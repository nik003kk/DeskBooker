// Package worker runs bounded background maintenance without overlapping sweeps.
package worker

import (
	"context"
	"errors"
	"time"
)

const NoShowInterval = time.Minute
const sweepTimeout = 30 * time.Second

type Sweep func(context.Context) (int64, error)
type Logf func(string, ...any)

// Run checks immediately on startup and then once per interval. The caller
// must cancel ctx and join Run before closing its database pool.
func Run(ctx context.Context, interval time.Duration, sweep Sweep, logf Logf) error {
	if interval <= 0 || sweep == nil {
		return errors.New("invalid worker configuration")
	}
	timer := time.NewTicker(interval)
	defer timer.Stop()
	run(ctx, timer.C, sweep, logf)
	return nil
}

func run(ctx context.Context, ticks <-chan time.Time, sweep Sweep, logf Logf) {
	for {
		if ctx.Err() != nil {
			return
		}
		attempt, cancel := context.WithTimeout(ctx, sweepTimeout)
		count, err := sweep(attempt)
		cancel()
		if ctx.Err() != nil {
			return
		}
		if logf != nil {
			if err != nil {
				logf("No-show worker: sweep failed; retrying on next tick")
			} else if count > 0 {
				logf("No-show worker: processed %d reservations", count)
			}
		}
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticks:
			if !ok {
				return
			}
		}
	}
}
