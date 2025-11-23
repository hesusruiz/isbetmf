package repository

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/jpath"
	"github.com/hesusruiz/isbetmf/types"
	"golang.org/x/exp/slog"
)

// TMFObjectMap represents a TMForum object as a map with utility methods
// This is designed to make fixing validation errors simple and efficient
type TMFObjectMap map[string]any

// NewTMFObject creates a new TMFObject from a JSON byte slice
func NewTMFObjectMap(data []byte) (TMFObjectMap, error) {
	var obj TMFObjectMap
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TMF object: %w", err)
	}
	return obj, nil
}

// NewTMFObjectMapFromRequest creates a new TMFObject from a JSON byte slice.
// It is intended to be used with data received from a remote TMF server.
func NewTMFObjectMapFromRequest(resourceName string, data []byte) (TMFObjectMap, error) {
	var obj TMFObjectMap
	err := json.Unmarshal(data, &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal TMF object: %w", errl.Error(err))
	}

	// Check that the object matches the resourceName
	objType, _ := obj["@type"].(string)
	if objType != "" {
		// If the object type is specified, it must match the resourceName passed by the caller
		if !strings.EqualFold(resourceName, objType) {
			return nil, errl.Errorf("invalid object type, expected %s but received %s", resourceName, objType)
		}

	} else {
		// No object type in the object, set it to the resourceName
		obj["@type"] = resourceName
	}

	return obj, nil

}

func NewTMFObjectMapFromUpstream(resourceName string, data map[string]any) (TMFObjectMap, ValidationResult) {
	obj := TMFObjectMap(maps.Clone(data))
	validations := obj.Validate(resourceName)

	return obj, validations
}

// NewTMFObjectFromMap creates a new TMFObject from an existing map
func NewTMFObjectFromMap(data map[string]any) TMFObjectMap {
	obj := TMFObjectMap(maps.Clone(data))
	return obj
}

func (obj TMFObjectMap) Validate(resourceName string) ValidationResult {
	result := ValidationResult{
		ObjectID:   obj.ID(),
		ObjectType: resourceName,
		Valid:      true,
		Timestamp:  time.Now(),
	}

	// Validate required fields
	obj.validateRequiredFields(resourceName, &result)

	// Validate related party requirements
	obj.validateRelatedParty(&result)

	// Determine overall validity (object is valid if it has zero validation errors)
	result.Valid = len(result.Errors) == 0

	return result
}

// validateRequiredFields checks if all required fields are present and optionally fixes them
func (obj TMFObjectMap) validateRequiredFields(resourceName string, result *ValidationResult) {

	// Special processiong for the object type: check that the object matches the resourceName and fix the object if needed.
	objType := obj.GetType()
	if objType != "" {
		// If the object type is specified, it must match the resourceName passed by the caller
		field := "@type"
		if !strings.EqualFold(resourceName, objType) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Object type field '%s' does not match resource type '%s'", objType, resourceName),
				Code:    "MISSING_INVALID_VALUE",
			})
		}

	} else {
		// No object type in the object, set it to the resourceName
		field := "@type"
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   field,
			Message: fmt.Sprintf("Recommended field '%s' is missing, setting it to %s", field, resourceName),
			Code:    "MISSING_RECOMMENDED_FIELD",
		})
		obj.SetType(resourceName)
	}

	// This checks the fields that are required for all objects
	for _, field := range RequiredFieldsForAllObjects {
		if !obj.HasField(field) {
			// TODO: Implement fixing logic for missing required fields
			// If v.config.FixValidationErrors is true, attempt to fix the missing field
			// and move the error to result.ErrorsFixed if successfully fixed

			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Required field '%s' is missing", field),
				Code:    "MISSING_REQUIRED_FIELD",
			})
		}
	}

	for _, field := range RecommendedFieldsForAllObjects {
		if !obj.HasField(field) {
			// TODO: Implement fixing logic for missing required fields
			// If v.config.FixValidationErrors is true, attempt to fix the missing field
			// and move the error to result.ErrorsFixed if successfully fixed

			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   field,
				Message: fmt.Sprintf("Recommended field '%s' is missing", field),
				Code:    "MISSING_RECOMMENDED_FIELD",
			})
		}
	}

}

