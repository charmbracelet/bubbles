package stopwatch

import (
	"testing"
	"time"
)

func TestNewDefaultInterval(t *testing.T) {
	t.Parallel()
	m := New()
	if m.Interval != time.Second {
		t.Fatalf("Interval = %v, want %v", m.Interval, time.Second)
	}
}

func TestWithInterval(t *testing.T) {
	t.Parallel()
	m := New(WithInterval(5 * time.Second))
	if m.Interval != 5*time.Second {
		t.Fatalf("Interval = %v, want %v", m.Interval, 5*time.Second)
	}
}

func TestElapsedUsesWallClock(t *testing.T) {
	m := New(WithInterval(time.Hour))
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})

	time.Sleep(50 * time.Millisecond)

	elapsed := m.Elapsed()
	if elapsed < 40*time.Millisecond {
		t.Fatalf("Elapsed() = %v, want at least 40ms", elapsed)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Elapsed() = %v, unexpectedly large", elapsed)
	}
}

func TestStopAccumulatesElapsed(t *testing.T) {
	m := New(WithInterval(time.Hour))
	m, _ = m.Update(StartStopMsg{ID: m.id, running: true})
	m.start = time.Now().Add(-100 * time.Millisecond)

	m, _ = m.Update(StartStopMsg{ID: m.id, running: false})

	if m.Elapsed() < 90*time.Millisecond {
		t.Fatalf("Elapsed() after stop = %v, want at least 90ms", m.Elapsed())
	}
	if m.running {
		t.Fatal("expected stopwatch to be stopped")
	}
}
