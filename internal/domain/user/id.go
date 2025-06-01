package user

import "github.com/google/uuid"

// ID represents a unique identifier for a user.
type ID uuid.UUID

// String returns the string representation of the ID.
func (u ID) String() string {
	return uuid.UUID(u).String()
}

// IsNil checks whether the ID is nil (zero value).
func (u ID) IsNil() bool {
	return uuid.UUID(u) == uuid.Nil
}

// NewUserID generates and returns a new unique ID.
func NewUserID() ID {
	return ID(uuid.New())
}

// ParseID converts a string representation of a UUID into a ID.
func ParseID(id string) (ID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return ID(uuid.Nil), err
	}

	return ID(uid), nil
}