func (obj TMFObjectMap) validateRelatedParty(result *ValidationResult) {

	// We just return if the object does not require Seller nor Buyer info
	if !obj.RequiresSellerInfo() {
		return
	}

	// Check that the object has a relatedParty object
	relatedParties := jpath.GetList(obj, "relatedParty")
	if len(relatedParties) == 0 {
		// msg := "Missing " + userRole + " and " + userOperatorRole + " fields"
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' object",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})

		return
	}

	var sellerDid string
	var sellerOperatorDid string
	var buyerDid string
	var buyerOperatorDid string

	sellerRole := strings.ToLower("Seller")
	sellerOperatorRole := strings.ToLower("SellerOperator")
	buyerRole := strings.ToLower("Buyer")
	buyerOperatorRole := strings.ToLower("BuyerOperator")

	for _, rp := range relatedParties {
		// Cast the entry to a map[string]any
		rpMap, _ := rp.(map[string]any)
		if len(rpMap) == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "relatedParty",
				Message: "Invalid 'relatedParty' object",
				Code:    "INVALID_RELATED_PARTY_INFO",
			})
			return
		}

		// Extract the "role" field in lowercase for case-insentitive comparison
		rpRole, _ := rpMap["role"].(string)
		rpRole = strings.ToLower(rpRole)

		// If the role of the entry is not one that we are looking for, continue the loop
		if rpRole != sellerRole && rpRole != sellerOperatorRole && rpRole != buyerRole && rpRole != buyerOperatorRole {
			continue
		}

		switch rpRole {
		case sellerRole:
			sellerDid, _ = rpMap["name"].(string)
			if sellerDid == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "relatedParty",
					Message: "Missing or invalid 'name' field in 'relatedParty' object for 'Seller' role",
					Code:    "MISSING_RELATED_PARTY_INFO",
				})
			}
		case sellerOperatorRole:
			sellerOperatorDid, _ = rpMap["name"].(string)
			if sellerOperatorDid == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "relatedParty",
					Message: "Missing or invalid 'name' field in 'relatedParty' object for 'SellerOperator' role",
					Code:    "MISSING_RELATED_PARTY_INFO",
				})
			}
		case buyerRole:
			buyerDid, _ = rpMap["name"].(string)
			if buyerDid == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "relatedParty",
					Message: "Missing or invalid 'name' field in 'relatedParty' object for 'Buyer' role",
					Code:    "MISSING_RELATED_PARTY_INFO",
				})
			}
		case buyerOperatorRole:
			buyerOperatorDid, _ = rpMap["name"].(string)
			if buyerOperatorDid == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "relatedParty",
					Message: "Missing or invalid 'name' field in 'relatedParty' object for 'BuyerOperator' role",
					Code:    "MISSING_RELATED_PARTY_INFO",
				})
			}
		}

	}

	// Set the error depending on what we have found for Seller and SellerOperator
	if sellerDid == "" && sellerOperatorDid == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' info for 'Seller' and 'SellerOperator' role",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})
	} else if sellerDid == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' info for 'Seller' role",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})
	} else if sellerOperatorDid == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' info for 'SellerOperator' role",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})
	}

	if !obj.RequiresBuyerInfo() {
		return
	}

	// Set the error depending on what we have found for Buyer and BuyerOperator
	if buyerDid == "" && buyerOperatorDid == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' info for 'Buyer' and 'BuyerOperator' role",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})
	} else if buyerDid == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' info for 'Buyer' role",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})
	} else if buyerOperatorDid == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "relatedParty",
			Message: "Missing 'relatedParty' info for 'BuyerOperator' role",
			Code:    "MISSING_RELATED_PARTY_INFO",
		})
	}

}

