package dto

type ServerToken struct {
	UserID string
	Token  string
	TTL    uint32
}
