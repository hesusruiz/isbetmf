package service

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/jpath"
	pdp "github.com/hesusruiz/isbetmf/pdp"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/hesusruiz/isbetmf/types"
)

// checkAuthorization evaluates access authorization for a TMF request against both hardcoded and user-defined policies.
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
func (svc *Service) checkAuthorization(
	ruleEngine Authorizer,
	req *Request,
	objectMap repo.TMFObjectMap,
) (authorized bool, err error) {

	// Evaluate the hardcoded policies, if they fail return immediately.
	// Otherwise, continue to see if the user policies allow access
	_, err = svc.hardcodedPolicies(req, objectMap)
	if err != nil {
		return false, errl.Error(err)
	}

	// The caller is the owner, at least according to hardcoded policies.
	// The user policies will determine the final decision.
	req.AuthUser.IsOwner = true

	if err := svc.userPolicies(ruleEngine, req, req.AuthUser.TokenMap, objectMap); err != nil {
		return false, errl.Error(err)
	}

	return true, nil

}

func (svc *Service) isServerOperator(user types.AuthUser) bool {
	return repo.SameOrganizations(user.OrganizationIdentifier, svc.ServerOperatorDid)
}

func (svc *Service) isTrustedParty(user types.AuthUser) bool {
	for _, trustedParty := range svc.AdditionalTrustedparties {
		if repo.SameOrganizations(user.OrganizationIdentifier, trustedParty.Did) {
			return true
		}
	}

	return false
}

// // hardcodedPolicies determines if a request is authorized based on predefined access control rules.
// // It evaluates the request against different policies depending on the object type and user authentication.
// //
// // The following rules are applied:
// // - GET requests to public resources are allowed for all users (authenticated or not)
// // - All other operations require authenticated users
// // - For Organization objects:
// //   - Server operator has full access
// //   - Organization members have full access to their own organization
// //
// // - For Individual objects:
// //   - Server operator has full access
// //   - Organizations listed as mandators in LEARCredential have full access
// //
// // - For Category objects:
// //   - Only server operator has full access
// //
// // - For all other objects:
// //   - Server operator has full access
// //   - Seller and SellerOperator have full access
// //   - Buyer and BuyerOperator have full access if buyer info exists
// //
// // Returns:
// //   - reason: descriptive string explaining why access was authorized (empty on failure)
// //   - err: error containing the reason for rejection (nil on authorization success)
// func (svc *Service) hardcodedPoliciesOld(req *Request, obj repo.TMFObjectMap) (reason string, err error) {

// 	// Try to retrieve the Seller and Buyer info in the object
// 	objSeller, objSellerOperator, _ := obj.GetSellerInfo("")
// 	objBuyer, objBuyerOperator, _ := obj.GetBuyerInfo("")

// 	// objSeller and objSellerOperator must be both empty or both not empty. We do not accept partial information.
// 	// The same applies to objBuyer and objBuyerOperator.
// 	if isPartiallySet(objSeller, objSellerOperator) {
// 		return "", errl.Errorf("objSeller and objSellerOperator must both be set or both be empty, got objSeller='%s', objSellerOperator='%s'", objSeller, objSellerOperator)
// 	}
// 	if isPartiallySet(objBuyer, objBuyerOperator) {
// 		return "", errl.Errorf("objBuyer and objBuyerOperator must both be set or both be empty, got objBuyer='%s', objBuyerOperator='%s'", objBuyer, objBuyerOperator)
// 	}

// 	// Read operations to public resources are allowed to all users, even unauthenticated ones.
// 	// (GET method includes READ and LIST actions)
// 	if req.Method == "GET" {
// 		// But this is true only if the object does not have Buyer info set (like a private productOffering for a special tender).
// 		if obj.IsPotentiallyPublic() && objBuyer == "" {
// 			return "public resource without buyer info", nil
// 		}
// 	}

// 	// Any other operation is only allowed to authenticated users.
// 	if !req.AuthUser.IsAuthenticated {
// 		return "", errl.Errorf("user not authenticated")
// 	}

