package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/user"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/logger"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/infra/pg"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/mock"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func defaultSecretInitRequest(t *testing.T) *secret.InitRequest {
	t.Helper()

	usr := user.New("test_user", user.RoleUser)

	return &secret.InitRequest{
		UserID:          usr.ID,
		User:            usr,
		SecretID:        uuid.New(),
		SecretName:      "test.txt",
		S3URL:           "some-url",
		VersionID:       uuid.New(),
		ParentVersionID: uuid.Nil,
		RequestType:     secret.RequestTypePut,
		Token:           123,
		ClientInfo:      "test-client",
		SecretSize:      1024,
		SecretHash:      []byte("hash"),
		SecretDEK:       []byte("dek"),
		MetaData:        secret.MetaData{"key": "value"},
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	}
}

type mockBehavior func(
	t *testing.T,
	pool pgxmock.PgxPoolIface,
	idClient *mock.MockIdentityManager,
	s3Client *mock.MockServerOperator,
	req *secret.InitRequest,
)

func mockSuccessCase(
	t *testing.T,
	pool pgxmock.PgxPoolIface,
	idClient *mock.MockIdentityManager,
	s3Client *mock.MockServerOperator,
	req *secret.InitRequest,
) {
	t.Helper()

	meta, err := req.MetaData.MarshalJSON()
	require.NoError(t, err)

	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"user_id", "secret_id", "secret_name", "s3_url", "version_id", "parent_version_id",
			"request_type", "token", "client_info", "secret_size", "secret_hash", "secret_dek",
			"meta", "created_at", "expires_at",
		}).AddRow(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		))

	idClient.EXPECT().
		GetToken(gomock.Any(), req.User).
		Return(&user.IdentityToken{AccessToken: "token"}, nil)

	s3Client.EXPECT().
		AssumeRole(gomock.Any(), "token", req.UploadDuration()).
		Return(&s3.TemporaryCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}, nil)
}

func mockInvalidParentVersion(
	t *testing.T,
	pool pgxmock.PgxPoolIface,
	_ *mock.MockIdentityManager,
	_ *mock.MockServerOperator,
	req *secret.InitRequest,
) {
	t.Helper()

	meta, err := req.MetaData.MarshalJSON()
	require.NoError(t, err)

	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnError(sql.ErrNoRows)
}

func mockConflictVersionOrTime(
	t *testing.T,
	pool pgxmock.PgxPoolIface,
	_ *mock.MockIdentityManager,
	_ *mock.MockServerOperator,
	req *secret.InitRequest,
) {
	t.Helper()

	meta, err := req.MetaData.MarshalJSON()
	require.NoError(t, err)

	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"user_id", "secret_id", "secret_name", "s3_url", "version_id", "parent_version_id",
			"request_type", "token", "client_info", "secret_size", "secret_hash", "secret_dek",
			"meta", "created_at", "expires_at",
		}).AddRow(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, uuid.New(), req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt.Add(-time.Hour), req.ExpiresAt,
		))
}

func mockIdentityTokenError(
	t *testing.T,
	pool pgxmock.PgxPoolIface,
	idClient *mock.MockIdentityManager,
	_ *mock.MockServerOperator,
	req *secret.InitRequest,
) {
	t.Helper()

	meta, err := req.MetaData.MarshalJSON()
	require.NoError(t, err)

	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"user_id", "secret_id", "secret_name", "s3_url", "version_id", "parent_version_id",
			"request_type", "token", "client_info", "secret_size", "secret_hash", "secret_dek",
			"meta", "created_at", "expires_at",
		}).AddRow(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		))

	idClient.EXPECT().
		GetToken(gomock.Any(), gomock.Any()).
		Return(nil, e.ErrUnavailable)
}

func mockSuccessAfterRetriesCase(
	t *testing.T,
	pool pgxmock.PgxPoolIface,
	idClient *mock.MockIdentityManager,
	s3Client *mock.MockServerOperator,
	req *secret.InitRequest,
) {
	t.Helper()

	meta, err := req.MetaData.MarshalJSON()
	require.NoError(t, err)

	// First retry: ConnectionFailure
	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionFailure})

	// Second retry: SQLClientUnableToEstablishSQLConnection
	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.SQLClientUnableToEstablishSQLConnection})

	// Third attempt succeeds
	pool.ExpectQuery(`WITH candidate\(parent_version_id\)`).
		WithArgs(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"user_id", "secret_id", "secret_name", "s3_url", "version_id", "parent_version_id",
			"request_type", "token", "client_info", "secret_size", "secret_hash", "secret_dek",
			"meta", "created_at", "expires_at",
		}).AddRow(
			req.UserID, req.SecretID, req.SecretName, req.S3URL, req.VersionID, req.ParentVersionID,
			pg.RequestType(req.RequestType), req.Token, req.ClientInfo, req.SecretSize,
			req.SecretHash, req.SecretDEK, meta, req.CreatedAt, req.ExpiresAt,
		))

	idClient.EXPECT().
		GetToken(gomock.Any(), req.User).
		Return(&user.IdentityToken{AccessToken: "token"}, nil)

	s3Client.EXPECT().
		AssumeRole(gomock.Any(), "token", req.UploadDuration()).
		Return(&s3.TemporaryCredentials{AccessKeyID: "key", SecretAccessKey: "secret"}, nil)
}

