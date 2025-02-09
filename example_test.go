package sync2_test

import (
	"context"
	"fmt"
	"time"

	"github.com/mkch/sync2"
)

func ExampleMutexGroup_LockAll() {
	g := sync2.NewMutexGroup()
	m1, m2 := g.NewMutex(), g.NewMutex()
	m1.Lock()
	m2.Lock()
	go func() {
		m1.Unlock()
		fmt.Println("m1 unlocked")
		m2.Unlock()
		fmt.Println("m2 unlocked")
	}()
	g.LockAll(m1, m2)
	fmt.Println("atomically locked m1 and m2")
	// Output:
	// m1 unlocked
	// m2 unlocked
	// atomically locked m1 and m2
}

func ExampleMutexGroup_Cancel() {
	ctx, cancel := context.WithCancel(context.Background())
	g := sync2.NewMutexGroupWithContext(ctx)
	m := g.NewMutex()
	m.Lock()
	go cancel()
	fmt.Println(m.Lock()) // Won't block for long.
	// Output:
	// context canceled
}

func ExampleSleep() {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Millisecond)
	defer cancel()
	// It wont' sleep for an hour
	fmt.Println(sync2.Sleep(ctx, time.Hour))
	// Output:
	// context deadline exceeded
}
