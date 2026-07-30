// Package ui (continued)
// See node.go for the full package documentation.
package ui

import "math"

// Tween animates a float32 value from StartValue to EndValue over Duration
// seconds using a pluggable easing function.
//
// Each frame, the caller must invoke tween.Update(dt). The OnUpdate callback
// receives the current interpolated value so the caller can apply it to a
// widget property (e.g. button.Scale = v). OnComplete fires once when the
// tween finishes. IsActive becomes false at that point.
//
// Example — scale a button on click:
//
//	tween := ui.NewTween(1.0, 1.2, 0.3, ui.EaseOutQuad,
//	    func(v float32) { btn.Scale = v },
//	    func()           { btn.Scale = 1.0 },
//	)
type Tween struct {
	StartValue float32
	EndValue   float32
	Duration   float32
	Elapsed    float32
	EasingFunc func(t float32) float32 // Easing function, t in [0,1]
	OnUpdate   func(value float32)
	OnComplete func()
	IsActive   bool
}

// NewTween creates a new tween.
func NewTween(start, end, duration float32, easing func(float32) float32, onUpdate func(float32), onComplete func()) *Tween {
	return &Tween{
		StartValue: start,
		EndValue:   end,
		Duration:   duration,
		EasingFunc: easing,
		OnUpdate:   onUpdate,
		OnComplete: onComplete,
		IsActive:   true,
	}
}

// Update advances the tween.
func (t *Tween) Update(dt float32) {
	if !t.IsActive {
		return
	}
	t.Elapsed += dt
	if t.Elapsed >= t.Duration {
		t.Elapsed = t.Duration
		t.IsActive = false
		if t.OnUpdate != nil {
			t.OnUpdate(t.EndValue)
		}
		if t.OnComplete != nil {
			t.OnComplete()
		}
	} else {
		progress := t.Elapsed / t.Duration
		if t.EasingFunc != nil {
			progress = t.EasingFunc(progress)
		}
		value := t.StartValue + (t.EndValue-t.StartValue)*progress
		if t.OnUpdate != nil {
			t.OnUpdate(value)
		}
	}
}

// Linear easing.
func LinearEasing(t float32) float32 {
	return t
}

// EaseOutQuad easing.
func EaseOutQuad(t float32) float32 {
	return 1 - (1-t)*(1-t)
}

// EaseInOutQuad easing.
func EaseInOutQuad(t float32) float32 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - float32(math.Pow(float64(-2*t+2), 2))/2
}
