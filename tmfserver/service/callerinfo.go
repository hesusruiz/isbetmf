package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-viper/mapstructure/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/jpath"
	"github.com/hesusruiz/isbetmf/types"
	"gitlab.com/greyxor/slogor"
)

// ProcessAccessToken verifies the Access Token received from the caller and
// creates a map ready to be passed to the rules engine.
//
// The access token may not exist, but if it does then it must be valid.
// For convenience of the policies, some calculated fields are created and returned in the 'user' object.
func (svc *Service) ProcessAccessToken(accessToken string) (user *types.AuthUser, err error) {

	authUser := &types.AuthUser{}

	// An empty token is not considered an error, and the caller should enforce its existence if needed
	if len(accessToken) == 0 {
		return authUser, nil
	}

	// TODO: replace with a setting
	// This is for testing purposes only. It allows to simulate a LEAR user without a real token.
	if accessToken == svc.adminToken {

		authUser.CommonName = svc.ServerOperatorName
		authUser.Country = svc.ServerOperatorCountry
		authUser.EmailAddress = svc.ServerEmailAddress
		authUser.Organization = svc.ServerOperatorName
		authUser.OrganizationIdentifier = svc.ServerOperatorOrganizationIdentifier
		authUser.SerialNumber = "1234567Y"

		authUser.IsAuthenticated = true
		authUser.IsLEAR = true
		authUser.IsOwner = true
		authUser.ProductCreatePower = true
		authUser.ProductUpdatePower = true
		authUser.ProductDeletePower = true

		authUser.TokenMap = make(map[string]any)
		authUser.TokenMap["tokenType"] = ISBEAccessToken
		authUser.TokenMap["user_identifier"] = authUser.SerialNumber
		authUser.TokenMap["organization"] = authUser.Organization
		authUser.TokenMap["organization_identifier"] = authUser.OrganizationIdentifier
		authUser.TokenMap["name"] = authUser.CommonName
		authUser.TokenMap["country"] = authUser.Country
		authUser.TokenMap["email"] = authUser.EmailAddress
		authUser.TokenMap["serial_number"] = authUser.SerialNumber

		return authUser, nil

	}

	// It is an error to send an invaild token with the request, so we have to verify it.
	// We verify the token and extract the claims, a verification error stops processing.

	var token *jwt.Token
	var theClaims = jwt.MapClaims{}

	if svc.Features.VerifyJWTSignature {

		// This is called by the JWT signature verifier to retrieve the verification key
		verifierPublicKeyFunc := func(tok *jwt.Token) (any, error) {
			if svc.oid == nil {
				return nil, errl.Errorf("openid support not initialized")
			}
			keyID, ok := tok.Header["kid"].(string)
			if !ok {
				return nil, errl.Errorf("invalid access token: kid not found in header")
			}
			vk, err := svc.oid.VerificationJWKKey(keyID)
			if err != nil {
				return nil, errl.Error(err)
			}
			slog.Debug("publicKeyFunc", "key", vk)
			return vk.Key, nil
		}

		// Validate and verify the token
		token, err = jwt.NewParser().ParseWithClaims(accessToken, theClaims, verifierPublicKeyFunc)
		if err != nil {
			slog.Error("invalid access token", slogor.Err(err), "token", accessToken)
			return nil, errl.Errorf("invalid access token: %w, token: %s", err, accessToken)
		}

	} else {

		// Parse the token without signature verification
		token, _, err = new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
		if err != nil {
			slog.Error("invalid access token", slogor.Err(err), "token", accessToken)
			return nil, errl.Errorf("invalid access token: %w, token: %s", err, accessToken)
		}

	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("JWT claims are not of type MapClaims")
		return nil, errors.New("invalid JWT claims format")
	}

	// The actual claims depends on the caller: the standard DOME or the ISBE one, which does not send the "vc" claim.
	if svc.IsDOME() {
		return svc.processDOMEAccessToken(claims, accessToken)
	} else {
		return svc.processISBEAccessToken(claims, accessToken)
	}

}

type accessTokenType string

const (
	DOMEAccessToken accessTokenType = "DOME"
	ISBEAccessToken accessTokenType = "ISBE"
)

