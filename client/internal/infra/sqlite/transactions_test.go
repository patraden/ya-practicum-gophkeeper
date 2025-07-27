//nolint:funlen // testing transaction logic with multiple cases
package sqlite_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/patraden/ya-practicum-gophkeeper/client/internal/infra/sqlite"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	name          string
	queryFn       sqlite.QueryFunc
	mockSetup     func(mock sqlmock.Sqlmock)
	expectedError error
}

func TestWithinTrx(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:    "success",
			queryFn: func(*sqlite.Queries) error { return nil },
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:    "query function fails",
			queryFn: func(*sqlite.Queries) error { return e.ErrInternal },
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			expectedError: e.ErrInternal,
		},
		{
			name:    "begin fails",
			queryFn: func(*sqlite.Queries) error { return nil },
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(e.ErrUnavailable)
			},
			expectedError: e.ErrUnavailable,
		},
		{
			name:    "commit fails",
			queryFn: func(*sqlite.Queries) error { return nil },
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(e.ErrInternal)
			},
			expectedError: e.ErrInternal,
		},
		{
			name:    "rollback fails",
			queryFn: func(*sqlite.Queries) error { return e.ErrInternal },
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback().WillReturnError(e.ErrInternal)
			},
			expectedError: e.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			// We'll pass a dummy Queries instance since sqlc-generated code is not needed for logic path testing.
			q := sqlite.New(db)

			fn := sqlite.WithinTrx(context.Background(), db, nil, tt.queryFn)
			err = fn(q)

			require.ErrorIs(t, err, tt.expectedError)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
