package history

import (
	"fmt"
	"testing"
)

func TestNewHistory(t *testing.T) {
	history := NewHistory(144)
	if history.start_logical_time != 0 {
		t.Error("wrong initial logical time")
	}
	if history.max_to_keep != 144 {
		t.Error("wrong max_to_keep")
	}
	if len(history.recent) != 0 {
		t.Error("wrong size of recent")
	}
}

func test_start_addupdate(t *testing.T) *History {
	history := NewHistory(12)
	for i := 0; i < 11; i++ {
		timestamp := history.AddUpdate(fmt.Sprintf("update-%d", i))
		if timestamp != uint64(i) {
			t.Error("wrong timestamp")
		}
		if len(history.recent) != i+1 {
			t.Error("wrong number of elements")
		}
		for j := 0; j <= i; j++ {
			if history.recent[j] != fmt.Sprintf("update-%d", j) {
				t.Error("wrong state of history")
			}
		}
		if history.start_logical_time != 0 {
			t.Error("wrong start of logical time")
		}
	}
	return history
}

func TestHistory_AddUpdate(t *testing.T) {
	test_start_addupdate(t)
}

func test_start_addupdate_and_trim_once(t *testing.T) *History {
	history := test_start_addupdate(t)
	if history.AddUpdate("update-11") != 11 {
		t.Error("wrong timestamp")
	}
	if len(history.recent) != 6 {
		t.Error("wrong number of elements")
	}
	for j := 6; j < 12; j++ {
		if history.recent[j-6] != fmt.Sprintf("update-%d", j) {
			t.Error("wrong state of history")
		}
	}
	if history.start_logical_time != 6 {
		t.Error("wrong start of logical time")
	}
	return history
}

func TestHistory_Trim_Once(t *testing.T) {
	test_start_addupdate_and_trim_once(t)
}

func TestHistory_Trim_Repeat(t *testing.T) {
	history := test_start_addupdate_and_trim_once(t)
	for i := 12; i < 1000; i++ {
		timestamp := history.AddUpdate(fmt.Sprintf("update-%d", i))
		if timestamp != uint64(i) {
			t.Error("wrong timestamp")
		}
		expected_element_count := (i+1)%6 + 6
		if expected_element_count != len(history.recent) {
			t.Errorf("wrong number of elements")
		}
		if history.start_logical_time != uint64(i-expected_element_count+1) {
			t.Errorf("wrong start of logical time: %d instead of expected %d", history.start_logical_time, i-expected_element_count)
		}
		for j := 0; j < expected_element_count; j++ {
			if history.recent[j] != fmt.Sprintf("update-%d", int(history.start_logical_time)+j) {
				t.Errorf("wrong element in history")
			}
		}
	}
}

func TestHistory_Since_Early(t *testing.T) {
	history := NewHistory(101)
	for i := 0; i < 100; i++ {
		timestamp := history.AddUpdate(fmt.Sprintf("update-%d", i))
		if timestamp != uint64(i) {
			t.Error("wrong timestamp")
		}
		found, results, now := history.Since(timestamp + 1)
		if now != timestamp+1 {
			t.Error("wrong timestamp from Since")
		}
		if !found {
			t.Errorf("should always be found for most recent")
		}
		if len(results) != 0 {
			t.Errorf("shouldn't be anything since timestamp")
		}
		for j := 0; j <= i; j++ {
			found, results, now := history.Since(uint64(j))
			if now != timestamp+1 {
				t.Errorf("wrong timestamp from Since")
			}
			if !found {
				t.Errorf("should always be found for these")
			}
			if len(results) != int(timestamp)+1-j {
				t.Errorf("wrong number of results from Since")
			}
			for off, v := range results {
				if v != fmt.Sprintf("update-%d", off+j) {
					t.Errorf("wrong element of results")
				}
			}
		}
	}
}

func TestHistory_Since_Later(t *testing.T) {
	history := test_start_addupdate_and_trim_once(t)
	for i := 12; i < 1000; i++ {
		timestamp := history.AddUpdate(fmt.Sprintf("update-%d", i))
		if timestamp != uint64(i) {
			t.Error("wrong timestamp")
		}
		for j := 0; j <= i+1; j++ {
			found, results, now := history.Since(uint64(j))
			if now != timestamp+1 {
				t.Errorf("wrong timestamp from Since")
			}
			if found != (uint64(j) >= history.start_logical_time) {
				t.Errorf("found mismatch")
			}
			if !found {
				continue
			}
			if len(results) != int(timestamp)+1-j {
				t.Errorf("wrong number of results from Since")
			}
			for off, v := range results {
				if v != fmt.Sprintf("update-%d", off+j) {
					t.Errorf("wrong element of results")
				}
			}
		}
	}
}
