package model

type BinaryType string

type Binary struct {
	UserID      string     `json:"user_id"`
	Type        BinaryType `json:"-"`
	Key         string     `json:"key"`
	Description string     `json:"description"`
}
