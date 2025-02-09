# sync2

Synchronization Primitives Beyond Go's Standard Library.

1. Interrupted Sleep:

   ```go
    func ExampleSleep() {
        ctx, cancel := context.WithTimeout(context.Background(),
            time.Millisecond)
        defer cancel()
        // It wont' sleep for an hour
        fmt.Println(sync2.Sleep(ctx, time.Hour))
        // Output:
        // context deadline exceeded
    }
   ```

2. Cancellable mutex:

   ```go
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
   ```

3. Atomic locking of multiple mutexes:

   ```go
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
   ```

The full example is available here: [Dining philosophers problem](https://github.com/mkch/sync2/tree/main/example/dinning)

![Dining philosophers](https://github.com/mkch/sync2/tree/main/example/dinning/demo.png)
