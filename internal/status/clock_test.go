package status

import (
	"testing"
	"time"
)

func TestFakeClock(t *testing.T) {
	time := time.Now()
	clock := NewFakeClock(time)

	result := clock.Now()
	if result != time {
		t.Errorf("Now() returned %v but expected %v", result, time)
	}
}
