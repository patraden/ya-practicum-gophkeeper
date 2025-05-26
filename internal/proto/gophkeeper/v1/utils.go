package proto

// IsPublicGRPCMethod returns true if the method is exempt from authentication checks.
func IsPublicGRPCMethod(method string) bool {
	switch method {
	case
		"/gophkeeper.v1.UserService/Login",
		"/gophkeeper.v1.UserService/Register":
		return true
	default:
		return false
	}
}
