package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"log/slog"

	"github.com/google/uuid"
	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/jpath"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// requiresAuthentication checks if the user is authenticated in the request.
func (svc *Service) requiresAuthentication(req *Request) *Response {
	if !req.AuthUser.IsAuthenticated {
		return ErrorResponsef(http.StatusUnauthorized, "user not authenticated")
	}
	return nil
}

// parseRequestBody parses the JSON body from the request into a TMFObjectMap.
func (svc *Service) parseRequestBody(req *Request) (repo.TMFObjectMap, *Response) {
	incomingObjectMap, err := repo.NewTMFObjectMapFromBytes(req.ResourceName, req.Body)
	if err != nil {
		return nil, ErrorResponsef(http.StatusBadRequest, "failed to bind request body: %w", errl.Error(err))
	}
	return incomingObjectMap, nil
}

// authorizeAction calls the PDP to check if the user is authorized for the current request.
func (svc *Service) authorizeAction(req *Request, obj repo.TMFObjectMap) *Response {
	if authorized, err := svc.takeDecision(svc.ruleEngine, req, obj); !authorized {
		return ErrorResponsef(http.StatusForbidden,
			"user %s is not authorized, object: %s, error: %w",
			req.AuthUser.OrganizationIdentifier,
			obj,
			err,
		)
	}
	slog.Debug("caller is authorized", "action", req.Action, "resource", req.ResourceName)
	return nil
}

// applyAttributeSelection filters the fields of a TMFObjectMap based on the fieldSet.
func (svc *Service) applyAttributeSelection(obj repo.TMFObjectMap, fieldSet map[string]bool) repo.TMFObjectMap {
	if len(fieldSet) == 0 {
		return obj
	}

	filteredObjectMap := make(map[string]any, 10)
	for key := range fieldSet {
		if val, ok := obj[key]; ok {
			filteredObjectMap[key] = val
		}
	}
	return filteredObjectMap
}

// parsePaginationParams extracts limit and offset from query parameters with defaults limit=10, offset=0
func (svc *Service) parsePaginationParams(req *Request) (limit int, offset int) {
	limit = 10
	offset = 0

	if val := req.QueryParams.Get("limit"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			limit = i
		}
	}

	if val := req.QueryParams.Get("offset"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			offset = i
		}
	}
	return
}

// parseFieldsParam parses the 'fields' query parameter into a field set.
func (svc *Service) parseFieldsParam(fieldsParam string) map[string]bool {
	if fieldsParam == "" {
		return nil
	}

	// Even if the user does not specify any fields, we need to return the following fields
	fieldSet := map[string]bool{
		"id":              true,
		"href":            true,
		"lastUpdate":      true,
		"version":         true,
		"@type":           true,
		"lifecycleStatus": true,
	}

	// Add the user specified fields
	if fieldsParam != "none" {
		for f := range strings.SplitSeq(fieldsParam, ",") {
			fieldSet[strings.TrimSpace(f)] = true
		}
	}
	return fieldSet
}