// 	// Determining ownership of an object depends on the type of object. There are some "special" objects,
// 	// like Organization, Individual or Category, which we treat differently.
// 	objType, _ := obj["@type"].(string)
// 	objType = strings.ToLower(objType)

// 	caller := req.AuthUser

// 	// Perform special processing for certain object types (organization, individual and category),
// 	// which do not have Seller or Buyer information.
// 	switch objType {
// 	case "organization":

// 		// If the caller is us (the server operator), then we can read/write/update/delete
// 		if svc.isServerOperator(caller) {
// 			return "caller is server operator", nil
// 		}

// 		// If the organization of the caller and of the object are the same, then the caller can read/write/update/delete
// 		objectOrganizationId := jpath.GetString(obj, "organizationIdentification.*.identificationId")
// 		if repo.SameOrganizations(objectOrganizationId, caller.OrganizationIdentifier) {
// 			return fmt.Sprintf("caller %s is same as in object %s", caller.OrganizationIdentifier, objectOrganizationId), nil
// 		}

// 		return "", errl.Errorf("caller (%s) is neither the same as in object (%s) or the server operator", caller.OrganizationIdentifier, objectOrganizationId)

// 	case "individual":

// 		// TODO: revise this policy to be more restrictive

// 		// If the caller is us (the server operator), then we can read/write/update/delete
// 		if svc.isServerOperator(caller) {
// 			return "caller is server operator", nil
// 		}

// 		// If the caller is the Organization that is the mandator in the LEARCredential of the employee
// 		// then the caller can read/write/update/delete
// 		individualIdentificationArray := jpath.GetList(obj, "individualIdentification")

// 		// Look for an entry with 'identificationType=learcredentialemployee'
// 		for _, individualIdentification := range individualIdentificationArray {
// 			individualIdentificationMap, ok := individualIdentification.(map[string]any)
// 			if ok && individualIdentificationMap["identificationType"] == "learcredentialemployee" {
// 				// The 'issuingAuthority' must be equal to the caller organizationIdentifier
// 				issuingAuthority, ok := individualIdentificationMap["issuingAuthority"].(string)
// 				if ok && repo.SameOrganizations(issuingAuthority, caller.OrganizationIdentifier) {
// 					return fmt.Sprintf("caller %s is same as mandator in Individual object %s", caller.OrganizationIdentifier, issuingAuthority), nil
// 				}
// 			}
// 		}

// 		// We didn't find any valid entry in the Individual object
// 		return "", errl.Errorf("caller (%s) is neither the mandator in Individual object or the server operator", caller.OrganizationIdentifier)

// 	case "category":

// 		// If the caller is us (the server operator), then we can read/write/update/delete
// 		if svc.isServerOperator(caller) {
// 			return "caller is server operator", nil
// 		}

// 		return "", errl.Errorf("caller %s is not the server operator %s", caller.OrganizationIdentifier, svc.ServerOperatorDid)

// 	}

// 	// For any other object type:
// 	// If the request is a CREATE, we implement a fix for callers which do not set the Seller info in the incomingobject.
// 	// For other requests, the Seller info must be already set in the object.
// 	// Note that we do not do the same for the Buyer info, which is optional and may not exist.
// 	if objSeller == "" {
// 		if req.Action != CREATE {
// 			return "", errl.Errorf("seller info is not set in the object")
// 		}

// 		objSeller = caller.OrganizationIdentifier
// 		objSellerOperator = svc.ServerOperatorDid
// 		err := obj.SetSellerInfo(objSellerOperator, objSeller, "v4")
// 		if err != nil {
// 			return "", errl.Errorf("error trying to set seller info: %w", err)
// 		}
// 	}

// 	// There are several types of callers and requests, with the following logic.
// 	//
// 	// 1. If the caller is the Server Operator, it has access to all operations in all objects, but limited by the powers that the caller has.
// 	// 2. If the caller is a Marketplace Operator, it has access to all operations in all objects managed by the Marketplace, but limited by the powers that the caller has.
// 	// 3. If the caller is a normal organization, it has access to all operations in all objects it owns, but limited by the powers that the caller has.

