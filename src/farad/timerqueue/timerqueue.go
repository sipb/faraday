package timerqueue

import "time"

type timerElem struct {
	expires time.Time
	entry   string
}

type TimerQueue struct {
	delay time.Duration
	// these should monotonically increase; time glitches are ignored because time.Time objects have a monotonic component
	queue  []timerElem
	endmap map[string]time.Time // allows us to figure out if this is the most recent instance of an element
}

func NewTimerQueue(delay time.Duration) *TimerQueue {
	if delay <= 0 {
		panic("timerqueues must have a positive delay")
	}
	return &TimerQueue{delay, []timerElem{}, map[string]time.Time{}}
}

func (t *TimerQueue) Add(entry string) {
	if t.delay <= 0 {
		panic("timerqueues must have been created by NewTimerQueue!")
	}
	expire_at := time.Now().Add(t.delay)
	t.queue = append(t.queue, timerElem{
		expires: expire_at,
		entry:   entry,
	})
	t.endmap[entry] = expire_at
}

func (t *TimerQueue) Query() (bool, string) {
	if t.delay <= 0 {
		panic("timerqueues must have been created by NewTimerQueue!")
	}
	for len(t.queue) > 0 && time.Now().After(t.queue[0].expires) {
		entry := t.queue[0]
		t.queue = t.queue[1:]
		// this time comparison is okay because it's the *EXACT SAME* time object
		if entry.expires == t.endmap[entry.entry] {
			delete(t.endmap, entry.entry)
			return true, entry.entry
		}
		// otherwise try again
	}
	return false, ""
}