//nolint:funlen // reason: allow table driven testing func to be lengthy.
func TestSecretRepoCreateSecretInitRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectErr    error
	}{
		{
			name:         "success",
			mockBehavior: mockSuccessCase,
			expectErr:    nil,
		},
		{
			name:         "invalid parent version",
			mockBehavior: mockInvalidParentVersion,
			expectErr:    e.ErrInvalidInput,
		},
		{
			name:         "conflict version or outdated time",
			mockBehavior: mockConflictVersionOrTime,
			expectErr:    e.ErrExists,
		},
		{
			name:         "identity token error",
			mockBehavior: mockIdentityTokenError,
			expectErr:    e.ErrInternal,
		},
		{
			name:         "success after retrying",
			mockBehavior: mockSuccessAfterRetriesCase,
			expectErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			req := defaultSecretInitRequest(t)
			idClient := mock.NewMockIdentityManager(ctrl)
			s3Client := mock.NewMockServerOperator(ctrl)
			log := logger.Stdout(zerolog.Disabled).GetZeroLog()
			db := &pg.DB{ConnPool: mockPool}
			repo := repository.NewSecretRepo(db, s3Client, idClient, log)

			tt.mockBehavior(t, mockPool, idClient, s3Client, req)

			result, err := repo.CreateSecretInitRequest(context.Background(), req)
			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)

				if errors.Is(err, e.ErrExists) {
					require.NotNil(t, result)
				} else {
					require.Nil(t, result)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}

			require.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

// func TestSecretRepoCreateSecretInitRequestIntegration(t *testing.T) {
// 	t.Parallel()
// 	t.Skip("Enable when dev environment (pod and containers) are up and running")

// 	log := logger.StdoutConsole(zerolog.DebugLevel).GetZeroLog()
// 	cfg := config.DefaultConfig()
// 	cfg.DatabaseDSN = "postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable"
// 	cfg.S3TLSCertPath = "../../../deployments/.certs/ca.cert"

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	db, err := pg.NewDB(ctx, cfg.DatabaseDSN)
// 	require.NoError(t, err)

// 	err = db.Ping(ctx)
// 	require.NoError(t, err)

// 	minIOClient, err := minio.NewClient(cfg, log)
// 	require.NoError(t, err)

// 	identityClient, err := identity.KeycloakPGManager(cfg, db, log)
// 	require.NoError(t, err)

// 	repo := repository.NewSecretRepo(db, minIOClient, identityClient, log)

// 	uid, err := uuid.Parse("e96aff31-bc4d-426e-a559-ad51bc9859e9")
// 	require.NoError(t, err)

// 	sid, err := uuid.Parse("6c9ba89f-eefe-4625-a781-394a8308e93e")
// 	require.NoError(t, err)

// 	iters := int32(100)
// 	failed := iters - 1
// 	var count atomic.Int32
// 	wg := sync.WaitGroup{}

// 	for range iters {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()

// 			// create request with new version
// 			req := secret.InitRequest{
// 				UserID:        uid,
// 				SecretID:      sid,
// 				SecretName:    "example.txt",
// 				S3URL:         "s3_path",
// 				Version:       uuid.New(),
// 				ParentVersion: uuid.Nil,
// 				RequestType:   secret.RequestTypePut,
// 				Token:         123,
// 				ClientInfo:    "test-agent",
// 				SecretSize:    100,
// 				SecretHash:    []byte("hash"),
// 				SecretDEK:     []byte("dek"),
// 				MetaData:      secret.MetaData{"k1": "v1"},
// 				CreatedAt:     time.Now().UTC(),
// 				ExpiresAt:     time.Now().Add(time.Hour),
// 			}

// 			result, err := repo.CreateSecretInitRequest(ctx, &req)
// 			if err != nil && errors.Is(err, e.ErrExists) {
// 				count.Add(1)
// 			} else {
// 				require.NotNil(t, result)
// 			}
// 		}()
// 	}

// 	wg.Wait()
// 	assert.Equal(t, failed, count.Load())
// }
