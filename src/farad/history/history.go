package history

// History IS UNSYNCHRONIZED
type History struct {
	recent             []string // once a subsequence of this is written, it is never unwritten
	start_logical_time uint64
	max_to_keep        int
}

func NewHistory(max_to_keep int) *History {
	if max_to_keep == 0 {
		panic("invalid setup for NewHistory")
	}
	return &History{max_to_keep: max_to_keep}
}

func (v *History) AddUpdate(value string) uint64 {
	time := v.start_logical_time + uint64(len(v.recent))
	v.recent = append(v.recent, value)
	if len(v.recent) >= v.max_to_keep {
		if v.max_to_keep == 0 {
			panic("history.History object should have been constructed with history.NewHistory!")
		}
		// move the second half to the first half and discard the old first half
		midpoint := v.max_to_keep / 2
		copy(v.recent[:midpoint], v.recent[midpoint:])
		v.recent = v.recent[:midpoint]
		v.start_logical_time += uint64(midpoint)
	}
	return time
}

func (v *History) Since(earliest uint64) (bool, []string, uint64) {
	now := v.start_logical_time + uint64(len(v.recent))
	if earliest >= v.start_logical_time {
		slice := v.recent[earliest-v.start_logical_time:]
		result := make([]string, len(slice))
		copy(result, slice)
		return true, result, now
	} else {
		return false, nil, now
	}
}
