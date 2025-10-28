package types

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrPermissionDenied(t *testing.T) {
	err := ErrPermissionDenied{Path: "/test/path"}
	assert.Equal(t, "permission denied: /test/path", err.Error())
}

func TestErrPathNotFound(t *testing.T) {
	err := ErrPathNotFound{Path: "/missing/path"}
	assert.Equal(t, "path not found: /missing/path", err.Error())
}

func TestErrTrashFull(t *testing.T) {
	err := ErrTrashFull{CurrentSize: 1000, MaxSize: 500}
	assert.Equal(t, "trash directory is full", err.Error())
}

func TestErrPluginLoadFailed(t *testing.T) {
	t.Run("with reason", func(t *testing.T) {
		reason := fmt.Errorf("file not found")
		err := ErrPluginLoadFailed{PluginName: "test-plugin", Reason: reason}
		assert.Equal(t, "failed to load plugin 'test-plugin': file not found", err.Error())
	})

	t.Run("without reason", func(t *testing.T) {
		err := ErrPluginLoadFailed{PluginName: "test-plugin"}
		assert.Equal(t, "failed to load plugin 'test-plugin'", err.Error())
	})
}

func TestErrPluginLoadFailed_Unwrap(t *testing.T) {
	reason := fmt.Errorf("underlying error")
	err := ErrPluginLoadFailed{PluginName: "test-plugin", Reason: reason}

	// Test that Unwrap returns the reason
	assert.Equal(t, reason, err.Unwrap())

	// Test that errors.Is works with wrapped errors
	assert.True(t, errors.Is(err, reason))
}

func TestErrorWrapping(t *testing.T) {
	// Test error wrapping with fmt.Errorf
	baseErr := ErrPathNotFound{Path: "/test"}
	wrappedErr := fmt.Errorf("failed to process: %w", baseErr)

	// Test errors.Is
	assert.True(t, errors.Is(wrappedErr, ErrPathNotFound{Path: "/test"}))

	// Test errors.As
	var pathErr ErrPathNotFound
	assert.True(t, errors.As(wrappedErr, &pathErr))
	assert.Equal(t, "/test", pathErr.Path)
}

func TestErrorTypes(t *testing.T) {
	// Test that custom errors can be distinguished
	permErr := ErrPermissionDenied{Path: "/test"}
	pathErr := ErrPathNotFound{Path: "/test"}

	var pe ErrPermissionDenied
	var pne ErrPathNotFound

	assert.True(t, errors.As(permErr, &pe))
	assert.False(t, errors.As(permErr, &pne))

	assert.True(t, errors.As(pathErr, &pne))
	assert.False(t, errors.As(pathErr, &pe))
}
