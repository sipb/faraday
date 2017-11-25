package common

const FARADAY_PROTOCOL_VERSION = 1

// updates the current state for us and queries the current state for everyone
type FaradRequest struct {
	Version        int
	Key            string
	Cursor         uint64
	IncludeMember  string
	ServerInstance string
}

// the response will include everything that has been changed since the specified cursor (or ever, if the cursor is 0),
// plus the member referenced by IncludeMember. The server may also include any other information, should it choose, and
// if a member no longer exists, it will not be included in the result.
type FaradResponse struct {
	CurrentCluster map[string]string // map of principals -> public keys
	Cursor         uint64
	ServerInstance string
}