// 	// Case 1: the caller is the Server Operator.
// 	// The caller can do (almost) anything, and it depends on the powers that the caller has:
// 	// - with power Onboard (it is a LEAR for the Server Operator), the caller can do anything and it is the Admin of the server.
// 	// - with power Product/(Create, Update, Delete), the user can do the relevant action in all objects.
// 	if svc.isServerOperator(caller) {

// 		// If caller is LEAR for the Server Operator, it can do anything to any object in this server
// 		if caller.IsLEAR {
// 			return fmt.Sprintf("caller %s is server operator %s and is a LEAR", caller.OrganizationIdentifier, svc.ServerOperatorDid), nil
// 		}

// 		// TODO: temporal fix because in ISBE the powers are malformed, and they will have to be fixed
// 		// // Reject if the caller does not have power to create products
// 		// if req.Action == CREATE && !caller.ProductCreatePower {
// 		// 	return false, errl.Errorf("caller %s is server operator but does not have power to create products", caller.OrganizationIdentifier)
// 		// }

// 		// If the caller is not LEAR, then check that it has the proper power for the action
// 		if req.Action == UPDATE && !caller.ProductUpdatePower {
// 			return "", errl.Errorf("caller %s is server operator but does not have power to update products", caller.OrganizationIdentifier)
// 		}
// 		if req.Action == DELETE && !caller.ProductDeletePower {
// 			return "", errl.Errorf("caller %s is server operator but does not have power to delete products", caller.OrganizationIdentifier)
// 		}

// 		// Accept if we reach here
// 		return fmt.Sprintf("caller %s is server operator and has power to operate the object", caller.OrganizationIdentifier), nil

// 	}

// 	// Case 2: the caller is a Marketplace Operator, that is, it is in our Trusted Parties list
// 	// it can only operate its own Sellers, and it must have the proper power for the action
// 	if svc.isTrustedParty(caller) {

// 		// Check that the object is owned by a Seller operated by the caller (Marketplace Operator).
// 		// In other words: the SellerOperator in the object is the same as the caller.
// 		if !repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier) {
// 			return "", errl.Errorf("Marketplace %s cannot modify object %s of Seller %s", caller.OrganizationIdentifier, obj.ID(), objSeller)
// 		}

// 		if caller.IsLEAR {
// 			return fmt.Sprintf("caller is LEAR of Marketplace %s", caller.OrganizationIdentifier), nil
// 		}

// 		// If the caller is not LEAR for the Marketplace, then check that it has the proper power for the action
// 		if req.Action == UPDATE && !caller.ProductUpdatePower {
// 			return "", errl.Errorf("caller %s is Marketplace Operator but does not have power to update products", caller.OrganizationIdentifier)
// 		}
// 		if req.Action == DELETE && !caller.ProductDeletePower {
// 			return "", errl.Errorf("caller %s is Marketplace Operator but does not have power to delete products", caller.OrganizationIdentifier)
// 		}

// 		// Accept if we reach here
// 		return fmt.Sprintf("caller %s is Marketplace Operator and has power to operate the object", caller.OrganizationIdentifier), nil

// 	}

// 	// Case 3: the caller is a "normal" organization.
// 	// It only has access to its own objects, but only if they are managed in this server.
// 	// In other words, the Seller in the object is the same as the caller, AND the SellerOperator in the object is the same as us (the ServerOperator).
// 	// In addition, we apply restrictions depending on the powers of the caller

// 	// The caller can not modify objects not managed in this server.
// 	// Reject if the SellerOperator of the object is not the same as the ServerOperator
// 	if !repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
// 		return "", errl.Errorf("only our CSPs are accepted")
// 	}

// 	// We first check the powers of the caller
// 	switch req.Action {
// 	case CREATE:
// 		if !caller.ProductCreatePower {
// 			return "", errl.Errorf("caller %s does not have power to create products", caller.OrganizationIdentifier)
// 		}
// 	case UPDATE:
// 		if !caller.ProductUpdatePower {
// 			return "", errl.Errorf("caller %s does not have power to update products", caller.OrganizationIdentifier)
// 		}
// 	case DELETE:
// 		if !caller.ProductDeletePower {
// 			return "", errl.Errorf("caller %s does not have power to delete products", caller.OrganizationIdentifier)
// 		}
// 	}

