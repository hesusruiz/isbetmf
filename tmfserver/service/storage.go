package service

import (
	"net/url"

	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// TMFStorage abstracts persistence operations for TMF objects.
// It is used for plugging-in different persistence systems
type TMFStorage interface {
	CreateObject(obj *repo.TMFRecord) error
	GetObject(id, resourceName string) (*repo.TMFRecord, error)
	UpdateObject(obj *repo.TMFRecord) error
	UpsertObject(obj *repo.TMFRecord) error
	DeleteObject(id, resourceName string) error
	ListObjects(healthRequest bool, resourceName string, queryParams url.Values, filter repo.ObjectFilter) ([]repo.TMFRecord, error)
}
