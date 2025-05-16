package cursor

import (
	"sync"
	"testing"
	"time"
)

// TestBlinkCmdDataRace tests for a race on [Cursor.blinkTag].
//
// The original [Model.BlinkCmd] implementation returned a closure over the pointer receiver:
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
//  1. [Model.BlinkCmd] is called e.g. by calling [Model.Focus] from
//     ["github.com/charmbracelet/bubbletea".Model.Update];
//  2. ["github.com/charmbracelet/bubbletea".handleCommands] is kept sufficiently busy that it does not recieve and
//     execute the [Model.BlinkCmd] e.g. by other long running command or commands;
//  3. at least [Mode.BlinkSpeed] time elapses;
//  4. [Model.BlinkCmd] is called again;
//  5. ["github.com/charmbracelet/bubbletea".handleCommands] gets around to receiving and executing the original
//     closure.
//
// Even if this did not formally race, the value of the tag fetched would be semantically incorrect (likely being the
// current value rather than the value at the time the closure was created).
func TestBlinkCmdDataRace(t *testing.T) {
	m := New()
	cmd := m.BlinkCmd()
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
		m.BlinkCmd()
	}()
	wg.Wait()
}
