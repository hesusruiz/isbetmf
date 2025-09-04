package proxy

import (
	"encoding/json"
	"time"
)

// TMFObject represents a generic TMForum object
type TMFObject struct {
	ID           string          `json:"id"`
	Href         string          `json:"href"`
	LastUpdate   string          `json:"lastUpdate"`
	Version      string          `json:"version"`
	Type         string          `json:"@type"`
	RelatedParty json.RawMessage `json:"relatedParty,omitempty"`
	// Additional fields stored as generic map
	AdditionalFields map[string]any `json:"-"`
}

// RelatedParty represents a related party reference
type RelatedParty struct {
	Role             string              `json:"role"`
	PartyOrPartyRole PartyRefOrPartyRole `json:"partyOrPartyRole"`
	Type             string              `json:"@type"`
}

type RelatedPartyV4 struct {
	ID           string `json:"id"`
	Href         string `json:"href"`
	Role         string `json:"role"`
	Name         string `json:"name"`
	ReferredType string `json:"@referredType"`
}

// PartyRefOrPartyRole represents a party or party role reference
type PartyRefOrPartyRole struct {
	ID   string `json:"id"`
	Href string `json:"href"`
	Name string `json:"name,omitempty"`
	Type string `json:"@type"`
}

// ValidationResult represents the result of validating a single object
type ValidationResult struct {
	ObjectID   string              `json:"object_id"`
	ObjectType string              `json:"object_type"`
	Valid      bool                `json:"valid"`
	Errors     []ValidationError   `json:"errors,omitempty"`
	Warnings   []ValidationWarning `json:"warnings,omitempty"`
	Timestamp  time.Time           `json:"timestamp"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Statistics holds the overall statistics of the validation process
type Statistics struct {
	TotalObjects   int                  `json:"total_objects"`
	ValidObjects   int                  `json:"valid_objects"`
	InvalidObjects int                  `json:"invalid_objects"`
	TotalErrors    int                  `json:"total_errors"`
	TotalWarnings  int                  `json:"total_warnings"`
	ObjectsByType  map[string]TypeStats `json:"objects_by_type"`
	ErrorsByType   map[string]int       `json:"errors_by_type"`
	WarningsByType map[string]int       `json:"warnings_by_type"`
	StartTime      time.Time            `json:"start_time"`
	EndTime        time.Time            `json:"end_time"`
	Duration       time.Duration        `json:"duration"`
}

// TypeStats holds statistics for a specific object type
type TypeStats struct {
	Count    int `json:"count"`
	Valid    int `json:"valid"`
	Invalid  int `json:"invalid"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	Config      *Config            `json:"config"`
	Statistics  *Statistics        `json:"statistics"`
	Results     []ValidationResult `json:"results"`
	GeneratedAt time.Time          `json:"generated_at"`
}

var RequiredFieldsForAllObjects = []string{
	"id", "href", "lastUpdate", "version",
}

var DoNotRequireRelatedParties = []string{
	"category",
	"individual",
	"organization",
}

var DoNotRequireBuyerInfo = []string{
	"catalog",
	"productOffering",
	"productSpecification",
	"productOfferingPrice",
	"resourceSpecification",
	"serviceSpecification",
}

// RequiredFields defines the required fields for each object type
var RequiredFields = map[string][]string{
	"productOffering": {
		"id", "href", "lastUpdate", "version",
	},
	"productSpecification": {
		"id", "href", "lastUpdate", "version",
	},
	"productOfferingPrice": {
		"id", "href", "lastUpdate", "version",
	},
	"category": {
		"id", "href", "lastUpdate", "version",
	},
	"individual": {
		"id", "href", "lastUpdate", "version",
	},
	"organization": {
		"id", "href", "lastUpdate", "version",
	},
	"productCatalog": {
		"id", "href", "lastUpdate", "version",
	},
	"customer": {
		"id", "href", "lastUpdate", "version",
	},
	"product": {
		"id", "href", "lastUpdate", "version",
	},
	"service": {
		"id", "href", "lastUpdate", "version",
	},
}

// RequiredRelatedPartyRoles defines the required related party roles for each object type
var RequiredRelatedPartyRoles = map[string][]string{
	"productOffering": {
		"Seller", "SellerOperator",
	},
	"productSpecification": {
		"Seller", "SellerOperator",
	},
	"productOfferingPrice": {
		"Seller", "SellerOperator",
	},
	"category": {
		"Seller", "SellerOperator",
	},
	"individual": {
		"Seller", "SellerOperator",
	},
	"organization": {
		"Seller", "SellerOperator",
	},
	"productCatalog": {
		"Seller", "SellerOperator",
	},
	"customer": {
		"Seller", "SellerOperator",
	},
	"product": {
		"Seller", "SellerOperator",
	},
	"service": {
		"Seller", "SellerOperator",
	},
}
