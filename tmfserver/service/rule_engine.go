package service

import "github.com/hesusruiz/isbetmf/pdp"

// RuleEngine abstracts authorization decisions for TMF requests.
// It allows plugging in different rule engines or mock implementations for testing.
type RuleEngine interface {
	Authorize(input pdp.StarTMFMap) (bool, error)
}
