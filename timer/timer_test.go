package timer

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// TestDuplicateTickStreamDoesNotPermanentlyDoubleSpeed reproduces
// https://github.com/charmbracelet/bubbles/issues/1007: the tag-based
// debounce guard in Update (mirrored from stopwatch) is dead code because
// nothing ever increments m.tag. As a result, delivering a second
// StartStopMsg to an already-running timer (e.g. a duplicate Start/Toggle,
// or Init being wired up twice) spawns a second, independent tick stream
// that is never rejected, and the timer decrements twice per interval,
// forever, with the overshoot growing without bound as more ticks arrive.
//
// With a working debounce, the two streams race exactly once (this first
// hit is unavoidable: both streams are genuinely untagged, so there's no
// way to tell them apart yet), but the guard then rejects the stale
// stream's every subsequent tick, so only one stream survives. The
// overshoot this causes is capped at exactly one interval no matter how
// many rounds follow, instead of scaling with the round count.
func TestDuplicateTickStreamDoesNotPermanentlyDoubleSpeed(t *testing.T) {
	m := New(10*time.Second, WithInterval(time.Millisecond))

	// Chain A: the legitimate tick stream started by Init.
	cmdA := m.Init()
	if cmdA == nil {
		t.Fatal("expected Init to return a tick command")
	}

	// Chain B: simulate a duplicate Start/Toggle delivered to the
	// already-running timer. Per the StartStopMsg handler this
	// unconditionally spawns a second, independent tick stream.
	m, cmdB := m.Update(StartStopMsg{ID: m.id, running: true})
	if cmdB == nil {
		t.Fatal("expected the duplicate StartStopMsg to spawn a second tick command")
	}

	initialTimeout := m.Timeout
	interval := m.Interval

	const rounds = 20
	for i := 0; i < rounds; i++ {
		if cmdA != nil {
			m, cmdA = updateWithTick(t, m, cmdA)
		}
		if cmdB != nil {
			m, cmdB = updateWithTick(t, m, cmdB)
		}
	}

	elapsed := initialTimeout - m.Timeout
	// Exactly one interval of unavoidable one-time overshoot from the
	// initial untagged race, and nothing more: with a working debounce
	// the stale stream is rejected on its very next tick and stops
	// spawning further commands, so the overshoot never grows past that
	// single interval, no matter how many rounds run.
	wantElapsed := time.Duration(rounds+1) * interval
	if elapsed != wantElapsed {
		t.Errorf("after %d rounds: elapsed = %v, want %v (overshoot from the duplicate tick stream is growing with the round count instead of settling after one interval; debounce guard is not rejecting the stale stream)", rounds, elapsed, wantElapsed)
	}
}

// updateWithTick executes cmd (expected to eventually deliver a TickMsg,
// possibly wrapped in a batch/sequence by the runtime), feeds the resulting
// TickMsg through Update, and returns the updated model along with the next
// tick command in the chain (or nil if the chain was rejected/terminated).
func updateWithTick(t *testing.T, m Model, cmd tea.Cmd) (Model, tea.Cmd) {
	t.Helper()

	msg := cmd()
	tick, ok := msg.(TickMsg)
	if !ok {
		t.Fatalf("expected TickMsg, got %T", msg)
	}

	newM, next := m.Update(tick)
	return newM, next
}
