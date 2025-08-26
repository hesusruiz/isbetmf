package service

import (
	"net/url"

	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// Storage abstracts persistence operations for TMF objects.
type Storage interface {
	CreateObject(obj *repo.TMFObject) error
	GetObject(id, objectType string) (*repo.TMFObject, error)
	UpdateObject(obj *repo.TMFObject) error
	DeleteObject(id, objectType string) error
	ListObjects(objectType string, queryParams url.Values) ([]repo.TMFObject, int, error)
}
