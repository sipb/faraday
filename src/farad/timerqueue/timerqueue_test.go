package timerqueue

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTimerQueue(t *testing.T) {
	tq := NewTimerQueue(time.Second * 2)
	if tq.delay != time.Second*2 {
		t.Error("")
	}
}

func TestTimerQueue_Add(t *testing.T) {
	tq := NewTimerQueue(time.Second)
	for i := 0; i < 10; i++ {
		before := time.Now()
		tq.Add(fmt.Sprintf("entry-%d", i))
		after := time.Now()
		original := tq.queue[len(tq.queue)-1].expires.Add(-time.Second)
		if original.Before(before) || original.After(after) {
			t.Errorf("time out of range")
		}
		if len(tq.queue) != i+1 {
			t.Errorf("wrong length of queue")
		}
		for qi, qv := range tq.queue {
			if qv.entry != fmt.Sprintf("entry-%d", qi) {
				t.Errorf("wrong element")
			}
		}
	}
}

func TestTimerQueue_Query_Simple(t *testing.T) {
	tq := NewTimerQueue(time.Nanosecond)
	found, _ := tq.Query()
	if found {
		t.Error("should not have been found")
	}
	tq.Add("entry1")
	time.Sleep(time.Nanosecond)
	tq.Add("entry2")
	time.Sleep(time.Nanosecond)
	tq.Add("entry3")
	time.Sleep(time.Nanosecond)
	tq.Add("entry1")
	time.Sleep(time.Nanosecond)
	tq.Add("entry4")
	time.Sleep(time.Nanosecond)

	found, val := tq.Query()
	if !found || val != "entry2" {
		t.Error("wrong first query")
	}
	found, val = tq.Query()
	if !found || val != "entry3" {
		t.Error("wrong first query")
	}
	found, val = tq.Query()
	if !found || val != "entry1" {
		t.Error("wrong first query")
	}
	found, val = tq.Query()
	if !found || val != "entry4" {
		t.Error("wrong first query")
	}
	found, _ = tq.Query()
	if found {
		t.Error("should not be found")
	}
	found, _ = tq.Query()
	if found {
		t.Error("should not be found")
	}
}

func TestTimerQueue_Query_Pause(t *testing.T) {
	tq := NewTimerQueue(time.Millisecond * 20)
	found, _ := tq.Query()
	if found {
		t.Error("should not have been found")
	}
	tq.Add("entry1")
	time.Sleep(time.Nanosecond)
	tq.Add("entry2")
	time.Sleep(time.Nanosecond)
	tq.Add("entry1")
	time.Sleep(time.Nanosecond)
	found, _ = tq.Query()
	if found {
		t.Error("should not be found")
	}
	time.Sleep(time.Millisecond * 10)
	found, _ = tq.Query()
	if found {
		t.Error("should not be found")
	}
	time.Sleep(time.Millisecond * 10)
	found, val := tq.Query()
	if !found {
		t.Error("should be found")
	}
	if val != "entry2" {
		t.Error("wrong first value")
	}
	found, val = tq.Query()
	if !found {
		t.Error("should be found")
	}
	if val != "entry1" {
		t.Error("wrong first value")
	}
	found, _ = tq.Query()
	if found {
		t.Error("should not be found")
	}
}

func TestTimerQueue_Query_Cyclic(t *testing.T) {
	tq := NewTimerQueue(time.Nanosecond)
	tq.Add("test-0")
	for i := 1; i <= 100; i++ {
		last := fmt.Sprintf("test-%d", i-1)
		next := fmt.Sprintf("test-%d", i)
		tq.Add(next)
		time.Sleep(time.Nanosecond)
		found, val := tq.Query()
		if !found {
			t.Error("expected to be found")
		}
		if val != last {
			t.Error("wrong value")
		}
	}
	found, val := tq.Query()
	if !found {
		t.Error("expected to be found")
	}
	if val != "test-100" {
		t.Error("wrong value")
	}
	found, val = tq.Query()
	if found {
		t.Error("expected not found")
	}
	if len(tq.queue) != 0 {
		t.Error("queue should be empty")
	}
	if len(tq.endmap) != 0 {
		t.Error("endmap should be empty")
	}
}
