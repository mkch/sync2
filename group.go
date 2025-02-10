package sync2

import (
	"context"
	"errors"
	"slices"
	"sync"
)

// MutexGroup provides atomic locking and cancellation for a group of [Mutex] es.
// See [MutexGroup.LockAll].
//
// A Mutex must not be copied.
type MutexGroup struct {
	c        sync.Cond
	canceled error // The cause
}

// NewMutexGroup creates a new MutexGroup.
func NewMutexGroup() *MutexGroup {
	return &MutexGroup{c: sync.Cond{L: &sync.Mutex{}}}
}

// NewMutexGroupWithContext creates a new MutexGroup that is cancelled when ctx is done.
func NewMutexGroupWithContext(ctx context.Context) (g *MutexGroup) {
	g = NewMutexGroup()
	context.AfterFunc(ctx, func() { g.cancel(ctx.Err()) })
	return g
}

// NewMutex creates a mutex associated with group g.
func (g *MutexGroup) NewMutex() *Mutex {
	return &Mutex{g: g}
}

// ErrCanceled is the cause of cancellation
// when a mutext group is canceled by [MutexGroup.Cancel].
var ErrCanceled = errors.New("mutex group canceled")

// Cancel tells g to cancel with cause [ErrCanceled].
// Cancel does not wait for g to cancel.
// Cancel may be called by multiple goroutines simultaneously.
// After the first call, subsequent calls to Cancel do nothing.
func (g *MutexGroup) Cancel() {
	g.cancel(ErrCanceled)
}

// cancel cancels g.
// After the first call, subsequent calls do nothing.
// Cause must not be nil.
func (g *MutexGroup) cancel(cause error) {
	g.c.L.Lock()
	defer g.c.L.Unlock()
	if g.canceled != nil {
		return
	}
	g.canceled = cause
	g.c.Broadcast()
}

// LockAll atomically locks all mutexes in ms.
// If any mutex in ms is already locked, the calling goroutine blocks until
// all mutexes are available.
// If g is canceled, LockAll returns the cause.
// It panics if ms is empty or if any mutex in ms was not created by g.
func (g *MutexGroup) LockAll(ms ...*Mutex) error {
	if len(ms) == 0 {
		panic("no mutex to lock")
	}
	g.c.L.Lock()
	defer g.c.L.Unlock()
	if slices.ContainsFunc(ms, func(m *Mutex) bool { return m.g != g }) {
		panic("mutex not created by this group")
	}

	for {
		if g.canceled != nil {
			return g.canceled
		}
		if slices.ContainsFunc(ms, func(m *Mutex) bool { return m.locked }) {
			g.c.Wait()
		} else {
			break
		}
	}

	for _, m := range ms {
		m.locked = true
	}

	return nil
}

// A Mutex is a mutual exclusion lock that is part of a [MutexGroup].
//
// A Mutex must not be copied.
//
// Mutex has the same locking and unlocking behavior as [sync.Mutex], with the
// added ability to interact with [MutexGroup].
type Mutex struct {
	noCopy noCopy
	g      *MutexGroup
	locked bool
}

// Lock locks m as [sync.Mutex.Lock] does.
// If the associated [MutexGroup] is canceled, Lock returns the cause.
func (m *Mutex) Lock() error {
	m.g.c.L.Lock()
	defer m.g.c.L.Unlock()
	for {
		if m.g.canceled != nil {
			return m.g.canceled
		}
		if m.locked {
			m.g.c.Wait()
		} else {
			break
		}
	}
	m.locked = true
	m.g.c.Broadcast()
	return nil
}

// Unlock unlocks m.
// It is a run-time error if m is not locked.
// If the associated [MutexGroup] is canceled, Unlock returns the cause.
func (m *Mutex) Unlock() error {
	m.g.c.L.Lock()
	defer m.g.c.L.Unlock()
	if m.g.canceled != nil {
		return m.g.canceled
	}
	if !m.locked {
		panic("mutex not locked")
	}
	m.locked = false
	m.g.c.Broadcast()
	return nil
}

// noCopy may be added to structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
//
// Note that it must not be embedded, due to the Lock and Unlock methods.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
