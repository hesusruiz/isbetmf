package reporting

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.BaseURL != "https://tmf.dome-marketplace-sbx.org" {
		t.Errorf("Expected BaseURL to be 'https://tmf.dome-marketplace-sbx.org', got %s", config.BaseURL)
	}

	if config.Timeout != 30 {
		t.Errorf("Expected Timeout to be 30, got %d", config.Timeout)
	}

	if len(config.ObjectTypes) == 0 {
		t.Error("Expected ObjectTypes to have default values")
	}

	if !config.PaginationEnabled {
		t.Error("Expected PaginationEnabled to be true")
	}

	if config.PageSize != 100 {
		t.Errorf("Expected PageSize to be 100, got %d", config.PageSize)
	}

	if config.MaxObjects != 10000 {
		t.Errorf("Expected MaxObjects to be 10000, got %d", config.MaxObjects)
	}

	if !config.ValidateRequiredFields {
		t.Error("Expected ValidateRequiredFields to be true")
	}

	if !config.ValidateRelatedParty {
		t.Error("Expected ValidateRelatedParty to be true")
	}
}

func TestConfigValidation(t *testing.T) {
	config := &Config{}

	// Test empty base URL
	if err := config.Validate(); err == nil {
		t.Error("Expected error for empty base URL")
	}

	// Test empty object types
	config.BaseURL = "https://example.com"
	if err := config.Validate(); err == nil {
		t.Error("Expected error for empty object types")
	}

	// Test invalid timeout
	config.ObjectTypes = []string{"productOffering"}
	config.Timeout = -1
	if err := config.Validate(); err == nil {
		t.Error("Expected error for negative timeout")
	}

	// Test valid config
	config.Timeout = 30
	if err := config.Validate(); err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}
}

func TestDefaultObjectTypes(t *testing.T) {
	types := DefaultObjectTypes()

	expectedTypes := []string{
		"productOffering",
		"productSpecification",
		"productOfferingPrice",
		"category",
		"individual",
		"organization",
		"productCatalog",
		"customer",
		"product",
		"service",
	}

	if len(types) != len(expectedTypes) {
		t.Errorf("Expected %d object types, got %d", len(expectedTypes), len(types))
	}

	for _, expectedType := range expectedTypes {
		found := false
		for _, actualType := range types {
			if actualType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected object type %s not found", expectedType)
		}
	}
}

func TestTMFObject(t *testing.T) {
	// Test V5 object
	v5RelatedParty := []RelatedParty{
		{
			Role: "Seller",
			PartyOrPartyRole: PartyRefOrPartyRole{
				ID:   "seller-1",
				Href: "https://example.com/seller/1",
			},
		},
	}

	v5RelatedPartyJSON, _ := json.Marshal(v5RelatedParty)

	obj := TMFObject{
		ID:           "test-id",
		Href:         "https://example.com/test",
		LastUpdate:   "2024-01-15T10:30:00Z",
		Version:      "1.0.0",
		Type:         "ProductOffering",
		RelatedParty: v5RelatedPartyJSON,
	}

	if obj.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", obj.ID)
	}

	if len(obj.RelatedParty) == 0 {
		t.Error("Expected RelatedParty to have data")
	}
}

func TestValidationResult(t *testing.T) {
	result := ValidationResult{
		ObjectID:   "test-obj",
		ObjectType: "ProductOffering",
		Valid:      true,
		Timestamp:  time.Now(),
	}

	if result.ObjectID != "test-obj" {
		t.Errorf("Expected ObjectID to be 'test-obj', got %s", result.ObjectID)
	}

	if !result.Valid {
		t.Error("Expected Valid to be true")
	}
}

func TestStatistics(t *testing.T) {
	stats := &Statistics{
		TotalObjects:   100,
		ValidObjects:   90,
		InvalidObjects: 10,
		TotalErrors:    15,
		TotalWarnings:  25,
		StartTime:      time.Now(),
		EndTime:        time.Now(),
	}

	if stats.TotalObjects != 100 {
		t.Errorf("Expected TotalObjects to be 100, got %d", stats.TotalObjects)
	}

	if stats.ValidObjects != 90 {
		t.Errorf("Expected ValidObjects to be 90, got %d", stats.ValidObjects)
	}

	if stats.InvalidObjects != 10 {
		t.Errorf("Expected InvalidObjects to be 10, got %d", stats.InvalidObjects)
	}
}

func TestRequiredFields(t *testing.T) {
	// Test that required fields are defined for all default object types
	for _, objType := range DefaultObjectTypes() {
		if fields, exists := RequiredFields[objType]; !exists {
			t.Errorf("Required fields not defined for object type: %s", objType)
		} else if len(fields) == 0 {
			t.Errorf("Required fields list is empty for object type: %s", objType)
		}
	}
}

func TestRequiredRelatedPartyRoles(t *testing.T) {
	// Test that required related party roles are defined for all default object types
	for _, objType := range DefaultObjectTypes() {
		if roles, exists := RequiredRelatedPartyRoles[objType]; !exists {
			t.Errorf("Required related party roles not defined for object type: %s", objType)
		} else if len(roles) == 0 {
			t.Errorf("Required related party roles list is empty for object type: %s", objType)
		}
	}
}

func TestValidatorWithV4AndV5(t *testing.T) {
	config := &Config{
		ValidateRequiredFields: true,
		ValidateRelatedParty:   true,
	}

	validator := NewValidator(config)

	// Test V5 object validation
	v5RelatedParty := []RelatedParty{
		{
			Role: "Seller",
			PartyOrPartyRole: PartyRefOrPartyRole{
				ID:   "seller-1",
				Href: "https://example.com/seller/1",
			},
		},
	}

	v5RelatedPartyJSON, _ := json.Marshal(v5RelatedParty)

	v5Obj := TMFObject{
		ID:           "v5-test-id",
		Href:         "https://example.com/v5-test",
		LastUpdate:   "2024-01-15T10:30:00Z",
		Version:      "1.0.0",
		RelatedParty: v5RelatedPartyJSON,
	}

	// Test with V5 config
	config.Version = VersionV5
	result := validator.ValidateObject(v5Obj, "productOffering")

	if !result.Valid {
		t.Errorf("Expected V5 object to be valid, got errors: %v", result.Errors)
	}

	// Test V4 object validation
	v4RelatedParty := []RelatedPartyV4{
		{
			ID:           "seller-1",
			Href:         "https://example.com/seller/1",
			Role:         "seller",
			Name:         "Test Seller",
			ReferredType: "Organization",
		},
	}

	v4RelatedPartyJSON, _ := json.Marshal(v4RelatedParty)

	v4Obj := TMFObject{
		ID:           "v4-test-id",
		Href:         "https://example.com/v4-test",
		LastUpdate:   "2024-01-15T10:30:00Z",
		Version:      "1.0.0",
		RelatedParty: v4RelatedPartyJSON,
	}

	// Test with V4 config
	config.Version = VersionV4
	result = validator.ValidateObject(v4Obj, "productOffering")

	if !result.Valid {
		t.Errorf("Expected V4 object to be valid, got errors: %v", result.Errors)
	}
}

func TestConfigVersionConstants(t *testing.T) {
	if VersionV4 != "v4" {
		t.Errorf("Expected VersionV4 to be 'v4', got %s", VersionV4)
	}

	if VersionV5 != "v5" {
		t.Errorf("Expected VersionV5 to be 'v5', got %s", VersionV5)
	}
}
