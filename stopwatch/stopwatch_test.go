package stopwatch

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Run("default interval is 1s", func(t *testing.T) {
		m := New()
		if m.Interval != time.Second {
			t.Errorf("expected default interval %v, got %v", time.Second, m.Interval)
		}
	})

	t.Run("custom interval via WithInterval", func(t *testing.T) {
		custom := 500 * time.Millisecond
		m := New(WithInterval(custom))
		if m.Interval != custom {
			t.Errorf("expected custom interval %v, got %v", custom, m.Interval)
		}
	})
}