func (obj TMFObjectMap) ToTMFObject(resourceName string) *TMFRecord {

	id := obj.ID()
	objectType := obj.GetType()
	if objectType == "" {
		objectType = resourceName
	}
	version := obj.Version()
	// TODO: support for v5 API
	apiVersion := "v4"
	lastUpdate := obj.LastUpdate()
	content, _ := obj.ToJSON()

	seller, _, _ := obj.GetSellerInfo("v4")
	buyer, _, _ := obj.GetBuyerInfo("v4")

	now := time.Now().Unix()

	o := &TMFRecord{
		ID:         id,
		Type:       objectType,
		Version:    version,
		APIVersion: apiVersion,
		Seller:     seller,
		Buyer:      buyer,
		LastUpdate: lastUpdate,
		Content:    content,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return o
}

// ToJSON converts the TMFObject to JSON bytes
func (obj TMFObjectMap) ToJSON() ([]byte, error) {
	return json.Marshal(obj)
}

// ToMap converts the TMFObject to a regular map[string]any
func (obj TMFObjectMap) ToMap() map[string]any {
	return maps.Clone(obj)
}

// Utility methods for well-known top-level attributes

// ID returns the object ID
func (obj TMFObjectMap) ID() string {
	if id, ok := obj["id"].(string); ok {
		return id
	}
	return ""
}

func (obj TMFObjectMap) IsIndividual() bool {
	return strings.EqualFold(obj.GetType(), "Individual")
}

func (obj TMFObjectMap) IsOrganization() bool {
	return strings.EqualFold(obj.GetType(), "Organization")
}

// SetID sets the object ID
func (obj TMFObjectMap) SetID(id string) {
	obj["id"] = id
}

// Href returns the object href
func (obj TMFObjectMap) Href() string {
	if href, ok := obj["href"].(string); ok {
		return href
	}
	return ""
}

// SetHref sets the object href
func (obj TMFObjectMap) SetHref(href string) {
	obj["href"] = href
}

// Version returns the object version
func (obj TMFObjectMap) Version() string {
	if version, ok := obj["version"].(string); ok {
		return version
	}
	return ""
}

// SetVersion sets the object version
func (obj TMFObjectMap) SetVersion(version string) {
	obj["version"] = version
}

// LastUpdate returns the object lastUpdate
func (obj TMFObjectMap) LastUpdate() string {
	if lastUpdate, ok := obj["lastUpdate"].(string); ok {
		return lastUpdate
	}
	return ""
}

// SetLastUpdate sets the object lastUpdate
func (obj TMFObjectMap) SetLastUpdate(lastUpdate string) {
	obj["lastUpdate"] = lastUpdate
}

// SetLastUpdateNow sets the object lastUpdate to current timestamp in RFC3339 format
func (obj TMFObjectMap) SetLastUpdateNow() {
	obj["lastUpdate"] = time.Now().Format(time.RFC3339)
}

// GetType returns the object @type
func (obj TMFObjectMap) GetType() string {
	if objType, ok := obj["@type"].(string); ok {
		return objType
	}
	return ""
}

// SetType sets the object @type
func (obj TMFObjectMap) SetType(objType string) {
	obj["@type"] = objType
}

// LifecycleStatus returns the object lifecycleStatus
func (obj TMFObjectMap) LifecycleStatus() string {
	if lifecycleStatus, ok := obj["lifecycleStatus"].(string); ok {
		return lifecycleStatus
	}
	return ""
}

// SetLifecycleStatus sets the object lifecycleStatus
func (obj TMFObjectMap) SetLifecycleStatus(lifecycleStatus string) {
	obj["lifecycleStatus"] = lifecycleStatus
}

// Name returns the object name
func (obj TMFObjectMap) Name() string {
	if name, ok := obj["name"].(string); ok {
		return name
	}
	return ""
}

// SetName sets the object name
func (obj TMFObjectMap) SetName(name string) {
	obj["name"] = name
}

// Description returns the object description
func (obj TMFObjectMap) Description() string {
	if description, ok := obj["description"].(string); ok {
		return description
	}
	return ""
}

// SetDescription sets the object description
func (obj TMFObjectMap) SetDescription(description string) {
	obj["description"] = description
}

// RelatedParty methods

// RelatedParty returns the relatedParty array
func (obj TMFObjectMap) RelatedParty() []map[string]any {
	if relatedParty, ok := obj["relatedParty"].([]any); ok {
		result := make([]map[string]any, 0, len(relatedParty))
		for _, item := range relatedParty {
			if rp, ok := item.(map[string]any); ok {
				result = append(result, rp)
			}
		}
		return result
	}
	return nil
}

// SetRelatedParty sets the relatedParty array
func (obj TMFObjectMap) SetRelatedParty(relatedParty []map[string]any) {
	// Convert []map[string]any to []any for JSON serialization
	items := make([]any, len(relatedParty))
	for i, rp := range relatedParty {
		items[i] = rp
	}
	obj["relatedParty"] = items
}

// AddRelatedParty adds a related party entry
func (obj TMFObjectMap) AddRelatedParty(relatedParty map[string]any) {
	current := obj.RelatedParty()
	if current == nil {
		current = make([]map[string]any, 0)
	}
	current = append(current, relatedParty)
	obj.SetRelatedParty(current)
}

// HasRelatedParty returns true if the object has related party information
func (obj TMFObjectMap) HasRelatedParty() bool {
	relatedParty := obj.RelatedParty()
	return len(relatedParty) > 0
}

// Validation helper methods

// HasField checks if a specific field exists and is not empty
func (obj TMFObjectMap) HasField(field string) bool {
	value, exists := obj[field]
	if !exists {
		return false
	}

	switch v := value.(type) {
	case string:
		return v != ""
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	default:
		return true // Other types are considered present if they exist
	}
}

// SetField sets a field value
func (obj TMFObjectMap) SetField(field string, value any) {
	obj[field] = value
}

// GetField returns a field value
func (obj TMFObjectMap) GetField(field string) any {
	return obj[field]
}

// GetStringField returns a field value as string
func (obj TMFObjectMap) GetStringField(field string) string {
	if value, ok := obj[field].(string); ok {
		return value
	}
	return ""
}

// SetStringField sets a field value as string
func (obj TMFObjectMap) SetStringField(field string, value string) {
	obj[field] = value
}

// GetArrayField returns a field value as []any
func (obj TMFObjectMap) GetArrayField(field string) []any {
	if value, ok := obj[field].([]any); ok {
		return value
	}
	return nil
}

// SetArrayField sets a field value as []any
func (obj TMFObjectMap) SetArrayField(field string, value []any) {
	obj[field] = value
}

// GetMapField returns a field value as map[string]any
func (obj TMFObjectMap) GetMapField(field string) map[string]any {
	if value, ok := obj[field].(map[string]any); ok {
		return value
	}
	return nil
}

// SetMapField sets a field value as map[string]any
func (obj TMFObjectMap) SetMapField(field string, value map[string]any) {
	obj[field] = value
}

// Clone creates a deep copy of the TMFObjectMap
func (obj TMFObjectMap) Clone() TMFObjectMap {
	// Use JSON marshaling/unmarshaling for deep copy
	data, err := obj.ToJSON()
	if err != nil {
		// If JSON fails, create a shallow copy
		return NewTMFObjectFromMap(obj.ToMap())
	}

	clone, err := NewTMFObjectMap(data)
	if err != nil {
		// If unmarshaling fails, create a shallow copy
		return NewTMFObjectFromMap(obj.ToMap())
	}

	return clone
}

// IsEmpty returns true if the object is empty
func (obj TMFObjectMap) IsEmpty() bool {
	return len(obj) == 0
}

// Keys returns all the keys in the object
func (obj TMFObjectMap) Keys() []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	return keys
}

