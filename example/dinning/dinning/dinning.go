package dinning

import (
	"context"
	"sync"
	"time"

	"github.com/mkch/sync2"
)

const N = 5                                    // Number of philosophers(chopsticks).
const eatingDuration = time.Millisecond * 1000 // The duration of each eating for the philosopher.

// StickState is the state of chopsticks.
type StickState int

const (
	OnTable     StickState = iota // The chopstick is on table
	HoldByLeft                    // The chopstick is hold by the philosopher to the left.
	HoldByRight                   // The chopstick is hold by the philosopher to the right.
)

// ChangeStick must be set to a non-nil function before calling
// any function in this package.
var ChangeStick func(i int, stick StickState)

// OnDinningChanged must be set to a non-nil function to receive
// dinning state.
var OnDinningChanged func(dinning bool)

// Running flag. True after [Mutex] or [MutexGroup] is called.
var dinning bool

// Lock for dinning.
var dinningLock sync.RWMutex

// context and cancel function used by [Mutex] and [MutexGroup].
var ctx context.Context
var cancel func()

// Semaphore of all running goroutines(philosophers).
var wg sync.WaitGroup

// Dinning returns the dinning state.
func Dinning() bool {
	dinningLock.RLock()
	defer dinningLock.RUnlock()
	return dinning
}

// sticks returns the indics of chopsticks of philosopher i.
func sticks(i int) (left, right int) {
	left = i
	right = i + 1
	if right == N {
		right = 0
	}
	return
}

// cancelled returns whether dinning is cancelled by [Stop].
func canceled() bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// Mutex runs the demo of Dining Philosophers Problem with [sync2.Mutex].
func Mutex() {
	dinningLock.Lock()
	defer dinningLock.Unlock()
	if dinning {
		return
	}
	dinning = true
	ctx, cancel = context.WithCancel(context.Background())
	g := sync2.NewMutexGroupWithContext(ctx)
	var mux [N]*sync2.Mutex
	for i := range mux {
		mux[i] = g.NewMutex()
	}
	wg.Add(N)
	for i := range N {
		go func() {
			defer wg.Done()
			for !canceled() {
				left, right := sticks(i)
				if mux[left].Lock() == nil {
					ChangeStick(left, HoldByRight)
				}
				if mux[right].Lock() == nil {
					ChangeStick(right, HoldByLeft)
				}
				sync2.Sleep(ctx, eatingDuration)
				if mux[left].Unlock() == nil {
					ChangeStick(left, OnTable)
				}
				if mux[right].Unlock() == nil {
					ChangeStick(right, OnTable)
				}
			}
		}()
	}
	go OnDinningChanged(dinning)
}

// Mutex runs the demo of Dining Philosophers Problem with [sync2.MutexGroup].
func MutexGroup() {
	dinningLock.Lock()
	defer dinningLock.Unlock()
	if dinning {
		return
	}
	dinning = true
	ctx, cancel = context.WithCancel(context.Background())
	var group = sync2.NewMutexGroupWithContext(ctx)
	var mux [N]*sync2.Mutex
	for i := range mux {
		mux[i] = group.NewMutex()
	}
	wg.Add(N)
	for i := range N {
		go func() {
			defer wg.Done()
			for !canceled() {
				left, right := sticks(i)
				if group.LockAll(mux[left], mux[right]) == nil {
					ChangeStick(left, HoldByRight)
					ChangeStick(right, HoldByLeft)
				}
				sync2.Sleep(ctx, eatingDuration)
				if mux[left].Unlock() == nil {
					ChangeStick(left, OnTable)
				}

				if mux[right].Unlock() == nil {
					ChangeStick(right, OnTable)
				}
			}
		}()
	}
	go OnDinningChanged(dinning)
}

// Stop stops running demo of Dining Philosophers Problem.
// Stop blocks until all philosopher goroutine exit.
func Stop() {
	dinningLock.Lock()
	defer dinningLock.Unlock()
	if !dinning {
		return
	}
	cancel()
	go func() {
		wg.Wait()
		dinning = false
		OnDinningChanged(dinning)
	}()
}
