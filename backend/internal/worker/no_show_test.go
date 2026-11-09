package worker

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func receive(t *testing.T, calls <-chan int) int {
	t.Helper()
	select {
	case call := <-calls:
		return call
	case <-time.After(time.Second):
		t.Fatal("worker did not respond")
		return 0
	}
}

func TestWorkerStartsImmediatelyRetriesAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticks := make(chan time.Time)
	calls := make(chan int, 4)
	done := make(chan struct{})
	var logged string
	go func() {
		defer close(done)
		i := 0
		run(ctx, ticks, func(attempt context.Context) (int64, error) {
			if _, ok := attempt.Deadline(); !ok {
				panic("missing query deadline")
			}
			i++
			calls <- i
			if i == 1 {
				return 0, errors.New("database password must not appear in log")
			}
			return 1, nil
		}, func(format string, _ ...any) { logged = format })
	}()
	if receive(t, calls) != 1 {
		t.Fatal("initial sweep missing")
	}
	ticks <- time.Now()
	if receive(t, calls) != 2 {
		t.Fatal("error prevented retry")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	if strings.Contains(logged, "password") {
		t.Fatal("private database error logged")
	}
}

func TestWorkerCancellationReachesInFlightSweep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan int, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		run(ctx, make(chan time.Time), func(attempt context.Context) (int64, error) { started <- 1; <-attempt.Done(); return 0, attempt.Err() }, nil)
	}()
	receive(t, started)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("in-flight sweep ignored cancellation")
	}
}

func TestWorkerDoesNotStartWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, time.Minute, func(context.Context) (int64, error) { t.Fatal("sweep after cancellation"); return 0, nil }, nil); err != nil {
		t.Fatal(err)
	}
	if Run(context.Background(), 0, nil, nil) == nil {
		t.Fatal("invalid worker configuration accepted")
	}
}
