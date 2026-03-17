package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/jpath"
	pdp "github.com/hesusruiz/isbetmf/pdp"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// takeDecision evaluates access authorization for a TMF request against both hardcoded and user-defined policies.
// It first checks hardcoded policies, and if they pass, proceeds to evaluate user policies in the PDP engine.
//
// Parameters:
//   - ruleEngine: PDP engine instance for evaluating user policies
//   - req: The request details containing auth and operation information
//   - tokenClaims: JWT claims from the authentication token
//   - objectMap: Map of TMF objects involved in the request
//
// Returns:
//   - authorized: boolean indicating if access is authorized
//   - err: error if any occurred during policy evaluation
//
// The function works in two stages:
// 1. Evaluates hardcoded policies first - if these fail, denies access immediately
// 2. If hardcoded policies pass, evaluates user-defined policies through the PDP engine
func (svc *Service) takeDecision(
	ruleEngine *pdp.PDP,
	req *Request,
	objectMap repo.TMFObjectMap,
) (authorized bool, err error) {

	// Check if the caller is authenticated and is the server operator.
	// If so, we grant access immediately.
	if req.AuthUser.IsAuthenticated && req.AuthUser.OrganizationIdentifier == svc.ServerOperatorDid {
		return true, nil
	}

	// Evaluate the hardcoded policies, if they fail return immediately.
	// Otherwise, continue to see if the user policies allow access
	decision, reason := svc.hardcodedPolicies(req, objectMap)
	if !decision {
		return false, reason
	}

	// The caller is the owner, at least according to hardcoded policies.
	// The user policies will determine the final decision.
	req.AuthUser.IsOwner = decision

	if err := svc.userPolicies(ruleEngine, req, req.TokenMap, objectMap); err != nil {
		return false, errl.Errorf("user policies in PDP engine: %w", err)
	}

	return true, nil

}