func (svc *Service) processDOMEAccessToken(claims jwt.MapClaims, accessToken string) (user *types.AuthUser, err error) {

	authUser := &types.AuthUser{}
	authUser.IsAuthenticated = true

	// Extract the Verifiable Credential from the claims
	verifiableCredential := jpath.GetMap(claims, "vc")
	if len(verifiableCredential) == 0 {
		// There is not a Verifiable Credential inside the token
		return nil, errl.Errorf("access token without verifiable credential: %s", accessToken)
	}

	credentialSubject := jpath.GetMap(verifiableCredential, "credentialSubject")
	if len(credentialSubject) == 0 {
		slog.Debug("JWT payload does not contain 'credentialSubject' field or it's not a map")
		return nil, errors.New("missing 'credentialSubject' in JWT claims")
	}

	mandate := jpath.GetMap(credentialSubject, "mandate")
	if len(mandate) == 0 {
		slog.Debug("JWT payload does not contain 'mandate' field or it's not a map")
		return nil, errors.New("missing 'mandate' in JWT claims")
	}

	mandatorData := jpath.GetMap(mandate, "mandator")
	if len(mandatorData) == 0 {
		slog.Debug("JWT payload does not contain 'mandator' field or it's not a map")
		return nil, errors.New("missing 'mandator' in JWT claims")
	}

	// Marshal and unmarshal to AuthUser struct for type safety and JSON tag mapping
	mandatorJSON, err := json.Marshal(mandatorData)
	if err != nil {
		slog.Error("Failed to marshal mandator data", slog.Any("error", err))
		return nil, fmt.Errorf("failed to process mandator data: %w", err)
	}

	if err := json.Unmarshal(mandatorJSON, authUser); err != nil {
		slog.Error("Failed to unmarshal mandator data to AuthUser", slog.Any("error", err))
		return nil, fmt.Errorf("failed to process mandator data: %w", err)
	}

	slog.Debug("Successfully parsed AuthUser from JWT",
		slog.String("organizationIdentifier", authUser.OrganizationIdentifier),
		slog.String("country", authUser.Country))

	claims["tokenType"] = DOMEAccessToken

	authUser.AccessToken = accessToken
	authUser.TokenMap = claims

	// Parse the user powers to look for the ones we only care about
	authUserPowers := jpath.GetList(verifiableCredential, "credentialSubject.mandate.power")

	return svc.procesPowers(authUserPowers, authUser), nil

}

func (svc *Service) processISBEAccessToken(claims jwt.MapClaims, accessToken string) (user *types.AuthUser, err error) {
	authUser := &types.AuthUser{}
	authUser.IsAuthenticated = true

	// Get the organization data

	authUser.Organization = jpath.GetString(claims, "organization")
	if len(authUser.Organization) == 0 {
		slog.Debug("JWT payload does not contain 'organization' field or it's not a string")
		return nil, errl.Errorf("missing 'organization' in JWT claims")
	}

	authUser.OrganizationIdentifier = jpath.GetString(claims, "organization_identifier")
	if len(authUser.OrganizationIdentifier) == 0 {
		slog.Debug("JWT payload does not contain 'organization_identifier' field or it's not a string")
		return nil, errl.Errorf("missing 'organization_identifier' in JWT claims")
	}

	authUser.CommonName = jpath.GetString(claims, "name")
	if len(authUser.CommonName) == 0 {
		slog.Debug("JWT payload does not contain 'name' field or it's not a string")
		return nil, errl.Errorf("missing 'name' in JWT claims")
	}

	authUser.SerialNumber = jpath.GetString(claims, "user_identifier")
	if len(authUser.SerialNumber) == 0 {
		slog.Debug("JWT payload does not contain 'user_identifier' field or it's not a string")
		return nil, errl.Errorf("missing 'user_identifier' in JWT claims")
	}

	// TODO: the token from ISBE should contain the country
	authUser.Country = jpath.GetString(claims, "country")
	if len(authUser.Country) == 0 {
		slog.Debug("JWT payload does not contain 'country' field or it's not a string")
		authUser.Country = "ES"
	}

	authUser.EmailAddress = jpath.GetString(claims, "email")
	if len(authUser.EmailAddress) == 0 {
		slog.Debug("JWT payload does not contain 'email' field or it's not a string")
		return nil, errl.Errorf("missing 'email' in JWT claims")
	}

	claims["tokenType"] = ISBEAccessToken

	authUser.AccessToken = accessToken
	authUser.TokenMap = claims

	authUserPowers := jpath.GetList(claims, "power")
	if len(authUserPowers) == 0 {
		return nil, errl.Errorf("missing 'power' in JWT claims")
	}

	return svc.procesPowers(authUserPowers, authUser), nil
}

func (svc *Service) procesPowers(authUserPowers []any, authUser *types.AuthUser) *types.AuthUser {

	// Parse the user powers for the ones we only care about
	for _, p := range authUserPowers {
		var userPower types.OnePower
		if err := mapstructure.Decode(p, &userPower); err != nil {
			slog.Error("error decoding power", slogor.Err(err))
			continue
		}

		if userPower.Includes(svc.LEARPower) {
			authUser.IsLEAR = true
			// A LEAR can create, update and delete product offerings
			authUser.ProductCreatePower = true
			authUser.ProductUpdatePower = true
			authUser.ProductDeletePower = true
		} else {
			if userPower.Includes(svc.ProductCreatePower) {
				authUser.ProductCreatePower = true
			}
			if userPower.Includes(svc.ProductUpdatePower) {
				authUser.ProductUpdatePower = true
			}
			if userPower.Includes(svc.ProductDeletePower) {
				authUser.ProductDeletePower = true
			}
		}

	}

	return authUser
}
