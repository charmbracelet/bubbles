package cursor

import (
	"sync"
	"testing"
	"time"
)

func TestBlinkMessageIsolation(t *testing.T) {
	first, firstMsg := focusedCursor(t)
	second, secondMsg := focusedCursor(t)
	if firstMsg.tag != secondMsg.tag {
		t.Fatal("blink tags must match so a foreign message is rejected by its ID")
	}

	for _, tc := range []struct {
		name    string
		model   Model
		own     BlinkMsg
		foreign BlinkMsg
	}{
		{"first cursor", first, firstMsg, secondMsg},
		{"second cursor", second, secondMsg, firstMsg},
	} {
		t.Run(tc.name, func(t *testing.T) {
			updated, cmd := tc.model.Update(tc.foreign)
			if updated.IsBlinked != tc.model.IsBlinked || updated.blinkTag != tc.model.blinkTag || cmd != nil {
				t.Fatal("a foreign blink message changed the cursor or scheduled another blink")
			}

			updated, cmd = updated.Update(tc.own)
			if updated.IsBlinked == tc.model.IsBlinked || cmd == nil {
				t.Fatal("the cursor's own blink message did not toggle and schedule another blink")
			}

			// The next command must also produce a message for this cursor.
			next, nextCmd := updated.Update(cmd())
			if next.IsBlinked != tc.model.IsBlinked || nextCmd == nil {
				t.Fatal("the next blink message did not toggle the cursor back")
			}
		})
	}
}

func TestBlinkMessageStaleTag(t *testing.T) {
	m, staleMsg := focusedCursor(t)
	currentMsg := m.Blink()()

	updated, cmd := m.Update(staleMsg)
	if updated.IsBlinked != m.IsBlinked || updated.blinkTag != m.blinkTag || cmd != nil {
		t.Fatal("a stale blink message changed the cursor or scheduled another blink")
	}

	updated, cmd = updated.Update(currentMsg)
	if updated.IsBlinked == m.IsBlinked || cmd == nil {
		t.Fatal("the current blink message did not toggle and schedule another blink")
	}

	replayed, cmd := updated.Update(currentMsg)
	if replayed.IsBlinked != updated.IsBlinked || replayed.blinkTag != updated.blinkTag || cmd != nil {
		t.Fatal("a replayed blink message changed the cursor or scheduled another blink")
	}
}

func focusedCursor(t *testing.T) (Model, BlinkMsg) {
	t.Helper()
	m := New()
	// An already-expired deadline makes the real command deterministic without sleeping.
	m.BlinkSpeed = 0
	cmd := m.Focus()
	if cmd == nil {
		t.Fatal("focusing a blinking cursor did not return a command")
	}
	msg, ok := cmd().(BlinkMsg)
	if !ok {
		t.Fatal("the blink command did not return a BlinkMsg")
	}
	return m, msg
}

// TestBlinkCmdDataRace tests for a race on [Cursor.blinkTag].
//
// The original [Model.Blink] implementation returned a closure over the pointer receiver:
//
//	return func() tea.Msg {
//		defer cancel()
//		<-ctx.Done()
//		if ctx.Err() == context.DeadlineExceeded {
//			return BlinkMsg{id: m.id, tag: m.blinkTag}
//		}
//		return blinkCanceled{}
//	}
//
// A race on “m.blinkTag” will occur if:
//  1. [Model.Blink] is called e.g. by calling [Model.Focus] from
//     ["charm.land/bubbletea/v2".Model.Update];
//  2. ["charm.land/bubbletea/v2".handleCommands] is kept sufficiently busy that it does not receive and
//     execute the [Model.BlinkCmd] e.g. by other long running command or commands;
//  3. at least [Mode.BlinkSpeed] time elapses;
//  4. [Model.Blink] is called again;
//  5. ["charm.land/bubbletea/v2".handleCommands] gets around to receiving and executing the original
//     closure.
//
// Even if this did not formally race, the value of the tag fetched would be semantically incorrect (likely being the
// current value rather than the value at the time the closure was created).
func TestBlinkCmdDataRace(t *testing.T) {
	m := New()
	cmd := m.Blink()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		time.Sleep(m.BlinkSpeed * 3)
		cmd()
	}()
	go func() {
		defer wg.Done()
		time.Sleep(m.BlinkSpeed * 2)
		m.Blink()
	}()
	wg.Wait()
}
