package service

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/jpath"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	"gitlab.com/greyxor/slogor"
)

const AllowFakeClaims = true

// extractCallerInfo retrieves the Access Token from the request, verifies it if it exists and
// creates a map ready to be passed to the rules engine.
//
// The access token may not exist, but if it does then it must be valid.
// For convenience of the policies, some calculated fields are created and returned in the 'user' object.
func (svc *Service) extractCallerInfo(r *Request) (tokenClaims map[string]any, err error) {

	var authUser *AuthUser

	// This is to support testing
	verify := true

	if len(r.AccessToken) == 0 {
		// The user did not provide an access token.
		// Normally this is forbidden, but for testing we can provide a fake one, and do not verify signature
		if AllowFakeClaims {
			slog.Debug("PDP: using fake claims for testing")
			r.AccessToken = FakeAT
			verify = false
		}
	}

	// An empty token is not considered an error, and the caller should enforce its existence
	if len(r.AccessToken) == 0 {
		return nil, nil
	}

	// It is an error to send an invaild token with the request, so we have to verify it.

	// Verify the token and extract the claims.
	// A verification error stops processing.

	tokenClaims, authUser, err = ParseJWT(svc, r.AccessToken, verify)
	if err != nil {
		slog.Error("invalid access token", slogor.Err(err), "token", r.AccessToken)
		return nil, errl.Errorf("invalid access token: %w", err)
	}

	verifiableCredential := jpath.GetMap(tokenClaims, "vc")

	if len(verifiableCredential) > 0 {
		authUser.isAuthenticated = true

		powers := jpath.GetList(verifiableCredential, "credentialSubject.mandate.power")
		for _, p := range powers {

			// This is to support old version of the Verifiable Credential
			ptype := jpath.GetString(p, "type")
			pdomain := jpath.GetString(p, "domain")
			pfunction := jpath.GetString(p, "function")
			paction := jpath.GetString(p, "action")

			// Check fields without regards to case
			if strings.EqualFold(ptype, "Domain") &&
				strings.EqualFold(pdomain, "DOME") &&
				strings.EqualFold(pfunction, "Onboarding") &&
				strings.EqualFold(paction, "execute") {

				authUser.isLEAR = true

			}

			// And this for the new version of the Verifiable Credential
			ptype = jpath.GetString(p, "tmf_type")
			pdomain = jpath.GetString(p, "tmf_domain")
			pfunction = jpath.GetString(p, "tmf_function")
			paction = jpath.GetString(p, "tmf_action")

			if strings.EqualFold(ptype, "Domain") &&
				strings.EqualFold(pdomain, "DOME") &&
				strings.EqualFold(pfunction, "Onboarding") &&
				strings.EqualFold(paction, "execute") {

				authUser.isLEAR = true
			}

		}

	} else {

		// There is not a Verifiable Credential inside the token
		return nil, errl.Errorf("access token without verifiable credential: %s", r.AccessToken)

	}

	r.AuthUser = authUser

	if len(authUser.OrganizationIdentifier) > 0 {
		// Create a new organization object. If it is created, we just receive an error which we ignore

		org := &repository.Organization{
			CommonName:             authUser.CommonName,
			Country:                authUser.Country,
			EmailAddress:           authUser.EmailAddress,
			Organization:           authUser.Organization,
			OrganizationIdentifier: authUser.OrganizationIdentifier,
			SerialNumber:           authUser.SerialNumber,
		}

		obj, _ := repository.TMFOrganizationFromToken(tokenClaims, org)

		if err := svc.createObject(obj); err != nil {
			if errors.Is(err, &ErrObjectExists{}) {
				slog.Debug("organization already exists", "organizationIdentifier", authUser.OrganizationIdentifier)
			} else {
				err = errl.Error(err)
				return nil, err
			}
		}
	}

	return tokenClaims, nil

}

// setSellerAndBuyerInfo adds the required fields to the incoming object argument
// Specifically, the Seller and SellerOperator roles are added to the relatedParty list
func setSellerAndBuyerInfo(tmfObjectMap map[string]any, organizationIdentifier string) (err error) {

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
			"href":          "urn:ngsi-ld:organization:" + config.ServerOperatorDid,
			"id":            "urn:ngsi-ld:organization:" + config.ServerOperatorDid,
			"name":          config.ServerOperatorDid,
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

func getSellerAndBuyerInfo(tmfObjectMap map[string]any) (sellerDid string, sellerOperatorDid string, err error) {

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

		if rpRole != "seller" && rpRole != "selleroperator" {
			// Go to next entry
			continue
		}

		if rpRole == "seller" {
			party, _ := rpMap["partyOrPartyRole"].(map[string]any)
			sellerDid, _ = party["name"].(string)
			continue
		}
		if rpRole == "selleroperator" {
			party, _ := rpMap["partyOrPartyRole"].(map[string]any)
			sellerOperatorDid, _ = party["name"].(string)
			continue
		}

	}

	if sellerDid == "" {
		err = errl.Errorf("no seller")
		return
	}
	if sellerOperatorDid == "" {
		err = errl.Errorf("no seller operator")
		return
	}

	return

}
