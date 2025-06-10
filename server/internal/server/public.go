package server

import pb "github.com/patraden/ya-practicum-gophkeeper/pkg/proto/gophkeeper/v1"

func PublicGRPCMethods(method string) bool {
	switch method {
	case
		pb.UserService_Login_FullMethodName,
		pb.UserService_Register_FullMethodName:
		return true
	}

	return false
}
