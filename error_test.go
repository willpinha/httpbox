package httpbox

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewError_BasicCreation(t *testing.T) {
	code := http.StatusBadRequest
	message := "bad request"

	err := NewError(code, message)

	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Nil(t, err.Details)
	assert.Nil(t, err.Err)
	assert.False(t, err.Log)
}

func TestNewError_WithDetails(t *testing.T) {
	code := http.StatusBadRequest
	message := "validation failed"
	details := map[string]string{
		"field": "email",
		"error": "invalid format",
	}

	err := NewError(code, message, WithDetails(details))

	require.NotNil(t, err.Details)
	detailsMap, ok := err.Details.(map[string]string)
	require.True(t, ok, "expected details to be map[string]string")
	assert.Equal(t, "email", detailsMap["field"])
	assert.Equal(t, "invalid format", detailsMap["error"])
}

func TestNewError_WithInternalError(t *testing.T) {
	code := http.StatusInternalServerError
	message := "database error"
	internalErr := errors.New("connection timeout")

	err := NewError(code, message, WithInternalError(internalErr))

	require.NotNil(t, err.Err)
	assert.Equal(t, "connection timeout", err.Err.Error())
}

func TestNewError_WithLog(t *testing.T) {
	code := http.StatusInternalServerError
	message := "critical error"

	err := NewError(code, message, WithLog())

	assert.True(t, err.Log)
}

func TestNewError_WithAllOptions(t *testing.T) {
	code := http.StatusInternalServerError
	message := "complete error"
	details := "detailed information"
	internalErr := errors.New("underlying cause")

	err := NewError(
		code,
		message,
		WithDetails(details),
		WithInternalError(internalErr),
		WithLog(),
	)

	assert.Equal(t, code, err.Code)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, details, err.Details)
	assert.Equal(t, internalErr, err.Err)
	assert.True(t, err.Log)
}

func TestError_MultipleOptionsOrder(t *testing.T) {
	firstDetails := "first"
	secondDetails := "second"

	// Apply WithDetails twice - second should overwrite first
	err := NewError(
		http.StatusBadRequest,
		"test",
		WithDetails(firstDetails),
		WithDetails(secondDetails),
	)

	assert.Equal(t, secondDetails, err.Details)
}