// hardcodedPolicies determines if a request is authorized based on predefined access control rules.
// It evaluates the request against different policies depending on the object type and user authentication.
//
// The following rules are applied:
// - GET requests to public resources are allowed for all users (authenticated or not)
// - All other operations require authenticated users
// - For Organization objects:
//   - Server operator has full access
//   - Organization members have full access to their own organization
//
// - For Individual objects:
//   - Server operator has full access
//   - Organizations listed as mandators in LEARCredential have full access
//
// - For Category objects:
//   - Only server operator has full access
//
// - For all other objects:
//   - Server operator has full access
//   - Seller and SellerOperator have full access
//   - Buyer and BuyerOperator have full access if buyer info exists
//
// Parameters:
//   - req: Contains request details including method, resource name and authenticated user info
//   - obj: The TMF object map being accessed
//
// Returns:
//   - decision: true if access is granted, false otherwise
//   - reason: error containing the reason for the decision
func (svc *Service) hardcodedPolicies(req *Request, obj repo.TMFObjectMap) (decision bool, reason error) {

	// Try to retrieve the Seller and Buyer info in the object
	objSeller, objSellerOperator, _ := obj.GetSellerInfo("v4")
	objBuyer, objBuyerOperator, _ := obj.GetBuyerInfo("v4")

	// objSeller and objSellerOperator must be both empty or both not empty. We do not accept partial information.
	// The same applies to objBuyer and objBuyerOperator.
	if (objSeller == "" && objSellerOperator != "") || (objSeller != "" && objSellerOperator == "") {
		return false, errl.Errorf("objSeller and objSellerOperator must both be set or both be empty, got objSeller='%s', objSellerOperator='%s'", objSeller, objSellerOperator)
	}
	if (objBuyer == "" && objBuyerOperator != "") || (objBuyer != "" && objBuyerOperator == "") {
		return false, errl.Errorf("objBuyer and objBuyerOperator must both be set or both be empty, got objBuyer='%s', objBuyerOperator='%s'", objBuyer, objBuyerOperator)
	}

	// Method GET includes actions READ and LIST
	if req.Method == "GET" {

		// Read operations (GET) to public resources are allowed to all users, even unauthenticated ones.
		// But this is true only if the object does not have Buyer info set (like a private productOffering for a special tender).
		if config.IsPublicResource(req.ResourceName) && objBuyer == "" && objBuyerOperator == "" {
			return true, errl.Errorf("GET request to a public resource %s", req.ResourceName)
		}

	}

	// Any other operation is only allowed to authenticated users.
	if !req.AuthUser.IsAuthenticated {
		return false, errl.Errorf("user not authenticated")
	}

	// Determining ownership of an object depends on the type of object. There are some "special" objects,
	// like Organization, Individual or Category, which we treat differently.
	objType, _ := obj["@type"].(string)
	objType = strings.ToLower(objType)

	caller := req.AuthUser

	switch objType {
	case "organization":

		// If the caller is Altia (the server operator), then we can read/write/update/delete
		if repo.SameOrganizations(caller.OrganizationIdentifier, "VATES-A15456585") {
			return true, errl.Errorf("caller %s is server operator %s", caller.OrganizationIdentifier, svc.ServerOperatorDid)
		}

		// If the caller is us (the server operator), then we can read/write/update/delete
		if repo.SameOrganizations(caller.OrganizationIdentifier, svc.ServerOperatorDid) {
			return true, errl.Errorf("caller %s is server operator %s", caller.OrganizationIdentifier, svc.ServerOperatorDid)
		}

		// If the organization of the caller and object are the same, then the caller can read/write/update/delete
		objectOrganizationId := jpath.GetString(obj, "organizationIdentification.*.identificationId")
		if repo.SameOrganizations(objectOrganizationId, caller.OrganizationIdentifier) {
			return true, errl.Errorf("caller %s is same as in object %s", caller.OrganizationIdentifier, objectOrganizationId)
		}

		return false, errl.Errorf("caller (%s) is neither the same as in object (%s) or the server operator", caller.OrganizationIdentifier, objectOrganizationId)

	case "individual":

		// TODO: revise this policy to be more restrictive

		// If the caller is us (the server operator), then we can read/write/update/delete
		if repo.SameOrganizations(caller.OrganizationIdentifier, svc.ServerOperatorDid) {
			return true, errl.Errorf("caller %s is server operator %s", caller.OrganizationIdentifier, svc.ServerOperatorDid)
		}

		// If the caller is the Organization that is the mandator in the LEARCredential of the employee
		// then the caller can read/write/update/delete
		individualIdentificationArray := jpath.GetList(obj, "individualIdentification")

		// Look for an entry with 'identificationType=learcredentialemployee'
		for _, individualIdentification := range individualIdentificationArray {
			individualIdentificationMap, _ := individualIdentification.(map[string]any)
			if individualIdentificationMap["identificationType"] == "learcredentialemployee" {
				// The 'issuingAuthority' must be equal to the caller organizationIdentifier
				issuingAuthority := individualIdentificationMap["issuingAuthority"].(string)
				if repo.SameOrganizations(issuingAuthority, caller.OrganizationIdentifier) {
					return true, errl.Errorf("caller %s is same as mandator in Individual object %s", caller.OrganizationIdentifier, issuingAuthority)
				} else {
					return false, errl.Errorf("caller (%s) is neither the mandator in Individual object (%s) or the server operator", caller.OrganizationIdentifier, issuingAuthority)
				}
			}
		}

		return false, errl.Errorf("caller (%s) is neither the mandator in Individual object or the server operator", caller.OrganizationIdentifier)

	case "category":

		// If the caller is us (the server operator), then we can read/write/update/delete
		if repo.SameOrganizations(caller.OrganizationIdentifier, svc.ServerOperatorDid) {
			return true, errl.Errorf("caller %s is server operator %s", caller.OrganizationIdentifier, svc.ServerOperatorDid)
		}

		return false, errl.Errorf("caller %s is not the server operator %s", caller.OrganizationIdentifier, svc.ServerOperatorDid)

	}

	// There are several items which play in the hardcoded rules (not modifiable by the users) for access control:
	// 1. The Caller, which we take from the authentication token. The Caller can be of different types:
	//    - It can be ourselves (the ServerOperator).
	//    - It can be an organization acting as federated Marketplace, with a specific marketplace agreement with ourselves.
	//    - It can be an organization which is "normal" CSP.
	// 2. The powers of the entity acting on behalf of the Caller, that is, the actual entity invoking the API.
	//    - The powers can be Onboard (only for the ServerOperator) and Product/Create (for anybody, including the ServerOperator)
	//    - If the ServerOperator has Onboard power, it is the Admin and can do anything.
	//    - If the ServerOperator has Product/Create power, it can do anything but only for its Sellers, not for those of other Marketplaces
	//    - For anybody else, they can manage their own products
	// 3. The Seller and SellerOperator information in the object
	//    - For CREATE, we check that the info is correct and can amend it in some circumstances
	//    - For UPDATE/DELETE, we check that the caller can manage the object
	// 4. The Buyer and BuyerOperator information in the object

	// If the request is a CREATE, we implement a fix for callers which do not set the Seller info in the incomingobject.
	// For other requests, the Seller info must be already set in the object.
	// Note that we do not do the same for the Buyer info, which is optional and may not exist.
	if req.Action == CREATE {
		// If Seller and SellerOperator are empty, they are set to the Caller and ServerOperator, respectively
		// Note that we do not do the same for the Buyer info, which is optional and may not exist.
		if objSeller == "" && objSellerOperator == "" {
			objSeller = caller.OrganizationIdentifier
			objSellerOperator = svc.ServerOperatorDid
			err := obj.SetSellerInfo(objSellerOperator, objSeller, "v4")
			if err != nil {
				return false, errl.Errorf("error trying to set seller info: %w", err)
			}
		}
	} else {
		// For other requests, the Seller info must be already set in the object.
		if objSeller == "" || objSellerOperator == "" {
			return false, errl.Errorf("seller info is not set in the object")
		}
	}

	// If the caller is us (the server operator), it may be because:
	// 1. The caller is an application operated by the ServerOperator and acting as itself or on behalf of a user who is not present,
	//    and th eapplication is presenting an access token obtained by the app after authenticating with a LEARCredentialMachine.
	// 2. The caller is an application acting on behalf of a user belonging to the ServerOperator organization, and the application is
	//    presenting the access token obtained by the user after authentication with a LEARCredentialEmployee.
	//
	// In this case, the Caller can do (almost) anything, and it depends on the powers that the caller has:
	// - with power Onboard, the caller can do anything and it is the Admin of the server.
	// - with power Product/(Create, Update, Delete), the user can do the relevant action in all objects managed by the ServerOperator. The
	//   caller can not do anything with objects managed by a federated Marketplace.
	if repo.SameOrganizations(caller.OrganizationIdentifier, svc.ServerOperatorDid) {

		// Accept if the caller is a LEAR
		if caller.IsLEAR {
			return true, errl.Errorf("caller %s is server operator %s and is a LEAR", caller.OrganizationIdentifier, svc.ServerOperatorDid)
		}

		// Reject if the caller does not have power to create products
		if req.Action == CREATE && !caller.ProductCreatePower {
			return false, errl.Errorf("caller %s is server operator but does not have power to create products", caller.OrganizationIdentifier)
		}

		// Reject if the caller does not have power to update products
		if req.Action == UPDATE && !caller.ProductUpdatePower {
			return false, errl.Errorf("caller %s is server operator but does not have power to update products", caller.OrganizationIdentifier)
		}

		// Reject if the caller does not have power to delete products
		if req.Action == DELETE && !caller.ProductDeletePower {
			return false, errl.Errorf("caller %s is server operator but does not have power to delete products", caller.OrganizationIdentifier)
		}

		// Reject if the SellerOperator is not us (the ServerOperator)
		if !repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
			return false, errl.Errorf("caller %s is server operator but can not operate Sellers of other Marketplaces", caller.OrganizationIdentifier)
		}

		// Reject if there is Buyer info in the object but the caller is not the BuyerOperator
		if objBuyerOperator != "" && !repo.SameOrganizations(objBuyerOperator, caller.OrganizationIdentifier) {
			return false, errl.Errorf("caller %s is server operator but can not operate Buyers of other Marketplaces", caller.OrganizationIdentifier)
		}

		// Accept if we reach here
		return true, errl.Errorf("caller %s is server operator and has power to operate the object", caller.OrganizationIdentifier)

	}

	// The caller is a "normal" CSP

	switch req.Action {
	case CREATE:
		if !caller.ProductCreatePower {
			return false, errl.Errorf("caller %s does not have power to create products", caller.OrganizationIdentifier)
		}
	case UPDATE:
		if !caller.ProductUpdatePower {
			return false, errl.Errorf("caller %s does not have power to update products", caller.OrganizationIdentifier)
		}
	case DELETE:
		if !caller.ProductDeletePower {
			return false, errl.Errorf("caller %s does not have power to delete products", caller.OrganizationIdentifier)
		}
	}

	// Reject if the SellerOperator is not us (the ServerOperator)
	if !repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
		return false, errl.Errorf("only our CSPs are accepted")
	}

	// If there is Buyer info in the object, the object is a private object only for the Buyer and the Seller
	if objBuyerOperator != "" {
		// Reject if there is Buyer info in the object but the BuyerOperator is not us (the ServerOperator)
		if !repo.SameOrganizations(objBuyerOperator, svc.ServerOperatorDid) {
			return false, errl.Errorf("only our CSPs are accepted")
		}

		// Reject is the caller is not either the Seller or the Buyer
		if !repo.SameOrganizations(objSeller, caller.OrganizationIdentifier) && !repo.SameOrganizations(objBuyer, caller.OrganizationIdentifier) {
			return false, errl.Errorf("the caller %s is not the Seller %s or the Buyer %s", caller.OrganizationIdentifier, objSeller, objBuyer)
		}

	} else {
		// Reject if the caller is not the Seller
		if !repo.SameOrganizations(objSeller, caller.OrganizationIdentifier) {
			return false, errl.Errorf("the caller %s is not the Seller %s", caller.OrganizationIdentifier, objSeller)
		}
	}

	// Accept if we reach here
	return true, errl.Errorf("the CSP is the owner and we are the operator")

}

