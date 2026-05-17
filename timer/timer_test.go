package timer

import (
	"testing"
	"time"
)

func TestDuplicateStartDoesNotScheduleTick(t *testing.T) {
	m := New(10*time.Second, WithInterval(time.Second))
	if !m.running {
		t.Fatal("new timer should be running")
	}

	_, cmd := m.Update(StartStopMsg{ID: m.id, running: true})
	if cmd != nil {
		t.Fatal("Start while already running should not schedule another tick")
	}
}

func TestDuplicateStopDoesNotScheduleTick(t *testing.T) {
	m := New(10*time.Second, WithInterval(time.Second))
	m.running = false

	_, cmd := m.Update(StartStopMsg{ID: m.id, running: false})
	if cmd != nil {
		t.Fatal("Stop while already stopped should not schedule a tick")
	}
}

func TestStartAfterStopSchedulesTick(t *testing.T) {
	m := New(10*time.Second, WithInterval(time.Second))
	m.running = false

	updated, cmd := m.Update(StartStopMsg{ID: m.id, running: true})
	if !updated.running {
		t.Fatal("timer should be running after Start")
	}
	if cmd == nil {
		t.Fatal("Start after stop should schedule a tick")
	}
}

func TestTickDoesNotDecrementWhenStopped(t *testing.T) {
	m := New(5*time.Second, WithInterval(time.Second))
	m.running = false

	updated, cmd := m.Update(TickMsg{ID: m.id, tag: m.tag})
	if updated.Timeout != 5*time.Second {
		t.Fatalf("stopped timer timeout = %v, want 5s", updated.Timeout)
	}
	if cmd != nil {
		t.Fatal("stopped timer should not schedule commands on tick")
	}
}