// DeleteField removes a field from the object
func (obj TMFObjectMap) DeleteField(field string) {
	delete(obj, field)
}

// String returns a string representation of the object (for debugging)
func (obj TMFObjectMap) String() string {
	data, err := obj.ToJSON()
	if err != nil {
		return fmt.Sprintf("TMFObject{error: %v}", err)
	}
	return string(data)
}

func (obj TMFObjectMap) IsOwner(caller types.AuthUser, serverOperatorDid string) (isOwner bool, reason string) {

	// Ownership of an object depends on the type of object
	objType, _ := obj["@type"].(string)
	objType = strings.ToLower(objType)

	switch objType {
	case "organization":

		// If the caller is us (the server operator), then we can read/write/update/delete
		if SameOrganizations(caller.OrganizationIdentifier, serverOperatorDid) {
			return true, fmt.Sprintf("caller %s is server operator %s", caller.OrganizationIdentifier, serverOperatorDid)
		}

		// If the organization of the caller and object are the same, then the caller can read/write/update/delete
		objectOrganizationId := jpath.GetString(obj, "organizationIdentification.*.identificationId")
		if SameOrganizations(objectOrganizationId, caller.OrganizationIdentifier) {
			return true, fmt.Sprintf("caller %s is same as in object %s", caller.OrganizationIdentifier, objectOrganizationId)
		}

		return false, fmt.Sprintf("caller (%s) is neither the same as in object (%s) or the server operator", caller.OrganizationIdentifier, objectOrganizationId)

	case "individual":

		// TODO: revise this policy to be restrictive

		// If the caller is us (the server operator), then we can read/write/update/delete
		if SameOrganizations(caller.OrganizationIdentifier, serverOperatorDid) {
			return true, fmt.Sprintf("caller %s is server operator %s", caller.OrganizationIdentifier, serverOperatorDid)
		}

		// If the caller is the Organization that is the mandator is the LEARCredential of the employee
		// then the caller can read/write/update/delete
		individualIdentificationArray := jpath.GetList(obj, "individualIdentification")

		// Look for an entry with 'identificationType=learcredentialemployee'
		for _, individualIdentification := range individualIdentificationArray {
			individualIdentificationMap, _ := individualIdentification.(map[string]any)
			if individualIdentificationMap["identificationType"] == "learcredentialemployee" {
				// The 'issuingAuthority' must be equal to the caller organizationIdentifier
				issuingAuthority := individualIdentificationMap["issuingAuthority"].(string)
				if SameOrganizations(issuingAuthority, caller.OrganizationIdentifier) {
					return true, fmt.Sprintf("caller %s is same as mandator in Individual object %s", caller.OrganizationIdentifier, issuingAuthority)
				} else {
					return false, fmt.Sprintf("caller (%s) is neither the mandator in Individual object (%s) or the server operator", caller.OrganizationIdentifier, issuingAuthority)
				}
			}
		}

		return false, fmt.Sprintf("caller (%s) is neither the mandator in Individual object or the server operator", caller.OrganizationIdentifier)

	case "category":

		// If the caller is us (the server operator), then we can read/write/update/delete
		if SameOrganizations(caller.OrganizationIdentifier, serverOperatorDid) {
			return true, fmt.Sprintf("caller %s is server operator %s", caller.OrganizationIdentifier, serverOperatorDid)
		}

		return false, fmt.Sprintf("caller %s is not the server operator %s", caller.OrganizationIdentifier, serverOperatorDid)

	default:

		// If the caller is us (the server operator), then we can read/write/update/delete
		if SameOrganizations(caller.OrganizationIdentifier, serverOperatorDid) {
			return true, fmt.Sprintf("caller %s is server operator %s", caller.OrganizationIdentifier, serverOperatorDid)
		}

		// For any other objects, we require that the object includes the Seller info, and then
		// the user must be either the server operator or the seller

		// Try to retrieve the Seller info
		objSellerDid, objSellerOperatorDid, err := obj.GetSellerInfo("v4")
		if err != nil {
			return false, fmt.Sprintf("object (%s) does not contain seller information", obj.ID())
		}

		// If the caller is the same as the object SellerOperator or the Seller, then is the owner
		if SameOrganizations(caller.OrganizationIdentifier, objSellerDid) || SameOrganizations(caller.OrganizationIdentifier, objSellerOperatorDid) {
			return true, fmt.Sprintf("caller %s is seller %s or seller operator %s", caller.OrganizationIdentifier, objSellerDid, objSellerOperatorDid)
		}

		// Try to retrieve the Buyer info, which may not exist.
		// We already checked for Seller info, so if Buyer info does not exist, caller is not the owner
		objBuyerDid, objBuyerOperatorDid, err := obj.GetBuyerInfo("v4")
		if err != nil {
			return false, fmt.Sprintf("object (%s) does not contain buyer information and caller (%s) is not the seller or seller operator", obj.ID(), caller.OrganizationIdentifier)
		}

		// If the caller is the same as the object BuyerOperator or the Buyer, then is the owner
		if SameOrganizations(caller.OrganizationIdentifier, objBuyerDid) || SameOrganizations(caller.OrganizationIdentifier, objBuyerOperatorDid) {
			return true, fmt.Sprintf("caller %s is buyer %s or buyer operator %s", caller.OrganizationIdentifier, objBuyerDid, objBuyerOperatorDid)
		}

		return false, fmt.Sprintf("caller %s is not seller or buyer in object %s", caller.OrganizationIdentifier, obj.ID())

	}

}

