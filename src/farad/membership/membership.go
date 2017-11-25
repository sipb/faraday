package membership

import (
	"time"
)

// Member IS UNSYNCHRONIZED
type Member struct {
	Key      string
	LastPing time.Time
}

// MemberContext IS UNSYNCHRONIZED
type MemberContext struct {
	members map[string]*Member
}

// UpdatePing(...) returns did_revision_occur
func (m *MemberContext) UpdatePing(principal string, key string) bool {
	now := time.Now()
	member := m.members[principal]
	if member == nil {
		m.members[principal] = &Member{
			Key:      key,
			LastPing: now,
		}
		return true
	} else {
		member.LastPing = now
		if member.Key != key {
			member.Key = key
			return true
		}
		return false // no revision
	}
}

func (m *MemberContext) EliminateStale(expiration_interval time.Duration) {
	expired_before := time.Now().Add(-expiration_interval)
	to_eliminate := []string{}
	for principal, mem := range m.members {
		if mem.LastPing.Before(expired_before) {
			to_eliminate = append(to_eliminate, principal)
		}
	}
	for _, elim := range to_eliminate {
		delete(m.members, elim)
	}
}

// Snapshot returns a map of principals -> public keys.
func (m *MemberContext) Snapshot() map[string]string {
	result := map[string]string{}
	for principal, mem := range m.members {
		result[principal] = mem.Key
	}
	return result
}

func (m *MemberContext) Subshot(subset []string) map[string]string {
	if len(subset) == 0 {
		return map[string]string{}
	}
	result := map[string]string{}
	for _, principal := range subset {
		found := m.members[principal]
		if found == nil {
			continue
		}
		result[principal] = found.Key
	}
	return result
}
