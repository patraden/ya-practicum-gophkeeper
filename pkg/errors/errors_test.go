package errors_test

import (
	"errors"
	"fmt"
	"testing"

	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/stretchr/testify/require"
)

//nolint:err113 // reason: testing errors requires dynamic errors.
func TestErrInternal(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("database failure")
	internalErr := &e.InternalError{Err: originalErr}

	gotMsg := internalErr.Error()
	wantMsg := "[internal error] database failure"

	require.Equal(t, wantMsg, gotMsg)

	unwrapped := internalErr.Unwrap()
	require.ErrorIs(t, unwrapped, originalErr)
	require.NotErrorIs(t, unwrapped, internalErr)

	require.ErrorIs(t, internalErr, e.ErrInternal)
	require.ErrorIs(t, internalErr, originalErr)

	wrappedErr := fmt.Errorf("extra context: %w", internalErr)
	require.ErrorIs(t, wrappedErr, internalErr)
	require.ErrorIs(t, wrappedErr, originalErr)
}