// 	// If there is Buyer info in the object, the object is a private object only for the Buyer and the Seller
// 	if objBuyerOperator != "" {
// 		// Reject if there is Buyer info in the object but the BuyerOperator is not us (the ServerOperator)
// 		if !repo.SameOrganizations(objBuyerOperator, svc.ServerOperatorDid) {
// 			return "", errl.Errorf("only our CSPs are accepted")
// 		}

// 		// Reject is the caller is not either the Seller or the Buyer
// 		if !repo.SameOrganizations(objSeller, caller.OrganizationIdentifier) && !repo.SameOrganizations(objBuyer, caller.OrganizationIdentifier) {
// 			return "", errl.Errorf("the caller %s is not the Seller %s or the Buyer %s", caller.OrganizationIdentifier, objSeller, objBuyer)
// 		}

// 	} else {
// 		// Reject if the caller is not the Seller
// 		if !repo.SameOrganizations(objSeller, caller.OrganizationIdentifier) {
// 			return "", errl.Errorf("the caller %s is not the Seller %s", caller.OrganizationIdentifier, objSeller)
// 		}
// 	}

// 	// Accept if we reach here
// 	return "the caller is the owner and we are the operator", nil

// }

// hardcodedPolicies is the main entry point for pre-PDP policy enforcement.
// It matches the top-down, operation-first structure and logic of hardcodedPolicies.md:
// 1. CREATE
// 2. READ / LIST
// 3. UPDATE
// 4. DELETE
func (svc *Service) hardcodedPolicies(req *Request, obj repo.TMFObjectMap) (reason string, err error) {

	// Attribute integrity check (Seller/Buyer info pairs must be consistent)
	if err := validateAttributeIntegrity(obj); err != nil {
		return "", err
	}

	// Dispatch by operation top-down
	switch req.Action {
	case ActionCREATE:
		return svc.evalCreatePolicy(req, obj)
	case ActionREAD, ActionLIST:
		return svc.evalReadListPolicy(req, obj)
	case ActionUPDATE:
		return svc.evalUpdatePolicy(req, obj)
	case ActionDELETE:
		return svc.evalDeletePolicy(req, obj)
	default:
		return "", errl.Errorf("unsupported action %s", req.Action)
	}
}

// -----------------------------------------------------------------------------
// Operation Policy Evaluators matching hardcodedPolicies.md
// -----------------------------------------------------------------------------

// Superuser Exception check for ServerOperator callers across all operations
func (svc *Service) checkServerOperatorAccess(req *Request) (bool, string, error) {
	caller := req.AuthUser
	if !svc.isServerOperator(caller) {
		return false, "", nil
	}

	if caller.IsLEAR {
		return true, fmt.Sprintf("caller %s is server operator and LEAR", caller.OrganizationIdentifier), nil
	}

	switch req.Action {
	case ActionREAD, ActionLIST:
		return true, fmt.Sprintf("caller %s is server operator (read/list)", caller.OrganizationIdentifier), nil
	case ActionCREATE:
		if svc.IsDOME() && !caller.ProductCreatePower {
			return false, "", errl.Errorf("caller %s is server operator but lacks create power", caller.OrganizationIdentifier)
		}
		return true, fmt.Sprintf("caller %s is server operator with create power", caller.OrganizationIdentifier), nil
	case ActionUPDATE:
		if !caller.ProductUpdatePower {
			return false, "", errl.Errorf("caller %s is server operator but lacks update power", caller.OrganizationIdentifier)
		}
		return true, fmt.Sprintf("caller %s is server operator with update power", caller.OrganizationIdentifier), nil
	case ActionDELETE:
		if !caller.ProductDeletePower {
			return false, "", errl.Errorf("caller %s is server operator but lacks delete power", caller.OrganizationIdentifier)
		}
		return true, fmt.Sprintf("caller %s is server operator with delete power", caller.OrganizationIdentifier), nil
	}
	return false, "", errl.Errorf("unsupported action %s", req.Action)
}

