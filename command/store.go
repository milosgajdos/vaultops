package command

// Store implements basic data store
type Store interface {
	// Write writes data to store
	Write(p []byte) (int, error)
	// Read reads data from store
	Read(p []byte) (int, error)
}