// ensureCreateMetadata handles the generation and validation of TMF metadata fields.
func (svc *Service) ensureCreateMetadata(req *Request, obj repo.TMFObjectMap) *Response {
	// Check for the different required name fields depending on the object type
	if obj.IsIndividual() {
		if givenName, _ := obj["givenName"].(string); givenName == "" {
			return ErrorResponsef(http.StatusBadRequest, "givenName is required in individual object")
		}
		if familyName, _ := obj["familyName"].(string); familyName == "" {
			return ErrorResponsef(http.StatusBadRequest, "familyName is required in individual object")
		}
	} else if obj.IsOrganization() {
		if tradingName, _ := obj["tradingName"].(string); tradingName == "" {
			return ErrorResponsef(http.StatusBadRequest, "tradingName is required in organization object")
		}
	} else if name, _ := obj["name"].(string); name == "" {
		return ErrorResponsef(http.StatusBadRequest, "name is required")
	}

	id := obj.ID()
	version := obj.Version()

	// If the incoming object specifies an 'id', this is only possible if it creates a new version.
	if id != "" && version == "" {
		return ErrorResponsef(http.StatusBadRequest, "id specified but version is missing")
	}

	if svc.Features.GenerateIDOnCreate {
		if id == "" {
			if obj.IsOrganization() {
				orgList := jpath.GetList(obj, "organizationIdentification")
				if len(orgList) == 0 {
					return ErrorResponsef(http.StatusBadRequest, "organizationIdentification is required in organization object")
				}
				org := jpath.GetString(orgList[0], "identificationId")
				if org == "" {
					return ErrorResponsef(http.StatusBadRequest, "organizationIdentification[0].identificationId is required")
				}
				id = fmt.Sprintf("urn:ngsi-ld:organization:%s", org)
			} else {
				id = fmt.Sprintf("urn:ngsi-ld:%s:%s", ToKebabCase(req.ResourceName), uuid.NewString())
			}
			obj.SetID(id)

			if version == "" {
				version = "0.1"
				obj.SetVersion(version)
			}
		}

		obj.SetHref(id)

		if obj.LastUpdate() == "" {
			obj.SetLastUpdateNow()
		}
	}

	// Add Seller and SellerOperator if missing (CREATE only)
	if req.Action == CREATE {
		objSeller, objSellerOperator, _ := obj.GetSellerInfo("v4")
		if objSeller == "" && objSellerOperator == "" {
			// Set default seller info from caller and server operator
			if err := obj.SetSellerInfo(svc.ServerOperatorDid, req.AuthUser.OrganizationIdentifier, "v4"); err != nil {
				return ErrorResponsef(http.StatusInternalServerError, "failed to set default seller info: %w", err)
			}
		}
	}

	// If the object requires a lifecycleStatus, add it if not specified by the caller
	if baseStatus, ok := LifecycleStatusMandatory[req.ResourceName]; ok {
		if lifecycleStatus := obj.LifecycleStatus(); lifecycleStatus == "" {
			obj.SetLifecycleStatus(baseStatus)
		}
	}

	// Set the @type if not specified
	if resourceType := obj.Type(); resourceType == "" {
		resourceType = strings.ToUpper(req.ResourceName[0:1]) + req.ResourceName[1:]
		obj.SetType(resourceType)
	}

	return nil
}

// ensureUpdateMetadata handles the validation and updates of metadata fields during an update operation.
func (svc *Service) ensureUpdateMetadata(req *Request, incomingObjMap repo.TMFObjectMap) *Response {
	// Check if the caller is trying to set the lifecycleStatus of a ProductOffering to "Launched"
	if strings.EqualFold(incomingObjMap.Type(), config.ProductOffering) && strings.EqualFold(incomingObjMap.LifecycleStatus(), "Launched") {
		if svc.Features.OfferingLaunchOnlyByAdmin {
			caller := req.AuthUser
			if !repo.SameOrganizations(caller.OrganizationIdentifier, svc.ServerOperatorDid) {
				return ErrorResponsef(http.StatusForbidden, "offering launch is only allowed by admin")
			}
		}
	}

	// But if the 'id' is present in the body, ensure it matches the 'id' in the URL
	if !svc.Features.AllowIDInBody {
		id, _ := incomingObjMap["id"].(string)
		if id != "" && id != req.ID {
			err := errl.Errorf("ID in body must match ID in URL")
			return ErrorResponsef(http.StatusBadRequest,
				"invalid object, request id: %s, id in body: %s, error: %w",
				req.ID,
				id,
				err,
			)
		}
	}

	// Set the lastUpdate property. We overwrite whatever the user set.
	incomingObjMap.SetLastUpdateNow()
	return nil
}

// mergeRFC7396 implements JSON Merge Patch (RFC 7396) to merge a patch object into a target object.
func (svc *Service) mergeRFC7396(target, patch map[string]any) {
	for k, v := range patch {
		if v == nil {
			delete(target, k)
			continue
		}

		vMap, vIsMap := v.(map[string]any)
		if vIsMap {
			if existingChild, ok := target[k]; ok {
				if existingChildMap, ok2 := existingChild.(map[string]any); ok2 {
					svc.mergeRFC7396(existingChildMap, vMap)
					target[k] = existingChildMap
					continue
				}
			}
			target[k] = vMap
			continue
		}
		target[k] = v
	}
}
