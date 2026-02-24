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
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/hesusruiz/isbetmf/types"
	"gitlab.com/greyxor/slogor"
)

const AllowFakeClaims = true
const VerifyTokenSignature = false

// ProcessAccessToken retrieves the Access Token from the request, verifies it if it exists and
// creates a map ready to be passed to the rules engine.
//
// The access token may not exist, but if it does then it must be valid.
// For convenience of the policies, some calculated fields are created and returned in the 'user' object.
func (svc *Service) ProcessAccessToken(accessToken string) (tokenClaims map[string]any, user *types.AuthUser, err error) {

	authUser := &types.AuthUser{}

	// This is to support testing
	verify := true

	// if len(r.AccessToken) == 0 {
	// 	// The user did not provide an access token.
	// 	// Normally this is forbidden, but for testing we can provide a fake one, and do not verify signature
	// 	if AllowFakeClaims && r.Method != "GET" {
	// 		slog.Debug("using fake claims for testing")
	// 		r.AccessToken = FakeATold
	// 		verify = false
	// 	}
	// }

	// An empty token is not considered an error, and the caller should enforce its existence if needed
	if len(accessToken) == 0 {
		return nil, authUser, nil
	}

	if accessToken == "eyJhdWQiOiJodHRwczovL2NhdGFsb2cuaX" {
		accessToken = FakeATold
	}

	// It is an error to send an invaild token with the request, so we have to verify it.
	// We verify the token and extract the claims, a verification error stops processing.

	// For testing
	if !VerifyTokenSignature {
		verify = false
	}

	tokenClaims, authUser, err = svc.parseAccessToken(accessToken, verify)
	if err != nil {
		slog.Error("invalid access token", slogor.Err(err), "token", accessToken)
		return nil, nil, errl.Errorf("invalid access token: %w, token: %s", err, accessToken)
	}
	authUser.IsAuthenticated = true

	if tokenClaims["tokenType"] == DOMEAccessToken {

		verifiableCredential := jpath.GetMap(tokenClaims, "vc")

		if len(verifiableCredential) == 0 {
			// There is not a Verifiable Credential inside the token
			return nil, nil, errl.Errorf("access token without verifiable credential: %s", accessToken)
		}

		// Parse the user powers for the ones we only care about
		authUserPowers := jpath.GetList(verifiableCredential, "credentialSubject.mandate.power")
		for _, p := range authUserPowers {
			var power types.OnePower
			if err := mapstructure.Decode(p, &power); err != nil {
				slog.Error("error decoding power", slogor.Err(err))
				continue
			}

			if power.Includes(svc.LEARPower) {
				authUser.IsLEAR = true
				// A LEAR can create, update and delete product offerings
				authUser.ProductCreatePower = true
				authUser.ProductUpdatePower = true
				authUser.ProductDeletePower = true
			}

			if power.Includes(svc.ProductCreatePower) {
				authUser.ProductCreatePower = true
			}
			if power.Includes(svc.ProductUpdatePower) {
				authUser.ProductUpdatePower = true
			}
			if power.Includes(svc.ProductDeletePower) {
				authUser.ProductDeletePower = true
			}
		}

		// If the user has an organization identifier, create a new organization object.
		// If it is already created, we just receive an error which we ignore
		// This is needed to be able to create a new organization object in the DOME Marketplace.
		if len(authUser.OrganizationIdentifier) > 0 {

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
					err = errl.Errorf("error creating organization object %s: %w", authUser.OrganizationIdentifier, err)
					return nil, nil, err
				}
			}

			// Check if this is not a LEARCredentialMachine
			// We identify this by looking at the Mandatee and if there is a field called "serviceName"
			serviceName := jpath.GetString(verifiableCredential, "credentialSubject.mandate.mandatee.serviceName")
			if len(serviceName) > 0 {
				// This is a LEARCredentialMachine, just return
				return tokenClaims, authUser, nil
			}

			// This must be a LEARCredentialEmployye, create a new user if not yet created
			individual, err := repository.TMFIndividualFromCredential(verifiableCredential, org)
			if err != nil {
				slog.Error("error parsing individual object", slogor.Err(err))
			} else {
				if err := svc.createObject(individual); err != nil {
					if errors.Is(err, &ErrObjectExists{}) {
						slog.Debug("individual already exists", "id", individual.ID)
					} else {
						err = errl.Errorf("error creating individual object %s: %w", individual.ID, err)
						return nil, nil, err
					}
				}

			}

		}

		return tokenClaims, authUser, nil

	} else {

		// Parse the user powers for the ones we only care about
		authUserPowers := jpath.GetList(tokenClaims, "power")
		for _, p := range authUserPowers {
			var power types.OnePower
			if err := mapstructure.Decode(p, &power); err != nil {
				slog.Error("error decoding power", slogor.Err(err))
				continue
			}

			if power.Includes(svc.LEARPower) {
				authUser.IsLEAR = true
				// A LEAR can create, update and delete product offerings
				authUser.ProductCreatePower = true
				authUser.ProductUpdatePower = true
				authUser.ProductDeletePower = true
			}

			if power.Includes(svc.ProductCreatePower) {
				authUser.ProductCreatePower = true
			}
			if power.Includes(svc.ProductUpdatePower) {
				authUser.ProductUpdatePower = true
			}
			if power.Includes(svc.ProductDeletePower) {
				authUser.ProductDeletePower = true
			}
		}

		// If the user has an organization identifier, create a new organization object.
		// If it is already created, we just receive an error which we ignore
		// This is needed to be able to create a new organization object in the DOME Marketplace.
		if len(authUser.OrganizationIdentifier) > 0 {

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
					err = errl.Errorf("error creating organization object %s: %w", authUser.OrganizationIdentifier, err)
					return nil, nil, err
				}
			}

		}

		return tokenClaims, authUser, nil
	}

}

