package reporting

import (
	"testing"
)

func TestGetObjectsURLBuilding(t *testing.T) {
	config := &Config{
		BaseURL: "https://tmf.example.com",
		Timeout: 30,
	}

	// Test cases for different object types
	testCases := []struct {
		objectType  string
		expectedURL string
		shouldError bool
	}{
		{
			objectType:  "productOffering",
			expectedURL: "https://tmf.example.com/tmf-api/productCatalogManagement/v4/productOffering",
			shouldError: false,
		},
		{
			objectType:  "productSpecification",
			expectedURL: "https://tmf.example.com/tmf-api/productCatalogManagement/v4/productSpecification",
			shouldError: false,
		},
		{
			objectType:  "category",
			expectedURL: "https://tmf.example.com/tmf-api/productCatalogManagement/v4/category",
			shouldError: false,
		},
		{
			objectType:  "individual",
			expectedURL: "https://tmf.example.com/tmf-api/party/v4/individual",
			shouldError: false,
		},
		{
			objectType:  "organization",
			expectedURL: "https://tmf.example.com/tmf-api/party/v4/organization",
			shouldError: false,
		},
		{
			objectType:  "unknownType",
			expectedURL: "",
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.objectType, func(t *testing.T) {
			// We can't actually call GetObjects without a real server,
			// but we can verify that the routes map contains the expected paths
			if !tc.shouldError {
				pathPrefix, exists := GeneratedDefaultResourceToPathPrefixV4[tc.objectType]
				if !exists {
					t.Errorf("Expected object type %s to exist in routes map", tc.objectType)
				}

				expectedPath := tc.expectedURL[len(config.BaseURL):]
				if pathPrefix != expectedPath {
					t.Errorf("Expected path prefix %s, got %s", expectedPath, pathPrefix)
				}
			} else {
				// Verify that unknown types don't exist in the routes map
				if _, exists := GeneratedDefaultResourceToPathPrefixV4[tc.objectType]; exists {
					t.Errorf("Expected object type %s to NOT exist in routes map", tc.objectType)
				}
			}
		})
	}
}

func TestRoutesMapCompleteness(t *testing.T) {
	// Test that all default object types have corresponding routes
	defaultTypes := DefaultObjectTypes()

	for _, objType := range defaultTypes {
		if _, exists := GeneratedDefaultResourceToPathPrefixV4[objType]; !exists {
			t.Errorf("Default object type %s missing from routes map", objType)
		}
	}

	// Test that all default object types have required fields defined
	for _, objType := range defaultTypes {
		if _, exists := RequiredFields[objType]; !exists {
			t.Errorf("Default object type %s missing from RequiredFields map", objType)
		}
	}

	// Test that all default object types have related party roles defined
	for _, objType := range defaultTypes {
		if _, exists := RequiredRelatedPartyRoles[objType]; !exists {
			t.Errorf("Default object type %s missing from RequiredRelatedPartyRoles map", objType)
		}
	}
}
