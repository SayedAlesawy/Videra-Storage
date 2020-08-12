package replication

// Replica is a model for replication mode info
type Replica struct {
	URL string //URL to send data to
	ID  string //File ID to append to
}