type accessTokenType string

const (
	DOMEAccessToken accessTokenType = "DOME"
	ISBEAccessToken accessTokenType = "ISBE"
)

// parseAccessToken parses a JWT string, extracts the mandator information, and returns an AuthUser.
// It does NOT verify the JWT signature.
func (svc *Service) parseAccessToken(tokenString string, verify bool) (tokenClaims map[string]any, u *types.AuthUser, err error) {

	var token *jwt.Token
	var theClaims = jwt.MapClaims{}

	if verify {

		// This is called by the JWT signature verifier to retrieve the verification key
		verifierPublicKeyFunc := func(*jwt.Token) (any, error) {
			if svc.oid == nil {
				return nil, errl.Errorf("openid support not initialized")
			}
			vk, err := svc.oid.VerificationJWK()
			if err != nil {
				return nil, errl.Error(err)
			}
			slog.Debug("publicKeyFunc", "key", vk)
			return vk.Key, nil
		}

		// Validate and verify the token
		token, err = jwt.NewParser().ParseWithClaims(tokenString, theClaims, verifierPublicKeyFunc)
		if err != nil {
			slog.Error("Failed to parse JWT unverified", slog.Any("error", err))
			return nil, nil, fmt.Errorf("failed to parse JWT: %w", err)
		}

	} else {

		// Parse the token without signature verification
		token, _, err = new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
		if err != nil {
			slog.Error("Failed to parse JWT unverified", slog.Any("error", err))
			return nil, nil, fmt.Errorf("failed to parse JWT: %w", err)
		}

	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Error("JWT claims are not of type MapClaims")
		return nil, nil, errors.New("invalid JWT claims format")
	}

	// The actual claims depends on the caller: the standard DOME or the ISBE one, which does not send the "vc" claim.

	tokenType := DOMEAccessToken

	// Extract mandator object from vc.credentialSubject.mandate.mandator
	vc, ok := claims["vc"].(map[string]any)
	if !ok {
		power := jpath.GetList(claims, "power")
		if len(power) == 0 {
			slog.Debug("JWT payload does not contain 'vc' field or it's not a map")
			return nil, nil, errors.New("missing 'vc' in JWT claims")
		}
		tokenType = ISBEAccessToken
	}

	if tokenType == DOMEAccessToken {

		credentialSubject, ok := vc["credentialSubject"].(map[string]any)
		if !ok {
			slog.Debug("JWT payload does not contain 'credentialSubject' field or it's not a map")
			return nil, nil, errors.New("missing 'credentialSubject' in JWT claims")
		}

		mandate, ok := credentialSubject["mandate"].(map[string]any)
		if !ok {
			slog.Debug("JWT payload does not contain 'mandate' field or it's not a map")
			return nil, nil, errors.New("missing 'mandate' in JWT claims")
		}

		mandatorData, ok := mandate["mandator"].(map[string]any)
		if !ok {
			slog.Debug("JWT payload does not contain 'mandator' field or it's not a map")
			return nil, nil, errors.New("missing 'mandator' in JWT claims")
		}

		// Marshal and unmarshal to AuthUser struct for type safety and JSON tag mapping
		mandatorJSON, err := json.Marshal(mandatorData)
		if err != nil {
			slog.Error("Failed to marshal mandator data", slog.Any("error", err))
			return nil, nil, fmt.Errorf("failed to process mandator data: %w", err)
		}

		var authUser types.AuthUser
		if err := json.Unmarshal(mandatorJSON, &authUser); err != nil {
			slog.Error("Failed to unmarshal mandator data to AuthUser", slog.Any("error", err))
			return nil, nil, fmt.Errorf("failed to process mandator data: %w", err)
		}

		slog.Debug("Successfully parsed AuthUser from JWT",
			slog.String("organizationIdentifier", authUser.OrganizationIdentifier),
			slog.String("country", authUser.Country))

		claims["tokenType"] = DOMEAccessToken
		return claims, &authUser, nil

	} else {

		var authUser types.AuthUser

		authUser.Organization = claims["organization"].(string)
		authUser.OrganizationIdentifier = claims["organization_identifier"].(string)
		authUser.CommonName = claims["name"].(string)
		authUser.SerialNumber = claims["user_identifier"].(string)

		// TODO: the token from ISBE should contain the country
		authUser.Country = "SPAIN"

		authUser.EmailAddress = claims["email"].(string)

		claims["tokenType"] = ISBEAccessToken
		return claims, &authUser, nil

	}

}
