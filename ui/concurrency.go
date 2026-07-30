// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"context"
	"runtime"
	"sync/atomic"
)

// ─── Global worker pool ───────────────────────────────────────────────────────
//
// Gru's worker pool provides a lightweight way to offload CPU-bound or I/O-
// bound work (image decoding, JSON parsing, data transforms) without touching
// raylib's main thread. The pool is sized to runtime.NumCPU() workers, which
// automatically scales to the host hardware.
//
// Rules:
//   - Never call raylib / OpenGL APIs from a worker.
//   - Never mutate widget state directly from a worker.
//   - Use doc.QueueMain(fn) to push results back to the main thread.
//
// See SubmitJob, SubmitAsync, and SubmitAsyncBg for usage.

var (
	// jobQueue is the worker pool input channel.
	// Buffered at 128 to absorb bursts without blocking the caller.
	jobQueue chan func()

	// activeJobs is the number of jobs currently executing in workers.
	activeJobs int64
)

func init() {
	n := runtime.NumCPU()
	if n < 2 {
		n = 2
	}
	jobQueue = make(chan func(), 128)
	for i := 0; i < n; i++ {
		go runWorker()
	}
}

func runWorker() {
	for fn := range jobQueue {
		atomic.AddInt64(&activeJobs, 1)
		fn()
		atomic.AddInt64(&activeJobs, -1)
	}
}

// SubmitJob submits fn to the worker pool and returns a buffered channel that
// will receive the single result. The channel has capacity 1, so the worker
// never blocks writing even if the caller reads lazily.
//
// If ctx is already cancelled when the job reaches a worker, fn is skipped
// and the channel is closed without sending a value — use the two-value receive
// form to detect this:
//
//	ch := ui.SubmitJob(ctx, expensiveCompute)
//	// ... later ...
//	if result, ok := <-ch; ok {
//	    doc.QueueMain(func() { label.Text.Set(result) })
//	}
func SubmitJob[T any](ctx context.Context, fn func() T) <-chan T {
	ch := make(chan T, 1)
	jobQueue <- func() {
		if ctx.Err() != nil {
			close(ch)
			return
		}
		result := fn()
		select {
		case ch <- result:
		case <-ctx.Done():
		}
	}
	return ch
}

// SubmitAsync submits fn to the worker pool without returning a result channel.
// Use this for fire-and-forget work that pushes its own result back via
// doc.QueueMain. If the pool queue is full, fn falls back to a new goroutine.
//
// fn is not called if ctx is already cancelled.
func SubmitAsync(ctx context.Context, fn func()) {
	wrapped := func() {
		if ctx.Err() != nil {
			return
		}
		fn()
	}
	select {
	case jobQueue <- wrapped:
	default:
		go wrapped()
	}
}

// SubmitAsyncBg is a convenience wrapper around SubmitAsync with
// context.Background() — the job is never cancelled by the caller.
// Use this for simple fire-and-forget jobs such as image loading.
func SubmitAsyncBg(fn func()) {
	SubmitAsync(context.Background(), fn)
}

// ActiveJobCount returns the number of worker-pool jobs currently executing.
// Useful for monitoring in the Inspector or perf overlay.
func ActiveJobCount() int { return int(atomic.LoadInt64(&activeJobs)) }

// WorkerCount returns the number of goroutines in the global worker pool
// (equal to runtime.NumCPU(), minimum 2).
func WorkerCount() int {
	n := runtime.NumCPU()
	if n < 2 {
		n = 2
	}
	return n
}