// CREATE policy implementation matching hardcodedPolicies.md
func (svc *Service) evalCreatePolicy(req *Request, obj repo.TMFObjectMap) (string, error) {
	caller := req.AuthUser
	if !caller.IsAuthenticated {
		return "", errl.Errorf("user not authenticated")
	}

	// Superuser Exception
	if isSO, reason, err := svc.checkServerOperatorAccess(req); isSO || err != nil {
		return reason, err
	}

	// Non-ServerOperator callers require ProductCreatePower
	if !caller.IsLEAR && !caller.ProductCreatePower {
		return "", errl.Errorf("caller %s lacks create power", caller.OrganizationIdentifier)
	}

	// Special Objects (Category, Organization, Individual)
	if isSpecialObject(obj) {
		switch strings.ToLower(obj.Type()) {
		case "category":
			return "", errl.Errorf("Category objects can only be created by server operator")
		case "organization":
			return "", errl.Errorf("Organization objects can only be created by server operator")
		case "individual":
			mandatorOrg := getIndividualIssuingAuthority(obj)
			if repo.SameOrganizations(mandatorOrg, caller.OrganizationIdentifier) {
				return fmt.Sprintf("caller %s is mandator in Individual object", caller.OrganizationIdentifier), nil
			}
			return "", errl.Errorf("caller %s is not mandator (%s) in Individual object", caller.OrganizationIdentifier, mandatorOrg)
		}
	}

	// On CREATE, auto-populate missing Seller/SellerOperator info
	if err := svc.ensureSellerInfoOnCreate(req, obj); err != nil {
		return "", err
	}

	objSeller, objSellerOperator, _ := obj.GetSellerInfo("")
	objBuyer, objBuyerOperator, _ := obj.GetBuyerInfo("")

	// Public objects: MUST be Seller or SellerOperator, and SellerOperator MUST be ServerOperator
	if obj.IsPotentiallyPublic() && objBuyer == "" {
		isSeller := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier)
		isSellerOp := repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier)
		if !isSeller && !isSellerOp {
			return "", errl.Errorf("caller %s is neither Seller %s nor SellerOperator %s for public object", caller.OrganizationIdentifier, objSeller, objSellerOperator)
		}
		if !repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
			return "", errl.Errorf("SellerOperator %s must be server operator %s", objSellerOperator, svc.ServerOperatorDid)
		}
		return "caller is Seller/SellerOperator and server operator matches", nil
	}

	// Private objects: MUST be Seller, SellerOperator, Buyer, or BuyerOperator
	isSeller := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier)
	isSellerOp := repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier)
	isBuyer := repo.SameOrganizations(objBuyer, caller.OrganizationIdentifier)
	isBuyerOp := repo.SameOrganizations(objBuyerOperator, caller.OrganizationIdentifier)

	if !isSeller && !isSellerOp && !isBuyer && !isBuyerOp {
		return "", errl.Errorf("caller %s is not party involved in private object (Seller=%s, Buyer=%s)", caller.OrganizationIdentifier, objSeller, objBuyer)
	}

	// Operator associated with caller's role MUST be ServerOperator
	if (isSeller || isSellerOp) && repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
		return "caller is Seller/SellerOperator on server operator node", nil
	}
	if (isBuyer || isBuyerOp) && repo.SameOrganizations(objBuyerOperator, svc.ServerOperatorDid) {
		return "caller is Buyer/BuyerOperator on server operator node", nil
	}

	return "", errl.Errorf("operator for caller's role does not match server operator %s", svc.ServerOperatorDid)
}