// setSellerAndBuyerInfo adds the required fields to the incoming object argument
// It calls the appropriate version-specific function based on the API version
func (obj TMFObjectMap) SetSellerInfo(serverOperatorDid string, organizationIdentifier string, apiVersion string) (err error) {
	// We do nothing for Individual or Organization objects, which are special and do not have Seller info
	objType := obj.GetType()
	objType = strings.ToLower(objType)
	if objType == "individual" || objType == "organization" {
		return nil
	}

	switch apiVersion {
	case "v4":
		return setSellerInfoV4(obj, serverOperatorDid, organizationIdentifier)
	case "v5":
		return setSellerInfoV5(obj, serverOperatorDid, organizationIdentifier)
	default:
		// Default to V5 for backward compatibility
		return setSellerInfoV4(obj, serverOperatorDid, organizationIdentifier)
	}
}

// setSellerInfoV4 adds the required fields to the incoming object argument for V4 API
// Specifically, the Seller and SellerOperator roles are added to the relatedParty list
func setSellerInfoV4(tmfObjectMap map[string]any, serverOperatorDid string, organizationIdentifier string) (err error) {

	// Normalize the organization identifier to the DID format
	if !strings.HasPrefix(organizationIdentifier, "did:elsi:") {
		organizationIdentifier = "did:elsi:" + organizationIdentifier
	}

	// Look for the "Seller", "SellerOperator", "Buyer" and "BuyerOperator" roles
	relatedParties := jpath.GetList(tmfObjectMap, "relatedParty")

	// Build the two entries for V4 format
	sellerEntry := map[string]any{
		"role":          "Seller",
		"id":            "urn:ngsi-ld:organization:" + organizationIdentifier,
		"href":          "urn:ngsi-ld:organization:" + organizationIdentifier,
		"name":          organizationIdentifier,
		"@referredType": "Organization",
	}
	sellerOperator := map[string]any{
		"role":          "SellerOperator",
		"id":            "urn:ngsi-ld:organization:" + serverOperatorDid,
		"href":          "urn:ngsi-ld:organization:" + serverOperatorDid,
		"name":          serverOperatorDid,
		"@referredType": "Organization",
	}

	// If the object does not have a 'relatedParty' array, create one with Seller=caller and SelletOperator=server_operator
	if len(relatedParties) == 0 {
		slog.Debug("setSellerAndBuyerInfoV4: no relatedParty, adding seller and sellerOperator")
		tmfObjectMap["relatedParty"] = []any{sellerEntry, sellerOperator}
		return nil
	}

	// We now search for Seller and SellerOperator entries and overwrite them or add them.
	// The relatedParty array may contain entries which are not these, and we should not touch them.

	foundSeller := false
	foundSellerOperator := false

	newRelatedParties := []any{}

	for _, rp := range relatedParties {

		// Convert entry to a map
		rpMap, _ := rp.(map[string]any)
		if len(rpMap) == 0 {
			return errl.Errorf("invalid relatedParty entry")
		}

		rpRole, _ := rpMap["role"].(string)
		rpRole = strings.ToLower(rpRole)

		if rpRole != "seller" && rpRole != "selleroperator" {
			newRelatedParties = append(newRelatedParties, rp)
			// Go to next entry
			continue
		}

		if rpRole == "seller" {
			// Overwrite the entry, because we can not allow the user to create fake info
			newRelatedParties = append(newRelatedParties, sellerEntry)
			foundSeller = true
			continue
		}
		if rpRole == "selleroperator" {
			// Overwrite the entry, because we can not allow the user to create fake info
			newRelatedParties = append(newRelatedParties, sellerOperator)
			foundSellerOperator = true
			continue
		}

	}

	if !foundSeller {
		// Add the seller if it is not already in the list
		slog.Debug("setSellerAndBuyerInfoV4: adding seller", "organizationIdentifier", organizationIdentifier)
		newRelatedParties = append(newRelatedParties, sellerEntry)
	}

	if !foundSellerOperator {
		// Add the seller operator if it is not already in the list
		slog.Debug("setSellerAndBuyerInfoV4: adding seller operator", "organizationIdentifier", organizationIdentifier)
		newRelatedParties = append(newRelatedParties, sellerOperator)
	}

	tmfObjectMap["relatedParty"] = newRelatedParties

	return nil

}

