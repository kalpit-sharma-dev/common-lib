package leaderelection

// Interface : Leader Election Service
type Interface interface {
	// Old implementation of leaderelection
	BecomeALeader() (peerID int, isLeader bool, err error)

	// New implementation
	RegisterCandidate(electionResource string, clientName string) error
	StartElection(callback func()) error
	ResignCandidate()
}
