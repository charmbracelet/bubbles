package timer

import (
	"testing"
	"time"
)

// TestDoubleStartDoesNotSpeedUpTimer verifies that calling Start() twice
// does not cause the timer to tick faster than the interval.
// This is a regression test for issue #867.
func TestDoubleStartDoesNotSpeedUpTimer(t *testing.T) {
	// Create a timer with 5 second timeout and 100ms interval
	m := New(5*time.Second, WithInterval(100*time.Millisecond))

	// Initially the timer is running=true but hasn't started ticking yet
	// First Start() - should work since we need to set running=false first
	m.running = false

	// Start the timer
	cmd1 := m.Start()
	if cmd1 == nil {
		t.Fatal("expected Start() to return a command when stopped")
	}

	// Apply the command
	msg1 := cmd1()
	startStop1, ok := msg1.(StartStopMsg)
	if !ok {
		t.Fatalf("expected StartStopMsg, got %T", msg1)
	}

	// Update model - this should start ticking
	m, cmd := m.Update(startStop1)
	if cmd == nil {
		t.Fatal("expected Update to return tick command when starting")
	}

	// Try starting again while already running
	cmd2 := m.Start()
	if cmd2 == nil {
		t.Fatal("expected Start() to return a command even when running")
	}

	// Apply the second command
	msg2 := cmd2()
	startStop2, ok := msg2.(StartStopMsg)
	if !ok {
		t.Fatalf("expected StartStopMsg, got %T", msg2)
	}

	// Update model - this should NOT spawn another ticker
	_, cmd = m.Update(startStop2)
	if cmd != nil {
		t.Error("expected Update to return nil when already running, preventing duplicate tickers")
	}
}

// TestDoubleStopDoesNotCauseIssues verifies that calling Stop() twice
// does not cause any issues.
func TestDoubleStopDoesNotCauseIssues(t *testing.T) {
	m := New(5*time.Second, WithInterval(100*time.Millisecond))
	m.running = true

	// Stop the timer
	cmd1 := m.Stop()
	if cmd1 == nil {
		t.Fatal("expected Stop() to return a command")
	}

	// Apply the command
	msg1 := cmd1()
	startStop1, ok := msg1.(StartStopMsg)
	if !ok {
		t.Fatalf("expected StartStopMsg, got %T", msg1)
	}

	// Update model - this should stop the timer
	m, _ = m.Update(startStop1)
	if m.Running() {
		t.Error("expected timer to be stopped")
	}

	// Stop again while already stopped
	cmd2 := m.Stop()
	if cmd2 == nil {
		t.Fatal("expected Stop() to return a command even when stopped")
	}

	// Apply the second command
	msg2 := cmd2()
	startStop2, ok := msg2.(StartStopMsg)
	if !ok {
		t.Fatalf("expected StartStopMsg, got %T", msg2)
	}

	// Update model - this should be a no-op
	_, cmd := m.Update(startStop2)
	if cmd != nil {
		t.Error("expected Update to return nil when already stopped")
	}
}

// TestToggleBehavior verifies that Toggle works correctly
func TestToggleBehavior(t *testing.T) {
	m := New(5*time.Second, WithInterval(100*time.Millisecond))

	// Initially running (created with running=true)
	if !m.Running() {
		t.Error("expected timer to be running initially")
	}

	// Toggle should stop it
	cmd := m.Toggle()
	if cmd == nil {
		t.Fatal("expected Toggle() to return a command")
	}

	// Apply the command to get the message
	msg := cmd()
	startStopMsg, ok := msg.(StartStopMsg)
	if !ok {
		t.Fatalf("expected StartStopMsg, got %T", msg)
	}
	if startStopMsg.running {
		t.Error("expected Toggle to stop the timer")
	}

	// Update the model with the message
	m, _ = m.Update(startStopMsg)
	if m.Running() {
		t.Error("expected timer to be stopped after Toggle")
	}

	// Toggle again should start it
	cmd = m.Toggle()
	if cmd == nil {
		t.Fatal("expected Toggle() to return a command when stopped")
	}

	msg = cmd()
	startStopMsg, ok = msg.(StartStopMsg)
	if !ok {
		t.Fatalf("expected StartStopMsg, got %T", msg)
	}
	if !startStopMsg.running {
		t.Error("expected Toggle to start the timer")
	}
}

// TestUpdateHandlesStartStopMsg verifies that Update correctly handles
// StartStopMsg and only ticks when state changes to running.
func TestUpdateHandlesStartStopMsg(t *testing.T) {
	m := New(5*time.Second, WithInterval(100*time.Millisecond))

	// Stop the timer first
	m.running = false

	// Send a StartStopMsg to start it
	msg := StartStopMsg{ID: m.id, running: true}
	newM, cmd := m.Update(msg)

	if !newM.Running() {
		t.Error("expected timer to be running after StartStopMsg")
	}
	if cmd == nil {
		t.Error("expected a command to tick when starting")
	}

	// Send another StartStopMsg with same running state
	m = newM
	msg = StartStopMsg{ID: m.id, running: true}
	newM, cmd = m.Update(msg)

	// Should still be running but no new tick command
	if !newM.Running() {
		t.Error("expected timer to still be running")
	}
	if cmd != nil {
		t.Error("expected no command when state doesn't change")
	}
}