// userPolicies constructs the policy input for a PDP (policy decision point)
// from the provided request, token claims, TMF object map and authenticated user,
// then evaluates the assembled input against the rules engine to determine
// whether the request is authorized.
//
// The assembled input contains the following top-level keys:
//   - "request": a dereferenced map representation of req (req.ToMap())
//   - "token": a dereferenced map representation of tokenClaims
//   - "tmf": a dereferenced map representation of the incoming TMF object map
//   - "user": a dereferenced map representation of the authenticated user
//
// The function will pretty-print the assembled input as JSON when the default
// slog logger is at debug level to aid debugging.
//
// Decision & error semantics:
//   - If ruleEngine is nil the function defaults to allowing the request (returns nil).
//   - If ruleEngine is present, ruleEngine.Authorize(input) is invoked.
//   - Any error returned by the PDP is treated as a rejection and returned to the caller.
//   - If the PDP returns false the request is considered denied and a non-nil error is returned.
//   - If the PDP returns true the request is authorized and the function returns nil.
//
// Side effects:
//   - Logs an informational message on authorization and prints debug JSON when enabled.
//   - Constructs a single "input" object (OPA-style) so policies can access request,
//     token, tmf and user together and support callbacks.
//
// Return value:
//   - nil when the request is authorized.
//   - non-nil error when the rules engine rejects the request or an internal error occurs.
func (svc *Service) userPolicies(
	ruleEngine *pdp.PDP,
	req *Request,
	tokenClaims map[string]any,
	objectMap repo.TMFObjectMap,
) (err error) {

	userArgument := pdp.StarTMFMap(req.AuthUser.ToMap())
	incomingObjectArgument := pdp.StarTMFMap(objectMap)
	requestArgument := pdp.StarTMFMap(req.ToMap())
	tokenArgument := pdp.StarTMFMap(tokenClaims)

	// Assemble all data in a single "input" argument, to the style of OPA.
	// We mutate the predeclared identifier, so the policy can access the data for this request.
	// We can also service possible callbacks from the rules engine.

	input := map[string]any{
		"request": requestArgument,
		"token":   tokenArgument,
		"tmf":     incomingObjectArgument,
		"user":    userArgument,
	}

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		b, err := json.MarshalIndent(input, "", "  ")
		if err == nil {
			fmt.Println("PDP input:", string(b))
		}
	}

	decision := true
	if ruleEngine != nil {
		decision, err = ruleEngine.Authorize(input)

		// An error is considered a rejection, continue with the next candidate object
		if err != nil {
			return errl.Errorf("rules engine rejected request due to an error: %w", err)
		}
	}

	// The rules engine rejected the request, continue with the next candidate object
	if !decision {
		return errl.Errorf("PDP: request rejected due to policy")
	}

	// The rules engine accepted the request, add the object to the final list
	slog.Info("PDP: request authorised")
	return nil
}
