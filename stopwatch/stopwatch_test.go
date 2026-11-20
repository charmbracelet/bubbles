package stopwatch

import (
	"testing"
	"time"
)

func TestNewDefaultInterval(t *testing.T) {
	m := New()
	if m.Interval != time.Second {
		t.Fatalf("Interval = %v, want 1s", m.Interval)
	}
}

func TestElapsedAdvancesWithTimeWhileRunning(t *testing.T) {
	m := New(WithInterval(time.Second))
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})

	time.Sleep(50 * time.Millisecond)

	// A delayed tick used to add a full Interval (1s) instead of wall-clock time.
	m, _ = m.Update(TickMsg{ID: m.id, tag: m.tag})

	elapsed := m.Elapsed()
	if elapsed < 40*time.Millisecond {
		t.Fatalf("elapsed did not advance with wall clock: %v", elapsed)
	}
	if elapsed >= time.Second {
		t.Fatalf("elapsed jumped by Interval instead of wall clock: %v", elapsed)
	}
}

func TestPauseFreezesElapsed(t *testing.T) {
	m := New(WithInterval(time.Second))
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	time.Sleep(40 * time.Millisecond)
	m, _ = m.Update(StartStopMsg{ID: m.id, running: false})

	paused := m.Elapsed()
	if paused < 20*time.Millisecond {
		t.Fatalf("expected elapsed to have advanced before pause, got %v", paused)
	}

	time.Sleep(40 * time.Millisecond)
	m, cmd := m.Update(TickMsg{ID: m.id, tag: m.tag})
	if cmd != nil {
		t.Fatal("paused stopwatch should not schedule ticks")
	}
	if got := m.Elapsed(); got != paused {
		t.Fatalf("elapsed changed while paused: got %v, want %v", got, paused)
	}

	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	time.Sleep(40 * time.Millisecond)
	resumed := m.Elapsed()
	if resumed < paused+20*time.Millisecond {
		t.Fatalf("elapsed did not continue after resume: paused %v, resumed %v", paused, resumed)
	}
}

func TestResetClearsElapsed(t *testing.T) {
	m := New(WithInterval(time.Second))
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	time.Sleep(40 * time.Millisecond)
	m, _ = m.Update(TickMsg{ID: m.id, tag: m.tag})
	if m.Elapsed() == 0 {
		t.Fatal("expected non-zero elapsed before reset")
	}

	m, _ = m.Update(ResetMsg{ID: m.id})
	if got := m.Elapsed(); got > 15*time.Millisecond {
		t.Fatalf("reset did not clear elapsed: %v", got)
	}

	m, _ = m.Update(StartStopMsg{ID: m.id, running: false})
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	time.Sleep(40 * time.Millisecond)
	m, _ = m.Update(StartStopMsg{ID: m.id, running: false})
	if m.Elapsed() == 0 {
		t.Fatal("expected non-zero elapsed before stopped reset")
	}
	m, _ = m.Update(ResetMsg{ID: m.id})
	if got := m.Elapsed(); got != 0 {
		t.Fatalf("reset while stopped did not clear elapsed: %v", got)
	}
}

func TestTagDebounceRejectsForeignTags(t *testing.T) {
	m := New(WithInterval(time.Second))
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	time.Sleep(20 * time.Millisecond)

	// Advance past the initial tag 0, which is always accepted.
	m, cmd := m.Update(TickMsg{ID: m.id, tag: m.tag})
	if cmd == nil {
		t.Fatal("initial tick should schedule a new tick")
	}
	tag := m.tag
	if tag == 0 {
		t.Fatal("expected tag to advance after the initial tick")
	}

	m, cmd = m.Update(TickMsg{ID: m.id, tag: tag + 7})
	if cmd != nil {
		t.Fatal("foreign tag should not schedule a new tick")
	}
	if m.tag != tag {
		t.Fatalf("tag changed on foreign tick: got %d, want %d", m.tag, tag)
	}
	if m.Elapsed() >= 2*time.Second {
		t.Fatalf("foreign ticks accelerated elapsed: %v", m.Elapsed())
	}

	m, cmd = m.Update(TickMsg{ID: m.id, tag: tag})
	if cmd == nil {
		t.Fatal("matching tag should schedule a new tick")
	}
	if m.tag != tag+1 {
		t.Fatalf("tag = %d, want %d", m.tag, tag+1)
	}

	m, cmd = m.Update(TickMsg{ID: m.id, tag: tag})
	if cmd != nil {
		t.Fatal("stale tag should not schedule a new tick")
	}
	if m.tag != tag+1 {
		t.Fatalf("tag changed on stale tick: got %d", m.tag)
	}
}

func TestViewAlignsToInterval(t *testing.T) {
	m := New()
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	time.Sleep(50 * time.Millisecond)

	if got, want := m.View(), "0s"; got != want {
		t.Fatalf("View() = %q, want %q", got, want)
	}
	if m.Elapsed() < 20*time.Millisecond {
		t.Fatalf("Elapsed() should return the raw duration, got %v", m.Elapsed())
	}
}
