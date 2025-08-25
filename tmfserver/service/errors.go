package service

import "fmt"

// ErrObjectExists is returned when trying to create an object that already exists.
type ErrObjectExists struct {
	ID   string
	Type string
}

func (e *ErrObjectExists) Error() string {
	return fmt.Sprintf("object with id %s and type %s already exists", e.ID, e.Type)
}

func (e *ErrObjectExists) Is(target error) bool {
	switch target.(type) {
	case *ErrObjectExists:
		return true
	default:
		return false
	}
}

// ErrObjectConflict is returned when trying to update an object with a version that is not the latest.
type ErrObjectConflict struct {
	ID      string
	Type    string
	Version string
}

func (e *ErrObjectConflict) Error() string {
	return fmt.Sprintf("conflict updating object with id %s and type %s. Version %s is not the latest", e.ID, e.Type, e.Version)
}