// setSellerInfoV5 adds the required fields to the incoming object argument
// Specifically, the Seller and SellerOperator roles are added to the relatedParty list
func setSellerInfoV5(tmfObjectMap map[string]any, serverOperatorDid string, organizationIdentifier string) (err error) {

	// Normalize all organization identifiers to the DID format
	if !strings.HasPrefix(organizationIdentifier, "did:elsi:") {
		organizationIdentifier = "did:elsi:" + organizationIdentifier
	}

	// Look for the "Seller", "SellerOperator", "Buyer" and "BuyerOperator" roles
	relatedParties := jpath.GetList(tmfObjectMap, "relatedParty")

	// Build the two entries
	sellerEntry := map[string]any{
		"role":  "Seller",
		"@type": "RelatedPartyRefOrPartyRoleRef",
		"partyOrPartyRole": map[string]any{
			"@type":         "PartyRef",
			"href":          "urn:ngsi-ld:organization:" + organizationIdentifier,
			"id":            "urn:ngsi-ld:organization:" + organizationIdentifier,
			"name":          organizationIdentifier,
			"@referredType": "Organization",
		},
	}
	sellerOperator := map[string]any{
		"role":  "SellerOperator",
		"@type": "RelatedPartyRefOrPartyRoleRef",
		"partyOrPartyRole": map[string]any{
			"@type":         "PartyRef",
			"href":          "urn:ngsi-ld:organization:" + serverOperatorDid,
			"id":            "urn:ngsi-ld:organization:" + serverOperatorDid,
			"name":          serverOperatorDid,
			"@referredType": "Organization",
		},
	}

	if len(relatedParties) == 0 {
		slog.Debug("setSellerAndBuyerInfo: no relatedParty, adding seller and sellerOperator")
		tmfObjectMap["relatedParty"] = []any{sellerEntry, sellerOperator}
		return nil
	}

	foundSeller := false
	foundSellerOperator := false

	newRelatedParties := []any{}

	for _, rp := range relatedParties {

		// Convert entry to a map
		rpMap, _ := rp.(map[string]any)
		if len(rpMap) == 0 {
			return errl.Errorf("invalid relatedParty entry")
		}

		rpRole, _ := rpMap["role"].(string)
		rpRole = strings.ToLower(rpRole)

		if rpRole != "seller" && rpRole != "selleroperator" {
			newRelatedParties = append(newRelatedParties, rp)
			// Go to next entry
			continue
		}

		if rpRole == "seller" {
			// Overwrite the entry, because we can not allow the user to create fake info
			newRelatedParties = append(newRelatedParties, sellerEntry)
			foundSeller = true
			continue
		}
		if rpRole == "selleroperator" {
			// Overwrite the entry, because we can not allow the user to create fake info
			newRelatedParties = append(newRelatedParties, sellerOperator)
			foundSellerOperator = true
			continue
		}

	}

	if !foundSeller {
		// Add the seller if it is not already in the list
		slog.Debug("setSellerAndBuyerInfo: adding seller", "organizationIdentifier", organizationIdentifier)
		newRelatedParties = append(newRelatedParties, sellerEntry)
	}

	if !foundSellerOperator {
		// Add the seller operator if it is not already in the list
		slog.Debug("setSellerAndBuyerInfo: adding seller operator", "organizationIdentifier", organizationIdentifier)
		newRelatedParties = append(newRelatedParties, sellerOperator)
	}

	tmfObjectMap["relatedParty"] = newRelatedParties

	return nil

}

