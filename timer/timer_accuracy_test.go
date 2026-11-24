package timer_test

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
)

func TestTimerTracksElapsedBetweenUpdates(t *testing.T) {
	m := timer.New(10*time.Second, timer.WithInterval(time.Millisecond))
	startBefore := time.Now()
	cmd := m.Init()
	startAfter := time.Now()
	for i := 0; i < 20; i++ {
		msg := cmd()
		// Simulate work in the parent model before it forwards the tick.
		time.Sleep(2 * time.Millisecond)
		before := time.Now()
		m, cmd = m.Update(msg)
		after := time.Now()
		elapsed := 10*time.Second - m.Timeout
		if elapsed < before.Sub(startAfter)-time.Millisecond || elapsed > after.Sub(startBefore)+time.Millisecond {
			t.Fatalf("tick %d: timer elapsed %v, wall clock between %v and %v", i, elapsed, before.Sub(startAfter), after.Sub(startBefore))
		}
	}
}

func TestTimerInitDoesNotCountTimeSinceNew(t *testing.T) {
	m := timer.New(time.Second, timer.WithInterval(time.Millisecond))
	time.Sleep(20 * time.Millisecond)
	before := time.Now()
	cmd := m.Init()
	msg, ok := cmd().(timer.TickMsg)
	if !ok {
		t.Fatal("Init must deliver a TickMsg")
	}
	m, _ = m.Update(msg)
	after := time.Now()
	if elapsed := time.Second - m.Timeout; elapsed > after.Sub(before)+time.Millisecond {
		t.Fatalf("timer counted time before Init: %v", elapsed)
	}
}

func TestTimerResumeRejectsPendingInitialTick(t *testing.T) {
	m := timer.New(time.Second, timer.WithInterval(time.Millisecond))
	oldTick := m.Init()()
	m, _ = m.Update(m.Stop()())
	paused := m.Timeout
	time.Sleep(20 * time.Millisecond)
	startBefore := time.Now()
	m, cmd := m.Update(m.Start()())
	startAfter := time.Now()
	for i := 0; i < 3; i++ {
		var stale tea.Cmd
		m, stale = m.Update(oldTick)
		if stale != nil || m.Timeout != paused {
			t.Fatal("pending tick from before pause was accepted after resume")
		}
	}
	msg := cmd()
	time.Sleep(2 * time.Millisecond)
	before := time.Now()
	m, _ = m.Update(msg)
	after := time.Now()
	if elapsed := paused - m.Timeout; elapsed < before.Sub(startAfter)-time.Millisecond || elapsed > after.Sub(startBefore)+time.Millisecond {
		t.Fatalf("resumed timer elapsed %v, active time between %v and %v", elapsed, before.Sub(startAfter), after.Sub(startBefore))
	}
}

func TestTimerPauseFreezesRemainingTime(t *testing.T) {
	m := timer.New(time.Second, timer.WithInterval(time.Millisecond))
	m, next := m.Update(m.Init()())
	m, _ = m.Update(m.Stop()())
	paused := m.Timeout
	time.Sleep(20 * time.Millisecond)
	m, cmd := m.Update(next())
	if cmd != nil || m.Timeout != paused {
		t.Fatal("paused timer accepted a pending tick")
	}
	m, cmd = m.Update(m.Stop()())
	if cmd != nil || m.Timeout != paused {
		t.Fatal("repeated Stop changed the paused timer")
	}
}

func TestTimerManualTicksStillUseInterval(t *testing.T) {
	m := timer.New(time.Second, timer.WithInterval(10*time.Millisecond))
	m, _ = m.Update(m.Init()())
	for _, id := range []int{0, m.ID()} {
		before := m.Timeout
		m, _ = m.Update(timer.TickMsg{ID: id})
		if m.Timeout != before-m.Interval {
			t.Fatalf("manual tick: timeout %v, want %v", m.Timeout, before-m.Interval)
		}
	}
}

func TestTimerUsesResetTimeout(t *testing.T) {
	m := timer.New(time.Second, timer.WithInterval(time.Millisecond))
	startBefore := time.Now()
	m, next := m.Update(m.Init()())
	startAfter := time.Now()
	m.Timeout = 2 * time.Second
	msg := next()
	time.Sleep(2 * time.Millisecond)
	before := time.Now()
	m, _ = m.Update(msg)
	after := time.Now()
	if elapsed := 2*time.Second - m.Timeout; elapsed < before.Sub(startAfter)-time.Millisecond || elapsed > after.Sub(startBefore)+time.Millisecond {
		t.Fatalf("reset timer elapsed %v, tick time between %v and %v", elapsed, before.Sub(startAfter), after.Sub(startBefore))
	}
}

func TestTimerDelayedTickCanExpire(t *testing.T) {
	m := timer.New(5*time.Millisecond, timer.WithInterval(time.Millisecond))
	msg := m.Init()()
	time.Sleep(10 * time.Millisecond)
	m, cmd := m.Update(msg)
	if !m.Timedout() || m.Running() || m.Timeout >= 0 {
		t.Fatalf("delayed tick did not expire timer: timeout %v", m.Timeout)
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatal("expected final tick and timeout commands")
	}
	var ticks, timeouts int
	for _, command := range batch {
		switch msg := command().(type) {
		case timer.TickMsg:
			ticks++
			if !msg.Timeout || msg.ID != m.ID() {
				t.Fatal("final tick did not report timeout")
			}
		case timer.TimeoutMsg:
			timeouts++
			if msg.ID != m.ID() {
				t.Fatal("timeout belongs to a different timer")
			}
		}
	}
	if ticks != 1 || timeouts != 1 {
		t.Fatalf("got %d final ticks and %d timeouts, want one of each", ticks, timeouts)
	}
}

func TestTimerNonPositiveTimeoutRemainsExpired(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		m := timer.New(timeout, timer.WithInterval(time.Millisecond))
		m, cmd := m.Update(m.Init()())
		if cmd != nil || m.Timeout != timeout || m.Running() || m.Start() != nil {
			t.Fatalf("expired timer changed: timeout %v, want %v", m.Timeout, timeout)
		}
	}
}
