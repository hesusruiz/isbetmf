package repository

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
)

const eIDASAuthority = "eIDAS"
const elsiIdentificationType = "did:elsi"

type Organization struct {
	CommonName             string `json:"commonName"`
	Country                string `json:"country"`
	EmailAddress           string `json:"emailAddress"`
	Organization           string `json:"organization"`
	OrganizationIdentifier string `json:"organizationIdentifier"`
	SerialNumber           string `json:"serialNumber"`
}

func (u *Organization) ToMap() map[string]any {
	return map[string]any{
		"commonName":             u.CommonName,
		"country":                u.Country,
		"emailAddress":           u.EmailAddress,
		"organization":           u.Organization,
		"organizationIdentifier": u.OrganizationIdentifier,
		"serialNumber":           u.SerialNumber,
	}
}

func TMFOrganizationFromToken(accessToken map[string]any, user *Organization) (*TMFObject, error) {

	id := "urn:ngsi-ld:organization:" + user.OrganizationIdentifier
	did := user.OrganizationIdentifier
	if !strings.HasPrefix(did, "did:elsi:") {
		did = "did:elsi:" + did
	}

	now := time.Now()
	lastUpdate := now.Format(time.RFC3339Nano)

	objectType := config.Organization
	version := "1.0"

	theIdentification := map[string]any{
		"@type":              "organizationIdentification",
		"identificationId":   did,
		"identificationType": elsiIdentificationType,
		"issuingAuthority":   eIDASAuthority,
	}

	if accessToken != nil {
		tokenJSON, err := json.Marshal(accessToken)
		if err != nil {
			return nil, errl.Errorf("error marshalling access token: %w", err)
		}

		// Attach the access token that justifies the creation of the object
		attch := map[string]any{
			"@type":       "attachment",
			"name":        "verifiableCredential",
			"contentType": "application/json",
			"content":     base64.StdEncoding.EncodeToString(tokenJSON),
		}

		theIdentification["attachment"] = attch

	}

	// Prepare organizationIdentification
	orgIdentification := []any{
		theIdentification,
	}

	// Prepare contactMedium
	var contactMedium []any
	if user.EmailAddress != "" {
		contactMedium = append(contactMedium, map[string]any{
			"@type":        "EmailContactMedium",
			"preferred":    true,
			"emailAddress": user.EmailAddress,
		})
	}

	orgMap := map[string]any{
		"@type":                      objectType,
		"isLegalEntity":              true,
		"id":                         id,
		"href":                       id,
		"version":                    version,
		"lastUpdate":                 lastUpdate,
		"name":                       user.Organization,
		"tradingName":                user.Organization,
		"contactMedium":              contactMedium,
		"organizationIdentification": orgIdentification,
		"externalReference": []any{
			map[string]any{
				"externalReferenceType": "idm_id",
				"name":                  user.OrganizationIdentifier,
			},
		},
	}

	content, err := json.Marshal(orgMap)
	if err != nil {
		return nil, errl.Errorf("error marshalling organization: %w", err)
	}

	org := &TMFObject{
		ID:         id,
		Type:       objectType,
		Version:    version,
		LastUpdate: lastUpdate,
		Content:    content,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return org, nil
}
