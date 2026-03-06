package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// CreateGenericObject creates a new TMF object using generalized parameters.
//
// The function performs the following checks and operations:
//
// 1.  **Authentication**:
//   - It requires an authenticated user, obtained by processing the Access Token.
//   - If the user is not authenticated, it returns a 401 Unauthorized error.
//
// 2.  **Incoming TMF Object Checks**:
//   - It parses the request body to get the incoming TMF object.
//   - It checks for mandatory name fields based on the resource type (`givenName` and `familyName` for `Individual`, `tradingName` for `Organization`, and `name` for others).
//   - If an `id` is provided in the incoming object, it ensures that a `version` is also present.
//   - If no `id` is provided, a new one is generated in the format "urn:ngsi-ld:{resource-in-kebab-case}:{uuid}".
//   - It sets a default `version` to "0.1" if not provided.
//   - It overwrites the `href` field to ensure it is correct and consistent with the 'id'.
//   - It adds a default `lifecycleStatus` if the resource type requires it and it's not present.
//   - It sets the `@type` field if it's not specified.
//   - It sets the `seller` and `buyer` information, overwriting any existing values.
//
// 3.  **Authorization**:
//   - It calls the Policy Decision Point (PDP) to check if the user is authorized to create the object.
//   - If the user is not authorized, it returns a 403 Forbidden error.
//
// 4.  **Object Creation**:
//   - If proxy mode is enabled, it forwards the request to the remote TMF server.
//   - It checks the remote server's response. If the remote server does not provide a `lastUpdate` field, it adds one and logs a warning.
//   - It stores the object returned by the remote server in the local database.
//   - If proxy mode is disabled, it creates the object directly in the local database.
//   - It sets the `lastUpdate` field to the current time.
//
// 5.  **Response and Notification**:
//   - It returns a 201 Created response with the created object in the body.
//   - It sends a "CreateEvent" notification to subscribed listeners.
func (svc *Service) CreateGenericObject(req *Request) *Response {
	var err error
	slog.Debug("CreateGenericObject called", slog.String("apiFamily", req.APIfamily), slog.String("resourceName", req.ResourceName))

	// ************************************************************************************************
	// Authentication: we require the user to be authenticated
	// ************************************************************************************************
	if len(req.TokenMap) == 0 {
		return ErrorResponsef(http.StatusUnauthorized, "user not authenticated")
	}

	var incomingObjectMap repo.TMFObjectMap
	var id string
	var version string

	// ************************************************************************************************
	// Parse the request body, which contains the TMForum object being created.
	// We perform some formal verifications and add default values if needed:
	// - Check for compulsory 'name' field (or equivalent fields for Individual or Organization)
	// - If the incoming object specifies an 'id', this is only possible if it creates a new version.
	//     That means the incoming object must include a 'version' property.
	// - Create a new 'id' if the user did not specify it
	// - Overwrite href in case it is specified by the caller.
	// - Set a proper lifecycleStatus, if it was not specified by the caller
	// - Set the '@type' based on the 'resource' name, if not specified by the user
	// - Add Seller and Buyer info. We overwrite whatever is in the incoming object.
	//     At this moment we set the Seller to the caller, and SellerOperator to us (the server operator).
	//     This is good when we are a single server, but if we replicate with other marketplaces, we need
	//     to have a trusted list of marketplaces. It the request comes from one of them, we trust it for
	//     the content of the Seller object, as long as the SellerOperator field is one of the marketplaces.
	//     In other words: if the organizationIdentifier of the caller is the same as the SellerOperator atribute,
	//     and it is different from our identifier, then we do not check the Seller attribute.
	// TODO: revise the Seller and SellerOperator logic when the trusted list of marketplaces is implemented
	// ************************************************************************************************
	{

		incomingObjectMap, err = repo.NewTMFObjectMapFromBytes(req.ResourceName, req.Body)
		if err != nil {
			return ErrorResponsef(http.StatusBadRequest, "failed to bind request body: %w", errl.Error(err))
		}

		// Check for the different required name fields depending on the object type
		if incomingObjectMap.IsIndividual() {
			// Individual requires givenName and familyName
			if givenName, _ := incomingObjectMap["givenName"].(string); givenName == "" {
				return ErrorResponsef(http.StatusBadRequest, "givenName is required in individual object: %s", incomingObjectMap)
			}
			if familyName, _ := incomingObjectMap["familyName"].(string); familyName == "" {
				return ErrorResponsef(http.StatusBadRequest, "familyName is required in individual object: %s", incomingObjectMap)
			}
		} else if incomingObjectMap.IsOrganization() {
			// Organization requires tradingName
			if tradingName, _ := incomingObjectMap["tradingName"].(string); tradingName == "" {
				return ErrorResponsef(http.StatusBadRequest, "tradingName is required in organization object: %s", incomingObjectMap)
			}
		} else if name, _ := incomingObjectMap["name"].(string); name == "" {
			// Other objects require name
			return ErrorResponsef(http.StatusBadRequest, "name is required: %s", incomingObjectMap)
		}

		id = incomingObjectMap.ID()
		version = incomingObjectMap.Version()

		// If the incoming object specifies an 'id', this is only possible if it creates a new version.
		// That means the incoming object must have a 'version' property.
		if id != "" && version == "" {
			return ErrorResponsef(http.StatusBadRequest, "id specified but version is missing: %s", incomingObjectMap)
		}

		// Create a new 'id' if the user did not specify it
		if id == "" {
			// DOME does not support using an 'id' in the incoming object.
			if !svc.proxyEnabled || !DOMEHacks {
				// If the incoming object does not have an 'id', we generate a new one
				// The format is "urn:ngsi-ld:{resource-in-kebab-case}:{uuid}"
				id = fmt.Sprintf("urn:ngsi-ld:%s:%s:%s", ToKebabCase(req.ResourceName), req.AuthUser.OrganizationIdentifier, uuid.NewString())
				incomingObjectMap.SetID(id)
				slog.Debug("Generated new ID for object", "id", id)

				// Set default 'version' if not provided by the user
				if version == "" {
					version = "0.1"
					incomingObjectMap.SetVersion(version)
					slog.Debug("Set default version", slog.String("version", version))
				}
			}

		}

		// Overwrite href in case it is specified by the caller.
		// We could be more strict and reject the request, but this is more permissive.
		if !svc.proxyEnabled || !DOMEHacks {
			incomingObjectMap.SetHref(fmt.Sprintf("/tmf-api/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, id))
			slog.Debug("Set href", slog.String("href", incomingObjectMap["href"].(string)))
		}

		// Set the lastUpdate property if the user did not specify one
		if incomingObjectMap.LastUpdate() == "" {
			incomingObjectMap.SetLastUpdateNow()
			slog.Debug("Set lastUpdate", slog.String("lastUpdate", incomingObjectMap["lastUpdate"].(string)))
		}

		// If the object requires a lifecycleStatus, add it if not specified by the caller
		if baseStatus, ok := LifecycleStatusMandatory[req.ResourceName]; ok {
			if lifecycleStatus := incomingObjectMap.LifecycleStatus(); lifecycleStatus == "" {
				slog.Debug("adding lifecycleStatus", slog.String("status", baseStatus))
				incomingObjectMap.SetLifecycleStatus(baseStatus)
			}
		}

		// Set the @type if not specified by the user
		if resourceType := incomingObjectMap.Type(); resourceType == "" {
			resourceType = strings.ToUpper(req.ResourceName[0:1]) + req.ResourceName[1:]
			incomingObjectMap.SetType(resourceType)
		}

	}

	// ************************************************************************************************
	// Before performing the action, check if the user can perform the operation on the object,
	// based on the rules defined by the user in the policy engine.
	// ************************************************************************************************

	if authorized, err := svc.takeDecision(svc.ruleEngine, req, incomingObjectMap); !authorized {
		return ErrorResponsef(http.StatusForbidden,
			"user %s is not authorized, object: %s, error: %w",
			req.AuthUser.OrganizationIdentifier,
			incomingObjectMap,
			err,
		)
	} else {
		slog.Debug("caller is authorized to create object", "reason", err)
	}

	// ************************************************************************************************
	// Now we can proceed, creating an object in the database or the remote server in proxy mode
	// ************************************************************************************************

	obj := incomingObjectMap.ToTMFRecord(req.ResourceName)

	respSt := svc.createLocalOrRemoteObject(req, obj)

	if respSt.StatusCode == http.StatusCreated {
		// Send TMForum notification
		eventType := toEventType(req.ResourceName, "CreateEvent")
		eventPayload := buildEventPayload(req, eventType, respSt.Body)
		svc.notif.PublishEvent(req.APIfamily, eventType, eventPayload)
	}

	// We respond with the created object with all the properties, including the default ones.
	// TODO: not critical, but we could send only the mandatory attributes and those requested by the caller.
	// Sending the whole object is compliant with the TMForum specification,
	// but sending only the mandatory attributes is more efficient.
	return respSt
}
