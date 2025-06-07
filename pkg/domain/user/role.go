package user

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Role string

func (r Role) String() string {
	return string(r)
}

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
