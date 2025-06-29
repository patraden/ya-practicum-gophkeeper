package app

import (
	"context"
	"fmt"

	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/dto"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/crypto/keystore"
	"github.com/patraden/ya-practicum-gophkeeper/server/internal/repository"
)

// SecretUseCase defines the core operations related to user sercrets.
type SecretUseCase interface {
	InitUploadRequest(ctx context.Context, req *secret.InitRequest) (*dto.SecretUploadInitResponse, error)
}

// SecretUC implements the SecretUseCase interface.
type SecretUC struct {
	SecretUseCase
	repoSecret repository.SecretRepository
	repoUser   repository.UserRepository
	keyStore   keystore.Keystore
}

func NewSecretUC(
	repoSecret repository.SecretRepository,
	repoUser repository.UserRepository,
	keyStore keystore.Keystore,
) *SecretUC {
	return &SecretUC{
		repoSecret: repoSecret,
		repoUser:   repoUser,
		keyStore:   keyStore,
	}
}

func (uc *SecretUC) InitUploadRequest(
	ctx context.Context,
	req *secret.InitRequest,
) (*dto.SecretUploadInitResponse, error) {
	rek, err := uc.keyStore.Get()
	if err != nil {
		return nil, err
	}

	err = req.Validate(rek)
	if err != nil {
		return nil, err
	}

	usr, err := uc.repoUser.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	req.User = usr

	resReq, err := uc.repoSecret.CreateSecretInitRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resReq.S3Creds == nil {
		return nil, fmt.Errorf("[%w] empty s3 credentials", e.ErrInternal)
	}

	return &dto.SecretUploadInitResponse{
		UserID:          resReq.UserID.String(),
		SecretID:        resReq.SecretID.String(),
		VersionID:       resReq.VersionID.String(),
		ParentVersionID: resReq.ParentVersionID.String(),
		S3URL:           resReq.S3URL,
		Token:           resReq.Token,
		S3Creds:         *resReq.S3Creds,
	}, nil
}
