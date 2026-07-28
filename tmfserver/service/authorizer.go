package service

import "github.com/hesusruiz/isbetmf/pdp"

// Authorizer abstracts authorization decisions for TMF requests.
// It allows plugging in different rule engines or mock implementations for testing.
type Authorizer interface {
	Authorize(input pdp.StarTMFMap) (bool, error)
}
