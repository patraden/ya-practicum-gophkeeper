package user

import pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"

const (
	RoleUser        = pb.UserRole_USER_ROLE_USER
	RoleAdmin       = pb.UserRole_USER_ROLE_ADMIN
	RoleUnspecified = pb.UserRole_USER_ROLE_UNSPECIFIED
)

type Role = pb.UserRole
