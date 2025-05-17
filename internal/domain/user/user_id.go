package user

import "github.com/google/uuid"

// UserID represents a unique identifier for a user.
type ID uuid.UUID

// String returns the string representation of the UserID.
func (u ID) String() string {
	return uuid.UUID(u).String()
}

// IsNil checks whether the UserID is nil (zero value).
func (u ID) IsNil() bool {
	return uuid.UUID(u) == uuid.Nil
}

// NewUserID generates and returns a new unique UserID.
func NewUserID() ID {
	return ID(uuid.New())
}

// ParseUserID converts a string representation of a UUID into a UserID.
func ParseUserID(id string) (ID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ID(uuid.Nil), err
	}

	return ID(uid), nil
}
