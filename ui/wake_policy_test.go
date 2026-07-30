package ui

import (
	"testing"
	"time"
)

func TestRenderIdlePolicyScrollTierWhenClean(t *testing.T) {
	now := time.Now()
	p := NewRenderIdlePolicy(now)
	fps := p.Update(now, WakeSummary{Reasons: WakeScroll}, true, 0, false)
	if fps != ScrollFPS {
		t.Fatalf("clean scroll wake: got target %d, want ScrollFPS=%d", fps, ScrollFPS)
	}
	if p.State() != "scroll" {
		t.Fatalf("state=%q, want scroll", p.State())
	}
}

func TestRenderIdlePolicyScrollRedrawUsesActiveFPS(t *testing.T) {
	now := time.Now()
	p := NewRenderIdlePolicy(now)
	fps := p.Update(now, WakeSummary{Reasons: WakeScroll}, false, 0, false)
	if fps != ActiveFPS {
		t.Fatalf("scroll + redraw: got target %d, want ActiveFPS=%d", fps, ActiveFPS)
	}
}

func TestRenderIdlePolicyGraceAfterNoteInteractiveWake(t *testing.T) {
	start := time.Now()
	p := NewRenderIdlePolicy(start.Add(-time.Minute))
	p.Update(start.Add(-time.Minute), WakeSummary{}, true, 0, false)

	click := start
	p.Update(click, WakeSummary{Reasons: WakeInput}, true, 0, false)
	p.NoteInteractiveWake(click)

	next := click.Add(20 * time.Millisecond)
	fps := p.Update(next, WakeSummary{}, true, 0, false)
	if fps != ActiveFPS {
		t.Fatalf("20ms after click: target=%d state=%q, want ActiveFPS grace", fps, p.State())
	}
	if p.State() != "grace" {
		t.Fatalf("state=%q, want grace", p.State())
	}
}

func TestRenderIdlePolicyIdleGraceUsesInteractiveWake(t *testing.T) {
	start := time.Now()
	p := NewRenderIdlePolicy(start)
	p.Update(start, WakeSummary{Reasons: WakeInput}, true, 0, false)
	p.NoteInteractiveWake(start)

	// Without interactive wake timestamp, a stale lastWake would drop immediately.
	next := start.Add(400 * time.Millisecond)
	fps := p.Update(next, WakeSummary{}, true, 0, false)
	if fps != ActiveFPS {
		t.Fatalf("400ms after click: target=%d state=%q, want ActiveFPS grace", fps, p.State())
	}
}

func TestRenderIdlePolicySceneLoadGrace(t *testing.T) {
	start := time.Now()
	p := NewRenderIdlePolicy(start.Add(-time.Minute))
	p.Update(start.Add(-time.Minute), WakeSummary{}, true, 0, false)
	p.NoteSceneLoad(start)

	mid := start.Add(500 * time.Millisecond)
	fps := p.Update(mid, WakeSummary{}, true, 0, false)
	if fps != ActiveFPS {
		t.Fatalf("500ms after scene load: target=%d state=%q, want ActiveFPS", fps, p.State())
	}
	if p.State() != "scene-grace" {
		t.Fatalf("state=%q, want scene-grace", p.State())
	}

	after := start.Add(sceneLoadGraceDuration + 50*time.Millisecond)
	fps = p.Update(after, WakeSummary{}, true, 0, false)
	if fps != DeepIdleFPS {
		t.Fatalf("after scene grace: target=%d state=%q, want DeepIdleFPS", fps, p.State())
	}
}

func TestRenderIdlePolicyAnimationOnlyUsesAnimationFPS(t *testing.T) {
	start := time.Now()
	p := NewRenderIdlePolicy(start.Add(-time.Minute))
	// Establish deep-idle baseline (no interactive wake recently).
	p.Update(start, WakeSummary{}, true, 0, false)

	later := start.Add(activityGraceDuration + time.Second)
	fps := p.Update(later, WakeSummary{Reasons: WakeAnimation}, true, 0, false)
	if fps != AnimationFPS {
		t.Fatalf("animation-only clean: target=%d state=%q, want AnimationFPS=%d", fps, p.State(), AnimationFPS)
	}
	if p.State() != "animation" {
		t.Fatalf("state=%q, want animation", p.State())
	}
}
