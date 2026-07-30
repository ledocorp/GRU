// Package ui provides reactive signals for fine-grained UI state management.
//
// # Signal[T]
//
// A Signal[T] is a generic observable value. Calling Set notifies all
// registered subscriber callbacks synchronously. Widgets expose their mutable
// state as Signal fields (e.g. Button.Text, Slider.Value) so that application
// code can read and write them reactively.
//
// # Effect
//
// NewEffect(fn) runs fn immediately while tracking which Signals are read
// inside it. Each read Signal automatically registers fn as a subscriber, so
// fn re-runs whenever any of those Signals change. This is the same dependency-
// tracking pattern used by SolidJS and MobX.
//
// Example — label tracks counter automatically:
//
//	count := ui.NewSignal(0)
//	label := ui.NewLabel("lbl", "0", ...)
//	ui.NewEffect(func() {
//	    label.Text.Set(fmt.Sprintf("Count: %d", count.Get()))
//	})
//	// Later: count.Set(count.Get()+1) → label updates immediately.
//
// # Subscriber Deduplication
//
// Subscribe uses a pointer-equality check to avoid adding the same function
// twice. Effect run functions are pointer-stable (stored in the Effect struct),
// so automatic dependency tracking via Get() never double-registers.
//
// # Debouncing
//
// SetDebounced(value, n) sets the value immediately but notifies subscribers
// only every n calls. This is useful for high-frequency signals like a Slider
// being dragged — the visual update fires every frame, but any expensive
// subscriber (e.g. a network request or layout recalculation) is called at
// most 1/n as often.
//
// # Thread Safety
//
// Signals are not thread-safe. All Signal reads and writes must happen on
// the main goroutine (the raylib render thread).
package ui

import "reflect"

// Signal represents a reactive value that notifies subscribers on change.
type Signal[T any] struct {
	value         T
	subscribers   []func()
	subPtrs       []uintptr // parallel slice of subscriber pointer identities for dedup
	debounceN     int       // notify every N sets (0 or 1 = always)
	debounceCount int       // rolling counter
}

// NewSignal creates a new signal with an initial value.
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{
		value: initial,
	}
}

// Get returns the current value.
//
// If called inside a running Effect, the Signal automatically registers the
// Effect's run function as a subscriber, enabling automatic re-execution when
// the value changes.
func (s *Signal[T]) Get() T {
	if currentEffect != nil {
		s.Subscribe(currentEffect.run)
	}
	return s.value
}

// Set updates the value and synchronously notifies all subscribers when the
// value actually changes. Assigning the current value is a no-op (prevents
// ping-pong loops when two signals mirror each other).
// If debouncing is configured, subscribers fire only every N calls.
func (s *Signal[T]) Set(value T) {
	if reflect.DeepEqual(s.value, value) {
		return
	}
	s.value = value
	if s.debounceN > 1 {
		s.debounceCount++
		if s.debounceCount < s.debounceN {
			return
		}
		s.debounceCount = 0
	}
	for _, sub := range s.subscribers {
		sub()
	}
}

// SetDebounced sets the debounce factor: subscribers are notified only once
// every n calls to Set. Set n=0 or n=1 to disable debouncing (default).
// Useful for Slider drag signals where the visual update should be immediate
// but downstream work (layout, network) can be throttled.
func (s *Signal[T]) SetDebounce(n int) {
	if n < 1 {
		n = 1
	}
	s.debounceN = n
	s.debounceCount = 0
}

// Subscribe adds a callback that runs when the signal changes.
// Duplicate registrations (same function pointer) are silently ignored.
func (s *Signal[T]) Subscribe(callback func()) {
	ptr := reflect.ValueOf(callback).Pointer()
	for _, p := range s.subPtrs {
		if p == ptr {
			return // already subscribed
		}
	}
	s.subscribers = append(s.subscribers, callback)
	s.subPtrs = append(s.subPtrs, ptr)
}

// SubscriberCount returns the number of unique subscribers. Useful for debugging.
func (s *Signal[T]) SubscriberCount() int { return len(s.subscribers) }

// Effect represents a reactive computation that runs when dependencies change.
type Effect struct {
	run func()
}

// NewEffect creates an Effect and runs fn immediately to collect dependencies.
//
// During the initial run, any Signal.Get() call registers this Effect as a
// subscriber. Subsequent Signal changes cause fn to re-run automatically.
// The returned *Effect is currently informational; there is no Cancel method.
func NewEffect(run func()) *Effect {
	e := &Effect{run: run}
	currentEffect = e
	e.run()
	currentEffect = nil
	return e
}

// currentEffect is the Effect currently executing its run function.
// Signal.Get() reads this to register the Effect as a subscriber.
// It is nil between Effect executions.
var currentEffect *Effect

// SetAsync runs computeFn on the global worker pool and, when the result is
// ready, calls s.Set on the main goroutine via doc.QueueMain. This is safe to
// call from any goroutine or from within an Effect / subscriber callback.
//
// Example — load a value from disk without blocking the frame:
//
//	mySignal.SetAsync(doc, func() string {
//	    data, _ := os.ReadFile("data.txt")
//	    return string(data)
//	})
func (s *Signal[T]) SetAsync(doc *Document, computeFn func() T) {
	SubmitAsyncBg(func() {
		result := computeFn()
		doc.QueueMain(func() {
			s.Set(result)
		})
	})
}
