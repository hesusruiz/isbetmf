package config

import "strings"

// Public TMF resources are accessible by all users, even unauthenticated ones
func IsPublicResource(resourceName string) bool {
	resourceName = strings.ToLower(resourceName)
	_, ok := PublicResources[resourceName]
	return ok
}

// The public objects are the following:
var PublicResources = map[string]bool{
	// TMF620 Product Catalog Management
	"category":             true,
	"catalog":              true,
	"productoffering":      true,
	"productofferingprice": true,
	"productspecification": true,

	// TMF633 Service Catalog Management
	"servicecatalog":       true,
	"servicecategory":      true,
	"servicecandidate":     true,
	"servicespecification": true,

	// TMF634 Resource Catalog Management
	"resourcecatalog":       true,
	"resourcecategory":      true,
	"resourcecandidate":     true,
	"resourcespecification": true,

	// Organization from TMF632 Party Management. But Individual is private.
	"organization": true,

	// TMF669 Party Role Management
	"partyrole": true,
}
