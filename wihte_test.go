package sync2

import (
	"context"
	"errors"
	"testing"
	"time"
)

func testMutexCancel(t *testing.T, reason error) {
	g := NewMutexGroup()
	m := g.NewMutex()
	if err := m.Lock(); err != nil {
		t.Fatal(err)
	}
	go g.cancel(reason)
	if err := m.Lock(); err != reason {
		t.Fatal(err)
	}
	// All locking methods should return the same error
	if err := m.Lock(); err != reason {
		t.Fatal(err)
	}
	if err := m.Unlock(); err != reason {
		t.Fatal(err)
	}
	if err := g.LockAll(m, g.NewMutex()); err != reason {
		t.Fatal(err)
	}
}

func TestMutexCancel(t *testing.T) {
	testMutexCancel(t, ErrCanceled)
	testMutexCancel(t, errors.New("error1"))
}

func TestSleepWithCond(t *testing.T) {
	const d = time.Millisecond * 200
	start := time.Now()
	if err := sleepWithCond(context.Background(), d); err != nil {
		t.Fatal(err)
	}
	diff := time.Since(start) - d
	if diff < 0 || diff > d/3 {
		t.Fatalf("inaccurate sleep %v", diff)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	if err := sleepWithCond(ctx, time.Second); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}

func BenchmarkSleep(b *testing.B) {
	for range b.N {
		Sleep(context.Background(), 0)
		Sleep(context.Background(), 1)
		Sleep(context.Background(), time.Millisecond*10)
	}
}

func BenchmarkSleepWithCond(b *testing.B) {
	for range b.N {
		sleepWithCond(context.Background(), 0)
		sleepWithCond(context.Background(), 1)
		sleepWithCond(context.Background(), time.Millisecond*10)
	}
}

func BenchmarkTimeSleep(b *testing.B) {
	for range b.N {
		time.Sleep(0)
		time.Sleep(1)
		time.Sleep(time.Millisecond * 10)
	}
}
