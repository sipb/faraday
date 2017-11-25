package timeutil

import (
	"sync"
	"time"
)

// Calls cb() every (period) amount of time, until halt() (the returned function) is called.
func Tick(cb func(), period time.Duration) func() {
	should_halt := false
	halt_mutex := sync.Mutex{}

	halt := func() {
		halt_mutex.Lock()
		defer halt_mutex.Unlock()
		should_halt = true
	}

	check_halt := func() bool {
		halt_mutex.Lock()
		defer halt_mutex.Unlock()
		return should_halt
	}

	go func() {
		for !check_halt() {
			time.Sleep(period)
			cb()
		}
	}()

	return halt
}
