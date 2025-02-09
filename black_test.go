package sync2_test

import (
	"context"
	"testing"
	"time"

	"github.com/mkch/sync2"
)

func TestMutex(t *testing.T) {
	g := sync2.NewMutexGroup()
	m := g.NewMutex()

	var n = 0
	m.Lock()
	go func() {
		defer m.Unlock()
		n = 100
	}()
	m.Lock()
	if n != 100 {
		t.Fatal()
	}
}

func TestLockAll(t *testing.T) {
	g := sync2.NewMutexGroup()
	m1 := g.NewMutex()
	m2 := g.NewMutex()
	m1.Lock()
	m2.Lock()
	t0 := time.Now()
	go func() {
		time.Sleep(time.Millisecond * 10)
		m1.Unlock()
	}()
	go func() {
		time.Sleep(time.Millisecond * 20)
		m2.Unlock()
	}()
	g.LockAll(m1, m2)
	if time.Since(t0) < time.Millisecond*20 {
		t.Fatal()
	}
}

func TestMutexCancel(t *testing.T) {
	g := sync2.NewMutexGroup()
	m := g.NewMutex()
	if err := m.Lock(); err != nil {
		t.Fatal(err)
	}
	go g.Cancel()
	if err := m.Lock(); err != sync2.ErrCanceled {
		t.Fatal(err)
	}
	// All locking methods should return the same error
	if err := m.Lock(); err != sync2.ErrCanceled {
		t.Fatal(err)
	}
	if err := m.Unlock(); err != sync2.ErrCanceled {
		t.Fatal(err)
	}
	if err := g.LockAll(m, g.NewMutex()); err != sync2.ErrCanceled {
		t.Fatal(err)
	}
}

func TestMutexCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	g := sync2.NewMutexGroupWithContext(ctx)
	m1 := g.NewMutex()
	m2 := g.NewMutex()
	m1.Lock()
	m2.Lock()
	cancel()
	if err := g.LockAll(m1, m2); err != context.Canceled {
		t.Fatal(err)
	}
}

func TestSleep(t *testing.T) {
	const d = time.Millisecond * 200
	start := time.Now()
	if err := sync2.Sleep(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	diff := time.Since(start) - d
	if diff < 0 || diff > d/3 {
		t.Fatalf("inaccurate sleep %v", diff)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	if err := sync2.Sleep(ctx, time.Second); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}
