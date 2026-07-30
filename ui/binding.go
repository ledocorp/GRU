// Package ui (continued)
// See node.go for the full package documentation.
package ui

// Binding[T] is a typed wrapper around a Signal[T] that provides a clean
// read/write interface for connecting UI widgets to application data.
//
// Use Binding for scalar values (a single string, number, or bool).
// Use ListBinding for ordered collections (VirtualList data).
type Binding[T any] struct {
	signal *Signal[T]
}

// NewBinding creates a new binding with an initial value.
func NewBinding[T any](initialValue T) *Binding[T] {
	return &Binding[T]{
		signal: NewSignal(initialValue),
	}
}

// Get returns the current value.
func (b *Binding[T]) Get() T {
	return b.signal.Get()
}

// Set updates the value and notifies subscribers.
func (b *Binding[T]) Set(value T) {
	b.signal.Set(value)
}

// Subscribe adds a callback that runs when the value changes.
func (b *Binding[T]) Subscribe(callback func()) {
	b.signal.Subscribe(callback)
}

// ListBinding[T] is a reactive data source for ordered lists.
//
// It holds two Signals: one for the item slice and one for the selected index.
// VirtualList subscribes to both via SubscribeItems and SubscribeSelection so
// it calls MarkDirty automatically whenever the data or selection changes.
//
// The selected index is -1 when nothing is selected.
type ListBinding[T any] struct {
	items    *Signal[[]T]
	selected *Signal[int] // -1 for no selection
}

// NewListBinding creates a new list binding with initial items.
func NewListBinding[T any](initialItems []T) *ListBinding[T] {
	return &ListBinding[T]{
		items:    NewSignal(initialItems),
		selected: NewSignal(-1),
	}
}

// GetItems returns the current list of items.
func (lb *ListBinding[T]) GetItems() []T {
	return lb.items.Get()
}

// SetItems updates the entire list.
func (lb *ListBinding[T]) SetItems(items []T) {
	lb.items.Set(items)
}

// GetSelectedIndex returns the currently selected item index.
func (lb *ListBinding[T]) GetSelectedIndex() int {
	return lb.selected.Get()
}

// SetSelectedIndex updates the selected item index.
func (lb *ListBinding[T]) SetSelectedIndex(index int) {
	lb.selected.Set(index)
}

// GetSelectedItem returns the currently selected item, or zero value if none.
func (lb *ListBinding[T]) GetSelectedItem() T {
	index := lb.selected.Get()
	items := lb.items.Get()
	if index >= 0 && index < len(items) {
		return items[index]
	}
	var zero T
	return zero
}

// SubscribeItems adds a callback that runs when items change.
func (lb *ListBinding[T]) SubscribeItems(callback func()) {
	lb.items.Subscribe(callback)
}

// SubscribeSelection adds a callback that runs when selection changes.
func (lb *ListBinding[T]) SubscribeSelection(callback func()) {
	lb.selected.Subscribe(callback)
}

// Len returns the number of items.
func (lb *ListBinding[T]) Len() int {
	return len(lb.items.Get())
}

// GetItem returns the item at the specified index.
func (lb *ListBinding[T]) GetItem(index int) T {
	items := lb.items.Get()
	if index >= 0 && index < len(items) {
		return items[index]
	}
	var zero T
	return zero
}