func (obj TMFObjectMap) RequiresSellerInfo() bool {
	objType := obj.GetType()
	objType = strings.ToLower(objType)
	return !slices.Contains(DoNotRequireRelatedParties, objType)
}

func (obj TMFObjectMap) RequiresBuyerInfo() bool {
	objType := obj.GetType()
	objType = strings.ToLower(objType)
	relp := slices.Contains(DoNotRequireRelatedParties, objType)
	buyp := slices.Contains(DoNotRequireBuyerInfo, objType)
	rr := relp || buyp
	return !rr
}

// GetSellerInfo finds the Seller and SellerOperator identifiers in the relatedParty array of the object.
// If some identifier is missing (or both), it returns an error. But it returns what it finds.
// So, even if the returned error is not nil, the caller may check the sellerDid and the sellerOperatorDid.
// This is useful if the caller has logic to handle cases where only one of the values is found.
func (obj TMFObjectMap) GetSellerInfo(apiVersion string) (sellerDid string, sellerOperatorDid string, err error) {
	if !obj.RequiresSellerInfo() {
		return
	}

	switch apiVersion {
	case "v4":
		return getUserAndUserOperatorInfoV4(obj, "Seller", "SellerOperator")
	case "v5":
		return getUserAndUserOperatorInfoV5(obj, "Seller", "SellerOperator")
	default:
		// Default to V4 for backward compatibility
		return getUserAndUserOperatorInfoV4(obj, "Seller", "SellerOperator")
	}
}

func (obj TMFObjectMap) GetBuyerInfo(apiVersion string) (sellerDid string, sellerOperatorDid string, err error) {
	if !obj.RequiresBuyerInfo() {
		return
	}

	switch apiVersion {
	case "v4":
		return getUserAndUserOperatorInfoV4(obj, "Buyer", "BuyerOperator")
	case "v5":
		return getUserAndUserOperatorInfoV5(obj, "Buyer", "BuyerOperator")
	default:
		// Default to V4 for backward compatibility
		return getUserAndUserOperatorInfoV4(obj, "Buyer", "BuyerOperator")
	}
}