// READ / LIST policy implementation matching hardcodedPolicies.md
func (svc *Service) evalReadListPolicy(req *Request, obj repo.TMFObjectMap) (string, error) {
	caller := req.AuthUser

	// Special objects READ/LIST rules
	if isSpecialObject(obj) {
		switch strings.ToLower(obj.Type()) {
		case "category", "organization":
			// Public object types: readable by any caller (auth or unauth)
			return "public special object (category/organization)", nil
		case "individual":
			// Private object type: requires authentication & org match
			if !caller.IsAuthenticated {
				return "", errl.Errorf("user not authenticated")
			}
			if svc.isServerOperator(caller) {
				return "caller is server operator", nil
			}
			mandatorOrg := getIndividualIssuingAuthority(obj)
			if repo.SameOrganizations(mandatorOrg, caller.OrganizationIdentifier) {
				return fmt.Sprintf("caller %s is mandator in Individual object", caller.OrganizationIdentifier), nil
			}
			return "", errl.Errorf("caller %s is not mandator in Individual object", caller.OrganizationIdentifier)
		}
	}

	// Public objects: readable by any caller (auth or unauth)
	objBuyer, _, _ := obj.GetBuyerInfo("")
	if obj.IsPotentiallyPublic() && objBuyer == "" {
		return "public resource without buyer info", nil
	}

	// Private objects require authentication
	if !caller.IsAuthenticated {
		return "", errl.Errorf("user not authenticated")
	}

	// Superuser Exception
	if isSO, reason, err := svc.checkServerOperatorAccess(req); isSO || err != nil {
		return reason, err
	}

	// Private objects: caller's organization must match Seller, SellerOperator, Buyer, or BuyerOperator
	objSeller, objSellerOperator, _ := obj.GetSellerInfo("")
	_, objBuyerOperator, _ := obj.GetBuyerInfo("")

	isParty := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier) ||
		repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier) ||
		repo.SameOrganizations(objBuyer, caller.OrganizationIdentifier) ||
		repo.SameOrganizations(objBuyerOperator, caller.OrganizationIdentifier)

	if isParty {
		return "caller is party involved in private object", nil
	}

	return "", errl.Errorf("caller %s is not party involved in private object", caller.OrganizationIdentifier)
}

// UPDATE policy implementation matching hardcodedPolicies.md
func (svc *Service) evalUpdatePolicy(req *Request, obj repo.TMFObjectMap) (string, error) {
	caller := req.AuthUser
	if !caller.IsAuthenticated {
		return "", errl.Errorf("user not authenticated")
	}

	// Superuser Exception
	if isSO, reason, err := svc.checkServerOperatorAccess(req); isSO || err != nil {
		return reason, err
	}

	// Non-ServerOperator callers require ProductUpdatePower
	if !caller.IsLEAR && !caller.ProductUpdatePower {
		return "", errl.Errorf("caller %s lacks update power", caller.OrganizationIdentifier)
	}

	// Special Objects
	if isSpecialObject(obj) {
		switch strings.ToLower(obj.Type()) {
		case "category":
			return "", errl.Errorf("Category objects can only be updated by server operator")
		case "organization":
			return "", errl.Errorf("Organization objects can only be updated by server operator")
		case "individual":
			mandatorOrg := getIndividualIssuingAuthority(obj)
			if repo.SameOrganizations(mandatorOrg, caller.OrganizationIdentifier) {
				return fmt.Sprintf("caller %s is mandator in Individual object", caller.OrganizationIdentifier), nil
			}
			return "", errl.Errorf("caller %s is not mandator in Individual object", caller.OrganizationIdentifier)
		}
	}

	objSeller, objSellerOperator, _ := obj.GetSellerInfo("")
	objBuyer, objBuyerOperator, _ := obj.GetBuyerInfo("")

	// Public objects: MUST be Seller or SellerOperator, and SellerOperator MUST be ServerOperator
	if obj.IsPotentiallyPublic() && objBuyer == "" {
		isSeller := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier)
		isSellerOp := repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier)
		if !isSeller && !isSellerOp {
			return "", errl.Errorf("caller %s is neither Seller %s nor SellerOperator %s for public object", caller.OrganizationIdentifier, objSeller, objSellerOperator)
		}
		if !repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
			return "", errl.Errorf("SellerOperator %s must be server operator %s", objSellerOperator, svc.ServerOperatorDid)
		}
		return "caller is Seller/SellerOperator and server operator matches", nil
	}

	// Private objects: MUST be Seller, SellerOperator, Buyer, or BuyerOperator
	isSeller := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier)
	isSellerOp := repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier)
	isBuyer := repo.SameOrganizations(objBuyer, caller.OrganizationIdentifier)
	isBuyerOp := repo.SameOrganizations(objBuyerOperator, caller.OrganizationIdentifier)

	if !isSeller && !isSellerOp && !isBuyer && !isBuyerOp {
		return "", errl.Errorf("caller %s is not party involved in private object", caller.OrganizationIdentifier)
	}

	if (isSeller || isSellerOp) && repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
		return "caller is Seller/SellerOperator on server operator node", nil
	}
	if (isBuyer || isBuyerOp) && repo.SameOrganizations(objBuyerOperator, svc.ServerOperatorDid) {
		return "caller is Buyer/BuyerOperator on server operator node", nil
	}

	return "", errl.Errorf("operator for caller's role does not match server operator %s", svc.ServerOperatorDid)
}

