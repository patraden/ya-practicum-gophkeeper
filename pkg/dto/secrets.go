package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/domain/secret"
	e "github.com/patraden/ya-practicum-gophkeeper/pkg/errors"
	pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"
	"github.com/patraden/ya-practicum-gophkeeper/pkg/s3"
)

// SecretUploadInitRequest represents an upload request that is in progress.
type SecretUploadInitRequest struct {
	UserID          string `json:"user_id"`
	SecretID        string `json:"secret_id"`
	SecretName      string `json:"secret_name"`
	VersionID       string `json:"version_id"`
	ParentVersionID string `json:"parent_version_id,omitempty"`
	ClientInfo      string `json:"client_info"`
	SecretSize      int64  `json:"secret_size"`
	SecretHash      []byte `json:"secret_hash,omitempty"`
	SecretDEK       []byte `json:"secret_dek,omitempty"`
	MetaData        string `json:"meta,omitempty"`
}

func SecretUploadInitRequestFromProto(req *pb.SecretUpdateInitRequest) *SecretUploadInitRequest {
	return &SecretUploadInitRequest{
		UserID:          req.GetUserId(),
		SecretID:        req.GetSecretId(),
		SecretName:      req.GetSecretName(),
		VersionID:       req.GetVersionId(),
		ParentVersionID: req.GetParentVersionId(),
		ClientInfo:      req.GetClientInfo(),
		SecretSize:      req.GetSize(),
		SecretHash:      req.GetHash(),
		SecretDEK:       req.GetEncryptedDek(),
		MetaData:        req.GetMetadataJson(),
	}
}

func (r *SecretUploadInitRequest) ToDomain() (*secret.InitRequest, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid userID", e.ErrValidation)
	}

	secretID, err := uuid.Parse(r.SecretID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid secretID", e.ErrValidation)
	}

	versionID, err := uuid.Parse(r.VersionID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid versionID", e.ErrValidation)
	}

	var parentID uuid.UUID

	if r.ParentVersionID != "" {
		pparent, err := uuid.Parse(r.ParentVersionID)
		if err != nil {
			return nil, fmt.Errorf("[%w] invalid parent versionID", e.ErrValidation)
		}

		parentID = pparent
	}

	var metaData secret.MetaData

	if err := metaData.UnmarshalJSON([]byte(r.MetaData)); err != nil {
		return nil, fmt.Errorf("[%w] invalid secret metadata", e.ErrUnmarshal)
	}

	now := time.Now().UTC()

	return &secret.InitRequest{
		UserID:          userID,
		SecretID:        secretID,
		SecretName:      r.SecretName,
		VersionID:       versionID,
		ParentVersionID: parentID,
		RequestType:     secret.RequestTypePut,
		ClientInfo:      r.ClientInfo,
		SecretSize:      r.SecretSize,
		SecretHash:      r.SecretHash,
		SecretDEK:       r.SecretDEK,
		MetaData:        metaData,
		CreatedAt:       now,
	}, nil
}

// SecretUploadInitResponse represents response to secret upload request.
type SecretUploadInitResponse struct {
	UserID          string `json:"user_id"`
	SecretID        string `json:"secret_id"`
	VersionID       string `json:"version_id"`
	ParentVersionID string `json:"parent_version_id,omitempty"`
	S3URL           string `json:"s3_url"`
	Token           int64  `json:"token"`
	S3Creds         s3.TemporaryCredentials
}

func (resp *SecretUploadInitResponse) ToProto() *pb.SecretUpdateInitResponse {
	return &pb.SecretUpdateInitResponse{
		UserId:          resp.UserID,
		SecretId:        resp.SecretID,
		VersionId:       resp.VersionID,
		ParentVersionId: resp.ParentVersionID,
		S3Url:           resp.S3URL,
		Token:           resp.Token,
		Credentials:     resp.S3Creds.ToProto(),
	}
}

// SecretUploadCommitRequest represents a finalized and committed upload.
type SecretUploadCommitRequest struct {
	UserID          string `json:"user_id"`
	SecretID        string `json:"secret_id"`
	VersionID       string `json:"version_id"`
	ParentVersionID string `json:"parent_version_id,omitempty"`
	ClientInfo      string `json:"client_info"`
	SecretSize      int64  `json:"secret_size"`
	SecretHash      []byte `json:"secret_hash,omitempty"`
	SecretDEK       []byte `json:"secret_dek,omitempty"`
	Token           int64  `json:"token"`
}

func SecretUploadCommitRequestFromProto(req *pb.SecretUpdateCommitRequest) *SecretUploadCommitRequest {
	return &SecretUploadCommitRequest{
		UserID:          req.GetUserId(),
		SecretID:        req.GetSecretId(),
		VersionID:       req.GetVersionId(),
		ParentVersionID: req.GetParentVersionId(),
		ClientInfo:      req.GetClientInfo(),
		SecretSize:      req.GetSize(),
		SecretHash:      req.GetHash(),
		SecretDEK:       req.GetEncryptedDek(),
		Token:           req.GetToken(),
	}
}

func (r *SecretUploadCommitRequest) ToDomain() (*secret.CommitRequest, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid userID", e.ErrValidation)
	}

	secretID, err := uuid.Parse(r.SecretID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid secretID", e.ErrValidation)
	}

	version, err := uuid.Parse(r.VersionID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid versionID", e.ErrValidation)
	}

	var parentID uuid.UUID

	if r.ParentVersionID != "" {
		pparent, err := uuid.Parse(r.ParentVersionID)
		if err != nil {
			return nil, fmt.Errorf("[%w] invalid parent versionID", e.ErrValidation)
		}

		parentID = pparent
	}

	return &secret.CommitRequest{
		UserID:          userID,
		SecretID:        secretID,
		VersionID:       version,
		ParentVersionID: parentID,
		RequestType:     secret.RequestTypePut,
		ClientInfo:      r.ClientInfo,
		SecretSize:      r.SecretSize,
		SecretHash:      r.SecretHash,
		SecretDEK:       r.SecretDEK,
		Token:           r.Token,
	}, nil
}

type Secret struct {
	ID              string
	UserID          string
	SecretName      string
	VersionID       string
	ParentVersionID string
	FilePath        string
	SecretSize      int64
	SecretHash      []byte
	SecretDek       []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	InSync          bool
}

func (r *Secret) ToDomain() (*secret.Secret, error) {
	userID, err := uuid.Parse(r.UserID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid userID", e.ErrValidation)
	}

	secretID, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid secretID", e.ErrValidation)
	}

	versionID, err := uuid.Parse(r.VersionID)
	if err != nil {
		return nil, fmt.Errorf("[%w] invalid versionID", e.ErrValidation)
	}

	var parentID uuid.UUID

	if r.ParentVersionID != "" {
		pparent, err := uuid.Parse(r.ParentVersionID)
		if err != nil {
			return nil, fmt.Errorf("[%w] invalid parent versionID", e.ErrValidation)
		}

		parentID = pparent
	}

	version := secret.Version{
		ID:         versionID,
		UserID:     userID,
		SecretID:   secretID,
		ParentID:   parentID,
		S3URL:      r.FilePath,
		SecretSize: r.SecretSize,
		SecretHash: r.SecretHash,
		SecretDEK:  r.SecretDek,
		CreatedAt:  r.CreatedAt,
	}

	return &secret.Secret{
		ID:               secretID,
		UserID:           userID,
		Name:             r.SecretName,
		CurrentVersionID: versionID,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
		CurrentVersion:   &version,
		Meta:             nil,
	}, nil
}
