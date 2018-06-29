package sshproviderconfig

// EtcdMember defines the configuration of an etcd member.
type EtcdMember struct {
	// ID is the member ID for this member.
	ID uint64 `json:"ID"`
	// Name is the human-readable name of the member.
	Name string `json:"name"`
	// PeerURLs is the list of URLs the member exposes to the cluster for communication.
	PeerURLs []string `json:"peerURLs"`
	// ClientURLs is the list of URLs the member exposes to clients for communication.
	ClientURLs []string `json:"clientURLs"`
}
