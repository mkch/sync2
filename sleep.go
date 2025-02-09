package sync2

import (
	"context"
	"sync"
	"time"
)

// Sleep behaves like [time.Sleep], but is cancellable using the provided context.
// It returns ctx.Err() if the context is done before the d elapses.
func Sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	ch := make(chan error, 2)
	timer := time.AfterFunc(d, func() {
		ch <- nil
	})
	context.AfterFunc(ctx, func() {
		timer.Stop()
		ch <- ctx.Err()
	})

	return <-ch
}

// sleepWithCond is the sync.Cond version of [Sleep].
// No significant performance benefit though.
func sleepWithCond(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	type Flag int
	const (
		Sleeping Flag = iota
		Timedout
		Canceled
	)
	var flag Flag
	var c = sync.NewCond(&sync.Mutex{})
	timer := time.AfterFunc(d, func() {
		c.L.Lock()
		defer c.L.Unlock()
		flag = Timedout
		c.Signal()
	})
	context.AfterFunc(ctx, func() {
		timer.Stop()
		c.L.Lock()
		defer c.L.Unlock()
		flag = Canceled
		c.Signal()
	})

	c.L.Lock()
	defer c.L.Unlock()
	for flag == Sleeping {
		c.Wait()
	}
	if flag == Canceled {
		return ctx.Err()
	}
	return nil
}
