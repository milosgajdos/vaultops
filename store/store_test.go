package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode(t *testing.T) {
	ec := ErrNotFound
	assert.Equal(t, ec.String(), "NotFound")

	ec = ErrorCode(100)
	assert.Equal(t, ec.String(), "Unknown")
}

func TestError(t *testing.T) {
	errMsg := "Caboom"

	e := Error{
		Code: ErrNotFound,
		Msg:  fmt.Errorf(errMsg),
	}
	assert.Equal(t, e.Error(), fmt.Sprintf("Store error: %v %v", e.Code, e.Msg.Error()))

	e = Error{
		Code: ErrNotFound,
		Msg:  nil,
	}
	assert.Equal(t, e.Error(), fmt.Sprintf("Store error: %v", e.Code))
}
