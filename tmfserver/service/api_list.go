package service

import (
	"log/slog"
	"net/http"

	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
)

// ListGenericObjects retrieves all TMF objects of a given type.
func (svc *Service) ListGenericObjects(req *Request) *Response {
	if !req.HealthRequest {
		slog.Debug("ListGenericObjects called", slog.String("resourceName", req.ResourceName))
	}

	// Parse pagination parameters
	userLimit, userOffset := svc.parsePaginationParams(req)

	// If the user specified explicitly limit=0, return an empty list. This may be used to test the API without returning all the objects.
	if userLimit == 0 {
		return &Response{StatusCode: http.StatusOK, Body: []repo.TMFObjectMap{}}
	}

	// Parse field selection parameters, which are the fields to be returned in the response.
	fieldsParam := req.QueryParams.Get("fields")
	fieldSet := svc.parseFieldsParam(fieldsParam)

	var responseData []repo.TMFObjectMap
	var responseHeaders map[string]string
	var resp *Response

	// Retrieve objects (Remote or Local)
	if svc.proxyEnabled {
		responseData, responseHeaders, resp = svc.listRemoteObjects(req, userLimit, userOffset, fieldSet)
		if resp != nil {
			return resp
		}
	} else {
		responseData, responseHeaders, resp = svc.listLocalObjects(req, userLimit, userOffset, fieldSet)
		if resp != nil {
			return resp
		}
	}

	return &Response{
		StatusCode: http.StatusOK,
		Headers:    responseHeaders,
		Body:       responseData,
	}
}