// TestStartReturnsNilWhenTimedOut verifies that Start has no effect
// when the timer has already timed out.
func TestStartReturnsNilWhenTimedOut(t *testing.T) {
	// Create an already timed-out timer
	m := New(0)

	if !m.Timedout() {
		t.Fatal("expected timer to be timed out with 0 duration")
	}

	// Start should have no effect when timed out
	cmd := m.Start()
	if cmd != nil {
		t.Error("expected Start() to return nil when timer has timed out")
	}
}

// TestInitStartsTimer verifies that Init returns a tick command.
func TestInitStartsTimer(t *testing.T) {
	m := New(5*time.Second, WithInterval(100*time.Millisecond))

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init() to return a command")
	}
}

// TestTickMsgHandling verifies that TickMsg is handled correctly.
func TestTickMsgHandling(t *testing.T) {
	m := New(5*time.Second, WithInterval(1*time.Second))

	// Send a tick message
	msg := TickMsg{ID: m.id, Timeout: false}
	newM, cmd := m.Update(msg)

	// Timer should decrease by interval
	expectedTimeout := 4 * time.Second
	if newM.Timeout != expectedTimeout {
		t.Errorf("expected timeout to be %v, got %v", expectedTimeout, newM.Timeout)
	}
	if cmd == nil {
		t.Error("expected a command after tick")
	}
}

// TestTimeoutMsgSentOnLastTick verifies that TimeoutMsg is sent when timer reaches 0.
func TestTimeoutMsgSentOnLastTick(t *testing.T) {
	// Create a timer with 1 second timeout
	m := New(1*time.Second, WithInterval(1*time.Second))

	// Send a tick message that will cause timeout
	msg := TickMsg{ID: m.id, Timeout: true}
	newM, cmd := m.Update(msg)

	if !newM.Timedout() {
		t.Error("expected timer to be timed out")
	}
	if newM.Running() {
		t.Error("expected timer to not be running after timeout")
	}

	// Should have a batch command with timeout message
	if cmd == nil {
		t.Error("expected a command after final tick")
	}
}

// TestTickMsgFromDifferentTimerIsIgnored verifies that tick messages
// from other timers are ignored.
func TestTickMsgFromDifferentTimerIsIgnored(t *testing.T) {
	m := New(5*time.Second, WithInterval(1*time.Second))
	originalTimeout := m.Timeout

	// Send a tick message from a different timer
	msg := TickMsg{ID: 99999, Timeout: false}
	newM, _ := m.Update(msg)

	// Timeout should not change
	if newM.Timeout != originalTimeout {
		t.Error("expected tick from different timer to be ignored")
	}
}

// TestStartStopMsgFromDifferentTimerIsIgnored verifies that StartStopMsg
// from other timers are ignored.
func TestStartStopMsgFromDifferentTimerIsIgnored(t *testing.T) {
	m := New(5*time.Second)
	m.running = false

	// Send a StartStopMsg from a different timer
	msg := StartStopMsg{ID: 99999, running: true}
	newM, cmd := m.Update(msg)

	// Should still be stopped
	if newM.Running() {
		t.Error("expected StartStopMsg from different timer to be ignored")
	}
	if cmd != nil {
		t.Error("expected no command when message is ignored")
	}
}

// TestTagPreventsStaleTicks verifies that the tag mechanism prevents
// processing of stale tick messages.
func TestTagPreventsStaleTicks(t *testing.T) {
	m := New(5*time.Second, WithInterval(1*time.Second))
	m.tag = 5 // Simulate being on the 5th tick

	// Send a tick with an old tag
	msg := TickMsg{ID: m.id, tag: 3, Timeout: false}
	newM, cmd := m.Update(msg)

	// Should ignore the stale tick
	if newM.Timeout != m.Timeout {
		t.Error("expected stale tick to be ignored")
	}
	if cmd != nil {
		t.Error("expected no command for stale tick")
	}
}

// TestIDReturnsCorrectID verifies that ID() returns the timer's ID.
func TestIDReturnsCorrectID(t *testing.T) {
	m1 := New(5*time.Second)
	m2 := New(5*time.Second)

	if m1.ID() == m2.ID() {
		t.Error("expected different timers to have different IDs")
	}
	if m1.ID() != m1.id {
		t.Error("expected ID() to return the internal id")
	}
}

// TestViewReturnsTimeoutString verifies that View returns the timeout as a string.
func TestViewReturnsTimeoutString(t *testing.T) {
	m := New(5*time.Second)
	view := m.View()

	if view != "5s" {
		t.Errorf("expected View to return '5s', got '%s'", view)
	}
}

// TestRunningReturnsFalseWhenTimedOut verifies that Running() returns false
// when the timer has timed out.
func TestRunningReturnsFalseWhenTimedOut(t *testing.T) {
	m := New(0)

	if m.Running() {
		t.Error("expected Running() to return false when timed out")
	}
}