// DELETE policy implementation matching hardcodedPolicies.md
func (svc *Service) evalDeletePolicy(req *Request, obj repo.TMFObjectMap) (string, error) {
	caller := req.AuthUser
	if !caller.IsAuthenticated {
		return "", errl.Errorf("user not authenticated")
	}

	// Superuser Exception
	if isSO, reason, err := svc.checkServerOperatorAccess(req); isSO || err != nil {
		return reason, err
	}

	// Non-ServerOperator callers require ProductDeletePower
	if !caller.IsLEAR && !caller.ProductDeletePower {
		return "", errl.Errorf("caller %s lacks delete power", caller.OrganizationIdentifier)
	}

	// Special Objects
	if isSpecialObject(obj) {
		switch strings.ToLower(obj.Type()) {
		case "category":
			return "", errl.Errorf("Category objects can only be deleted by server operator")
		case "organization":
			return "", errl.Errorf("Organization objects can only be deleted by server operator")
		case "individual":
			mandatorOrg := getIndividualIssuingAuthority(obj)
			if repo.SameOrganizations(mandatorOrg, caller.OrganizationIdentifier) {
				return fmt.Sprintf("caller %s is mandator in Individual object", caller.OrganizationIdentifier), nil
			}
			return "", errl.Errorf("caller %s is not mandator in Individual object", caller.OrganizationIdentifier)
		}
	}

	objSeller, objSellerOperator, _ := obj.GetSellerInfo("")
	objBuyer, objBuyerOperator, _ := obj.GetBuyerInfo("")

	// Public objects: MUST be Seller or SellerOperator, and SellerOperator MUST be ServerOperator
	if obj.IsPotentiallyPublic() && objBuyer == "" {
		isSeller := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier)
		isSellerOp := repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier)
		if !isSeller && !isSellerOp {
			return "", errl.Errorf("caller %s is neither Seller %s nor SellerOperator %s for public object", caller.OrganizationIdentifier, objSeller, objSellerOperator)
		}
		if !repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
			return "", errl.Errorf("SellerOperator %s must be server operator %s", objSellerOperator, svc.ServerOperatorDid)
		}
		return "caller is Seller/SellerOperator and server operator matches", nil
	}

	// Private objects: MUST be Seller, SellerOperator, or BuyerOperator (Buyer directly is REJECTED)
	isSeller := repo.SameOrganizations(objSeller, caller.OrganizationIdentifier)
	isSellerOp := repo.SameOrganizations(objSellerOperator, caller.OrganizationIdentifier)
	isBuyerOp := repo.SameOrganizations(objBuyerOperator, caller.OrganizationIdentifier)

	if !isSeller && !isSellerOp && !isBuyerOp {
		return "", errl.Errorf("caller %s is not Seller, SellerOperator, or BuyerOperator for private object delete", caller.OrganizationIdentifier)
	}

	if (isSeller || isSellerOp) && repo.SameOrganizations(objSellerOperator, svc.ServerOperatorDid) {
		return "caller is Seller/SellerOperator on server operator node", nil
	}
	if isBuyerOp && repo.SameOrganizations(objBuyerOperator, svc.ServerOperatorDid) {
		return "caller is BuyerOperator on server operator node", nil
	}

	return "", errl.Errorf("operator for caller's role does not match server operator %s", svc.ServerOperatorDid)
}

