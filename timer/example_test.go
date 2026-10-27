package timer_test

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/timer"
)

func ExampleWithFormatter() {
	m := timer.New(50*time.Hour+3*time.Minute+4*time.Second,
		timer.WithFormatter(func(remaining time.Duration) string {
			const day = 24 * time.Hour
			return fmt.Sprintf("%dd %s", remaining/day, remaining%day)
		}),
	)
	fmt.Println(m.View())
	// Output: 2d 2h3m4s
}
