package keystore_test

import (
	"context"
	"testing"

	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//nolint:funlen // reason: table driven testing functions are ok to be long
func TestGRPCServerStatusValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		isLoaded    bool
		method      string
		expectError bool
		expectCode  codes.Code
	}{
		{
			name:        "request allowed when keystore is loaded",
			isLoaded:    true,
			method:      pb.SecretService_SecretUpdateInit_FullMethodName,
			expectError: false,
		},
		{
			name:        "request blocked when server is sealed",
			isLoaded:    false,
			method:      pb.SecretService_SecretUpdateInit_FullMethodName,
			expectError: true,
			expectCode:  codes.Unavailable,
		},
		{
			name:        "unseal method allowed when keystore is not loaded",
			isLoaded:    false,
			method:      pb.AdminService_Unseal_FullMethodName,
			expectError: false,
		},
		{
			name:        "login method allowed when keystore is not loaded",
			isLoaded:    false,
			method:      pb.UserService_Login_FullMethodName,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockKS := mock.NewMockKeystore(ctrl)
			mockKS.EXPECT().IsLoaded().Return(tt.isLoaded).AnyTimes()

			interceptor := keystore.GRPCServerStatusValidator(mockKS)

			handler := func(_ context.Context, _ any) (any, error) {
				return "ok", nil
			}

			resp, err := interceptor(
				context.Background(),
				"dummy request",
				&grpc.UnaryServerInfo{FullMethod: tt.method},
				handler,
			)

			if tt.expectError {
				require.Error(t, err)
				st, _ := status.FromError(err)
				require.Equal(t, tt.expectCode, st.Code())
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, "ok", resp)
			}
		})
	}
}