// -----------------------------------------------------------------------------
// Pipeline Step Helpers
// -----------------------------------------------------------------------------

// Step 1: Ensure Seller/SellerOperator and Buyer/BuyerOperator pairs are not partially set.
func validateAttributeIntegrity(obj repo.TMFObjectMap) error {
	objSeller, objSellerOperator, _ := obj.GetSellerInfo("")
	objBuyer, objBuyerOperator, _ := obj.GetBuyerInfo("")

	if isPartiallySet(objSeller, objSellerOperator) {
		return errl.Errorf("objSeller and objSellerOperator must both be set or both be empty, got objSeller='%s', objSellerOperator='%s'", objSeller, objSellerOperator)
	}
	if isPartiallySet(objBuyer, objBuyerOperator) {
		return errl.Errorf("objBuyer and objBuyerOperator must both be set or both be empty, got objBuyer='%s', objBuyerOperator='%s'", objBuyer, objBuyerOperator)
	}
	return nil
}

// Helper to check if object is a special type (organization, individual, category).
func isSpecialObject(obj repo.TMFObjectMap) bool {
	objType := obj.Type()
	switch strings.ToLower(objType) {
	case "organization", "individual", "category":
		return true
	default:
		return false
	}
}

// Helper to extract issuingAuthority from Individual object.
func getIndividualIssuingAuthority(obj repo.TMFObjectMap) string {
	for _, item := range jpath.GetList(obj, "individualIdentification") {
		m, ok := item.(map[string]any)
		if ok && m["identificationType"] == "learcredentialemployee" {
			issuingAuth, ok := m["issuingAuthority"].(string)
			if ok {
				return issuingAuth
			}
		}
	}
	return ""
}

// Auto-assign Seller info if omitted on CREATE.
func (svc *Service) ensureSellerInfoOnCreate(req *Request, obj repo.TMFObjectMap) error {
	if isSpecialObject(obj) {
		return nil
	}
	objSeller, _, _ := obj.GetSellerInfo("")
	if objSeller != "" {
		return nil
	}
	caller := req.AuthUser
	if err := obj.SetSellerInfo(svc.ServerOperatorDid, caller.OrganizationIdentifier, "v4"); err != nil {
		return errl.Errorf("error setting seller info: %w", err)
	}
	return nil
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
	ruleEngine Authorizer,
	req *Request,
	tokenClaims map[string]any,
	objectMap repo.TMFObjectMap,
) (err error) {

	userArgument := pdp.StarTMFMap(req.AuthUser.ToMap())
	incomingObjectArgument := pdp.StarTMFMap(objectMap)
	requestArgument := pdp.StarTMFMap(req.ToMap())
	tokenArgument := pdp.StarTMFMap(tokenClaims)

	// Assemble all data in a single input argument, to the style of OPA.
	input := map[string]any{
		"request": requestArgument,
		"token":   tokenArgument,
		"tmf":     incomingObjectArgument,
		"user":    userArgument,
	}

	decision := true
	if ruleEngine != nil {
		decision, err = ruleEngine.Authorize(input)

		// An error is considered a rejection
		if err != nil {
			return errl.Errorf("rules engine rejected request due to an error: %w", err)
		}
	}

	// The rules engine rejected the request
	if !decision {
		return errl.Errorf("rules engine rejected request due to policy")
	}

	// The rules engine accepted the request
	slog.Info("rules engine accepted request")
	return nil
}

// isPartiallySet returns true if exactly one of the fields is empty.
func isPartiallySet(a, b string) bool {
	return (a == "") != (b == "")
}
