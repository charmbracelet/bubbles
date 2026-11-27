package timer_test

import (
	"fmt"
	"testing"
	"time"

	"charm.land/bubbles/v2/timer"
)

func TestViewDefault(t *testing.T) {
	for _, tt := range []struct {
		name    string
		timeout time.Duration
		want    string
	}{
		{"zero", 0, "0s"},
		{"fraction", 1500 * time.Millisecond, "1.5s"},
		{"negative", -1500 * time.Millisecond, "-1.5s"},
		{"days", 50*time.Hour + 3*time.Minute + 4*time.Second, "50h3m4s"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, m := range []timer.Model{
				timer.New(tt.timeout),
				timer.New(tt.timeout, timer.WithFormatter(nil)),
				{Timeout: tt.timeout},
			} {
				if got := m.View(); got != tt.want {
					t.Errorf("View() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestViewFormatter(t *testing.T) {
	for _, timeout := range []time.Duration{50 * time.Hour, 1500 * time.Millisecond, 0, -1500 * time.Millisecond} {
		t.Run(timeout.String(), func(t *testing.T) {
			calls := 0
			m := timer.New(timeout, timer.WithFormatter(func(remaining time.Duration) string {
				calls++
				if remaining != timeout {
					t.Errorf("formatter received %v, want %v", remaining, timeout)
				}
				return "custom"
			}))
			if got := m.View(); got != "custom" {
				t.Errorf("View() = %q, want custom", got)
			}
			if calls != 1 || m.Timeout != timeout {
				t.Errorf("View() called formatter %d times and left Timeout = %v", calls, m.Timeout)
			}
		})
	}
}

func TestViewFormatterAfterTick(t *testing.T) {
	m := timer.New(2*time.Second, timer.WithInterval(500*time.Millisecond), timer.WithFormatter(func(remaining time.Duration) string {
		return fmt.Sprintf("%dms", remaining.Milliseconds())
	}))
	m, _ = m.Update(timer.TickMsg{ID: m.ID()})
	if got := m.View(); got != "1500ms" {
		t.Errorf("View() after tick = %q, want 1500ms", got)
	}
}

func TestViewFormatterEmpty(t *testing.T) {
	m := timer.New(time.Second, timer.WithFormatter(func(time.Duration) string { return "" }))
	if got := m.View(); got != "" {
		t.Errorf("View() = %q, want empty string", got)
	}
}

func TestWithFormatterNilRestoresDefault(t *testing.T) {
	m := timer.New(time.Second,
		timer.WithFormatter(func(time.Duration) string { return "custom" }),
		timer.WithFormatter(nil),
	)
	if got := m.View(); got != "1s" {
		t.Errorf("View() = %q, want 1s", got)
	}
}
