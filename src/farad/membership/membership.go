package membership

import (
	"errors"
	"farad/timerqueue"
	"time"
)

// MemberContext IS UNSYNCHRONIZED
type MemberContext struct {
	member_keys map[string]string
	tq          *timerqueue.TimerQueue
}

func NewMemberContext(expiration_time time.Duration) *MemberContext {
	return &MemberContext{
		member_keys: map[string]string{},
		tq:          timerqueue.NewTimerQueue(expiration_time),
	}
}

// UpdatePing(...) returns did_revision_occur
func (m *MemberContext) UpdatePing(principal string, key string) (bool, error) {
	if principal == "" {
		return false, errors.New("should not be an empty principal")
	}
	if key == "" {
		return false, errors.New("should not be an empty key")
	}
	old_key := m.member_keys[principal]
	revision := false
	if old_key == "" || old_key != key {
		m.member_keys[principal] = key
		revision = true
	}
	// to track when this should expire
	m.tq.Add(principal)
	for {
		found, elem := m.tq.Query()
		if !found {
			break
		}
		delete(m.member_keys, elem)
	}
	return revision, nil
}

// Snapshot returns a map of principals -> public keys.
func (m *MemberContext) Snapshot() map[string]string {
	result := map[string]string{}
	for principal, mem := range m.member_keys {
		result[principal] = mem
	}
	return result
}

func (m *MemberContext) Subshot(subset []string) map[string]string {
	result := map[string]string{}
	for _, principal := range subset {
		found := m.member_keys[principal]
		if found == "" {
			continue
		}
		result[principal] = found
	}
	return result
}
