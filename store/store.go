package store

import "fmt"

const (
	// ErrNotFound signals the data could not be found
	ErrNotFound ErrorCode = iota + 1
)

// ErrorCode defines Store operation error code
type ErrorCode int

// String method implementation to satisfy fmt.Stringer interface
func (ec ErrorCode) String() string {
	switch ec {
	case ErrNotFound:
		return "NotFound"
	default:
		return "Unknown"
	}
}

// Error encapsulates ErrorCode and adds a simple error description
type Error struct {
	// Code is error code
	Code ErrorCode
	// Msg is error description message
	Msg error
}

// Error interface implementation to satisfy builtin error interface
func (e *Error) Error() string {
	code := e.Code.String()
	if e.Msg != nil {
		return fmt.Sprintf("Store error: %v %v", code, e.Msg)
	}

	return fmt.Sprintf("Store error: %v", code)
}

// Store implements basic data store
type Store interface {
	// Write writes data to store
	Write(p []byte) (int, error)
	// Read reads data from store
	Read(p []byte) (int, error)
}
