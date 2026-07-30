// Package ui (continued)
// See node.go for the full package documentation.
package ui

// EventType identifies the kind of UI event.
type EventType string

// Standard event types fired by widgets.
const (
	EventClick     EventType = "click"     // Mouse button pressed on widget
	EventFocus     EventType = "focus"     // Widget gained keyboard focus
	EventBlur      EventType = "blur"      // Widget lost keyboard focus
	EventKeyPress  EventType = "keypress"  // Key pressed while focused
	EventMouseMove EventType = "mousemove" // Mouse moved over widget
)

// Event carries information about a UI interaction.
// Events bubble upward through the parent chain unless a handler sets Bubble = false.
type Event struct {
	Type   EventType
	Target Node        // The widget that originally emitted the event
	Data   interface{} // Optional payload (key code, position, etc.)
	Bubble bool        // Whether the event should propagate to parent nodes
}

// EventEmitter provides widget-level event subscription and emission.
// Element implements this interface; all widgets inherit it.
type EventEmitter interface {
	// On registers handler to be called whenever events of eventType are emitted.
	On(eventType EventType, handler func(Event))
	// Emit fires an event of the given type with optional data payload.
	Emit(eventType EventType, data interface{})
}