// getUserAndUserOperatorInfoV4 finds the relatedparty entries with the specified roles.
// It returns what it finds. So, even if the returned error is not nil, the caller may check the sellerDid and the sellerOperatorDid.
// This is useful if the caller has logic to handle cases where only one of the values is found.
func getUserAndUserOperatorInfoV4(tmfObjectMap map[string]any, userRole string, userOperatorRole string) (sellerDid string, sellerOperatorDid string, err error) {
	// In V4, relatedParty is a list of maps with fields like "role", "id", "href", "name", "@referredType"
	// We need to extract the "name" for the "Seller" and "SellerOperator" roles

	// Convert roles to lowercase to facilitate case-insensitive comparison
	userRole = strings.ToLower(userRole)
	userOperatorRole = strings.ToLower(userOperatorRole)

	relatedParties := jpath.GetList(tmfObjectMap, "relatedParty")

	if len(relatedParties) == 0 {
		err = errl.Errorf("no relatedParty")
		return "", "", err
	}

	for _, rp := range relatedParties {
		// Cast the entry to a map[string]any
		rpMap, _ := rp.(map[string]any)
		if len(rpMap) == 0 {
			return "", "", errl.Errorf("invalid relatedParty entry")
		}

		// Extract the "role" field in lowercase for case-insentitive comparison
		rpRole, _ := rpMap["role"].(string)
		rpRole = strings.ToLower(rpRole)

		// If the role of the entry is not one that we are looking for, continue the loop
		if rpRole != userRole && rpRole != userOperatorRole {
			continue
		}

		if rpRole == userRole {
			sellerDid, _ = rpMap["name"].(string)
			continue
		}
		if rpRole == userOperatorRole {
			sellerOperatorDid, _ = rpMap["name"].(string)
			continue
		}
	}

	// Set the error depending on what we have found
	if sellerDid == "" && sellerOperatorDid == "" {
		err = errl.Errorf("no %s or %s", userRole, userOperatorRole)
	} else if sellerDid == "" {
		err = errl.Errorf("no %s", userRole)
	} else if sellerOperatorDid == "" {
		err = errl.Errorf("no %s", userOperatorRole)
	}

	// return what we have foud. So, even if the err is not nil, the caller may check the sellerDid and the sellerOperatorDid.
	// This is useful if the caller has logic to handle cases where only one of the values is found.
	return sellerDid, sellerOperatorDid, err

}

func getUserAndUserOperatorInfoV5(tmfObjectMap map[string]any, userRole string, userOperatorRole string) (sellerDid string, sellerOperatorDid string, err error) {

	userRole = strings.ToLower(userRole)
	userOperatorRole = strings.ToLower(userOperatorRole)

	// Look for the "Seller", "SellerOperator", "Buyer" and "BuyerOperator" roles
	relatedParties := jpath.GetList(tmfObjectMap, "relatedParty")

	if len(relatedParties) == 0 {
		err = errl.Errorf("no relatedParty")
		return
	}

	for _, rp := range relatedParties {

		// Convert entry to a map
		rpMap, _ := rp.(map[string]any)
		if len(rpMap) == 0 {
			return "", "", errl.Errorf("invalid relatedParty entry")
		}

		rpRole, _ := rpMap["role"].(string)
		rpRole = strings.ToLower(rpRole)

		if rpRole != userRole && rpRole != userOperatorRole {
			// Go to next entry
			continue
		}

		if rpRole == userRole {
			party, _ := rpMap["partyOrPartyRole"].(map[string]any)
			sellerDid, _ = party["name"].(string)
			continue
		}
		if rpRole == userOperatorRole {
			party, _ := rpMap["partyOrPartyRole"].(map[string]any)
			sellerOperatorDid, _ = party["name"].(string)
			continue
		}

	}

	if sellerDid == "" && sellerOperatorDid == "" {
		err = errl.Errorf("no %s or %s", userRole, userOperatorRole)
		return
	}

	if sellerDid == "" {
		err = errl.Errorf("no %s", userRole)
		return
	}
	if sellerOperatorDid == "" {
		err = errl.Errorf("no %s", userOperatorRole)
		return
	}

	return

}
