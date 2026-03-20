package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/mattn/go-sqlite3"
)

// ErrObjectExists is returned when trying to create an object that already exists.
type ErrObjectExists struct {
	ID   string
	Type string
}

func (e *ErrObjectExists) Error() string {
	return fmt.Sprintf("object with id %s and type %s already exists", e.ID, e.Type)
}

func (e *ErrObjectExists) Is(target error) bool {
	switch target.(type) {
	case *ErrObjectExists:
		return true
	default:
		return false
	}
}

// ErrObjectConflict is returned when trying to update an object with a version that is not the latest.
type ErrObjectConflict struct {
	ID      string
	Type    string
	Version string
}

func (e *ErrObjectConflict) Error() string {
	return fmt.Sprintf("conflict updating object with id %s and type %s. Version %s is not the latest", e.ID, e.Type, e.Version)
}

// createObject creates a new TMF object. Returns &ErrObjectExists if the object already existed.
func (svc *Service) createObject(obj *repo.TMFRecord) error {
	if obj == nil {
		return errl.Errorf("object is nil")
	}
	slog.Debug("dbLayer: createObject", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.CreateObject(obj)
	}

	// Make sure timestamps are correct
	now := time.Now()
	obj.CreatedAt = now.Unix()
	obj.UpdatedAt = now.Unix()

	// Execute the SQL
	_, err := svc.db.NamedExec(`INSERT INTO tmf_object
		(id, type, version, api_version, seller, buyer, last_update, content, created_at, updated_at)
		VALUES (:id, :type, :version, :api_version, :seller, :buyer, :last_update, :content, :created_at, :updated_at)`,
		obj,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return &ErrObjectExists{ID: obj.ID, Type: obj.Type}
			}
		}
		err = errl.Errorf("failed to create object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}
	return err
}

func (svc *Service) createObjectIfNotExist(obj *repo.TMFRecord) error {
	slog.Debug("dbLayer: createObjectIfNotExist", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.CreateObject(obj)
	}

	// Make sure timestamps are correct
	now := time.Now()
	obj.CreatedAt = now.Unix()
	obj.UpdatedAt = now.Unix()

	// Execute the SQL
	_, err := svc.db.NamedExec(`INSERT INTO tmf_object
		(id, type, version, api_version, seller, buyer, last_update, content, created_at, updated_at)
		VALUES (:id, :type, :version, :api_version, :seller, :buyer, :last_update, :content, :created_at, :updated_at)
		ON CONFLICT DO NOTHING`,
		obj,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return &ErrObjectExists{ID: obj.ID, Type: obj.Type}
			}
		}
		err = errl.Errorf("failed to create object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}
	return err
}

func (svc *Service) upsertObject(obj *repo.TMFRecord) error {
	slog.Debug("dbLayer: upsertObject", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.CreateObject(obj)
	}

	// Make sure timestamps are correct
	now := time.Now()
	obj.CreatedAt = now.Unix()
	obj.UpdatedAt = now.Unix()

	// Execute the SQL
	_, err := svc.db.NamedExec(`INSERT INTO tmf_object 
		(id, type, version, api_version, seller, buyer, last_update, content, created_at, updated_at)
		VALUES (:id, :type, :version, :api_version, :seller, :buyer, :last_update, :content, :created_at, :updated_at)
		ON CONFLICT DO
		UPDATE SET version = :version, seller = :seller, buyer = :buyer, last_update = :last_update, content = :content, updated_at = :updated_at`,
		obj,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return &ErrObjectExists{ID: obj.ID, Type: obj.Type}
			}
		}
		err = errl.Errorf("failed to upsert object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}
	return err
}

// createLocalOrRemoteObject creates an object in the remote server and then in the local database, if the proxy is enabled.
// Othewise, it just creates the object in the local database.
func (svc *Service) createLocalOrRemoteObject(req *Request, obj *repo.TMFRecord) *Response {

	objMap, err := obj.ToTMFObjectMapCreate()
	if err != nil {
		err = errl.Errorf("failed to marshal object: %w", err)
		return ErrorResponsef(http.StatusInternalServerError, "failed to marshal object: %w", err)
	}

	// Create the object only in the local database if the proxy is not enabled
	if !svc.proxyEnabled {
		if err := svc.createObject(obj); err != nil {
			if errors.Is(err, &ErrObjectExists{}) {
				return ErrorResponsef(http.StatusBadRequest, "object %s already exists: %w", obj.ID, err)
			} else {
				return ErrorResponsef(http.StatusInternalServerError, "failed to create object locally: %w", err)
			}
		}

		// Set the headers of the reply to the caller, as per TMF specs
		headers := map[string]string{
			"Location": objMap.Href(),
		}
		slog.Info("Object created successfully", slog.String("id", objMap.ID()), slog.String("resourceName", req.ResourceName), slog.String("location", objMap.Href()))

		return &Response{StatusCode: http.StatusCreated, Headers: headers, Body: objMap}

	}

	// With the proxy enabled, first create the object in the remote server and then locally
	// We do not have to worry about transaction integrity, because if the remote server fails, we do not create the object locally
	// If the local server fails, we do not have to do anything, because the object is already in the remote server and our cache will
	// be updated later on the next call from the user

	remoteObjectMap, errs := ForwardTMFPost(req, svc.RemoteTMFServer, objMap)
	if len(errs) > 0 {
		return ErrorResponsef(http.StatusInternalServerError, "failed to proxy request: %w", errs[0])
	}

	// Prepare the object for the database
	lastUpdate := remoteObjectMap.LastUpdate()

	// It is an error if the remote server does not return a 'lastUpdate', but we just log a warning
	if lastUpdate == "" {
		slog.Warn("remote server did not return lastUpdate, fixing it", slog.String("id", remoteObjectMap.ID()))
	}

	// Create the new object in the local database
	if err := svc.createObject(remoteObjectMap.ToTMFRecord(req.ResourceName)); err != nil {
		// If we get an error, just log the error because we return the object created remotely
		slog.Error("failed to create local object", slog.String("id", remoteObjectMap.ID()), slog.String("resourceName", req.ResourceName), slog.String("location", remoteObjectMap.Href()))
	}

	// Set the headers of the reply to the caller, as per TMF specs
	headers := map[string]string{
		"Location": remoteObjectMap.Href(),
	}
	slog.Info("Object created successfully", slog.String("id", remoteObjectMap.ID()), slog.String("resourceName", req.ResourceName), slog.String("location", remoteObjectMap.Href()))

	return &Response{StatusCode: http.StatusCreated, Headers: headers, Body: remoteObjectMap}

}

// getObject retrieves a TMF object by its ID and type, returning the latest version.
// If the object is not found anywhere, it returns a nil object and no error.
func (svc *Service) getObject(id, objectType string) (*repo.TMFRecord, error) {
	slog.Debug("dbLayer: getObject", slog.String("id", id), slog.String("type", objectType))

	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.GetObject(id, objectType)
	}

	var obj repo.TMFRecord
	err := svc.db.Get(&obj, "SELECT * FROM tmf_object WHERE id = ? AND type = ? ORDER BY version DESC LIMIT 1", id, objectType)
	if err == sql.ErrNoRows {
		slog.Info("Service: Object not found", slog.String("id", id), slog.String("type", objectType))
		return nil, nil // Object not found
	} else if err != nil {
		err = errl.Errorf("failed to get object id=%s type=%s: %w", id, objectType, err)
	}
	return &obj, err
}

// getLocalOrRemoteObject retrieves the object from the database.
// If the proxy is enabled and the object is not found locally or is stale, we try to get it from the remote server.
// If the object is not found anywhere, it returns a nil object and no error.
// There is no way to force the retrieval from the remote server if the objec exists locally and is fresh enough.
func (svc *Service) getLocalOrRemoteObject(req *Request) (*repo.TMFRecord, error) {

	// For organization resources, the req.ID is the organization identifier.
	if req.ResourceName == "organization" {

		if strings.HasPrefix(req.ID, "urn:ngsi-ld:organization:") && !strings.HasPrefix(req.ID, "urn:ngsi-ld:organization:did:elsi:") {
			req.ID = "urn:ngsi-ld:organization:did:elsi:" + strings.TrimPrefix(req.ID, "urn:ngsi-ld:organization:")
		}

	}

	// Check if we have the object locally
	obj, err := svc.getObject(req.ID, req.ResourceName)
	if err != nil {
		return nil, errl.Errorf("failed to get object %s from local service: %w", req.ID, err)
	}

	// If proxy is not enabled, we return whatever we found, which is nil if object is not found.
	// Not finding an object is not an error at this level but the caller is responsible for returning a 404 error to the user.
	if !svc.proxyEnabled {
		return obj, nil
	}

	// If we found the object and the database record is not stale, return it
	if obj != nil {
		if time.Since(obj.GetUpdatedAt()) < svc.fressness {
			slog.Debug("Object found in local database and fresh")
			return obj, nil
		}
		slog.Debug("Object found in local database but stale, retrieving from remote")
	} else {
		slog.Debug("Object not found in local database, retrieving from remote")
	}

	// The object was not found or is stale, so we have to retrieve remotely and update the local database
	remoteObj, err := svc.getRemoteObject(req)
	if err != nil {
		return nil, errl.Errorf("failed to get object %s from remote service: %w", req.ID, err)
	}

	// Store the object locally and return it to caller
	if err := svc.createObject(remoteObj); err != nil {
		slog.Error("failed to cache object", slog.Any("error", err))
		// Return the stale object or nil
		return obj, nil
	}

	slog.Info("Object retrieved from remote and cached successfully", slog.String("id", req.ID), slog.String("resourceName", req.ResourceName))
	return remoteObj, nil

}

func (svc *Service) getRemoteObject(req *Request) (*repo.TMFRecord, error) {
	slog.Debug("retrieving object from remote", slog.String("id", req.ID))

	// Send the access token
	headers := map[string]string{
		"Authorization": "Bearer " + req.AccessToken,
	}

	// Build the path for the request according to TMForum specs
	path := fmt.Sprintf("/tmf-api/%s/%s/%s/%s", req.APIfamily, req.APIVersion, req.ResourceName, req.ID)

	// Send the request to the remote with our HTTP Client
	resp, err := svc.tmfClient.Get(path, headers)
	if err != nil {
		return nil, errl.Errorf("failed to proxy request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errl.Errorf("failed to read response body: %w", err)
	}

	// Not found is not an error at this level, but the caller must check for a nil object
	if resp.StatusCode == 404 {
		return nil, nil
	}

	// TODO: handle caching replies, with status codes in the 3xx range
	if resp.StatusCode >= 300 {
		return nil, errl.Errorf("remote server returned error: %s", string(body))
	}

	// Build the object from the reply
	receivedObjectMap, err := repo.NewTMFObjectMapFromBytes(req.ResourceName, body)
	if err != nil {
		return nil, errl.Errorf("failed to bind request body: %w", err)
	}

	// Prepare the object for the database
	id := receivedObjectMap.ID()
	version := receivedObjectMap.Version()
	lastUpdate := receivedObjectMap.LastUpdate()

	// It is an error if the remote server does not return a 'lastUpdate', but we fix it and log a warning
	if lastUpdate == "" {
		slog.Warn("remote server did not return lastUpdate, fixing it", slog.String("id", id))

		now := time.Now()
		lastUpdate = now.Format(time.RFC3339Nano)
		receivedObjectMap["lastUpdate"] = lastUpdate
	}

	// Instead of storing what we receive, we store a compact serialization of the JSON object with possible updated fields
	receivedContent, err := json.Marshal(receivedObjectMap)
	if err != nil {
		return nil, errl.Errorf("failed to marshal object content: %w", err)
	}

	obj := repo.NewTMFRecord(id, req.ResourceName, version, req.APIVersion, lastUpdate, receivedContent)

	return obj, nil
}

// updateObject updates an existing TMF object.
func (svc *Service) updateObject(obj *repo.TMFRecord) error {
	slog.Debug("dbLayer: updateObject", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.UpdateObject(obj)
	}

	// Make sure timestamps are correct
	obj.UpdatedAt = time.Now().Unix()

	// Execute the SQL, updating only the allowed fields. E.g. seller and buyer can not be updated.
	_, err := svc.db.NamedExec(`UPDATE tmf_object
		SET version = :version, last_update = :last_update, content = :content, updated_at = :updated_at
		WHERE id = :id AND type = :type`,
		obj,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return &ErrObjectExists{ID: obj.ID, Type: obj.Type}
			}
		}
		err = errl.Errorf("failed to update object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}
	return err
}

// deleteObject deletes a TMF object by its ID and type.
func (svc *Service) deleteObject(id, objectType string) error {
	slog.Debug("dbLayer: deleteObject", slog.String("id", id), slog.String("type", objectType))

	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.DeleteObject(id, objectType)
	}

	// Execute the SQL
	_, err := svc.db.Exec("DELETE FROM tmf_object WHERE id = ? AND type = ?", id, objectType)
	if err != nil {
		err = errl.Errorf("failed to delete object id=%s type=%s: %w", id, objectType, err)
	}
	return err
}

// objectFilter is a function that filters TMF objects. If it returns false, the object is excluded from the result.
type objectFilter func(obj *repo.TMFRecord) bool

// listObjects retrieves TMF objects of a given type, returning only the latest version for each unique ID.
// It supports pagination, filtering, and sorting according to TMF630 guidelines.
func (svc *Service) listObjects(req *Request, filter objectFilter) ([]repo.TMFRecord, int, error) {
	if !req.HealthRequest {
		slog.Debug("dbLayer: listObjects", "type", req.ResourceName, "queryParams", req.QueryParams)
	}
	// Direct to another storage system if defined
	if svc.storage != nil {
		return svc.storage.ListObjects(req.ResourceName, req.APIVersion, req.QueryParams)
	}

	// Parse the parameters according to TM Forum specs and build the SELECT
	baseQuery, args, limit, offset, err := BuildSelectFromParms(req.ResourceName, req.QueryParams)
	if err != nil {
		return nil, 0, errl.Errorf("failed to build select query: %w", err)
	}
	if !req.HealthRequest {
		fmt.Printf("SQL: %s\nARGS: %v\n", baseQuery, args)
	}

	var objs []repo.TMFRecord
	var totalCount int
	var offsetCounter int

	// Run the SQL
	rows, err := svc.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, errl.Errorf("performing query %s with args %v: %w", baseQuery, args, err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var obj repo.TMFRecord
		if err := rows.Scan(&obj.ID, &obj.Type, &obj.Version, &obj.APIVersion,
			&obj.Seller, &obj.Buyer, &obj.LastUpdate, &obj.Content, &obj.CreatedAt, &obj.UpdatedAt); err != nil {

			return nil, 0, errl.Errorf("iterating over rows in query %s with args %v: %w", baseQuery, args, err)

		}

		// Callback to the caller for filtering/ammendment of the object
		if filter != nil && !filter(&obj) {
			// The callback said that we should not include this object
			continue
		}

		// If the object has passed all checks, we still have to discard the first 'offset' objects, as specified by the user
		if offsetCounter < offset {
			offsetCounter++
			continue
		}

		// Now we add the object to the result array and check if we reached the limit as specified by the user
		objs = append(objs, obj)

		if len(objs) >= limit {
			break
		}

	}
	if err = rows.Err(); err != nil {
		return nil, 0, errl.Errorf("performing query %s with args %v: %w", baseQuery, args, err)
	}

	return objs, totalCount, err
}

// BuildSelectFromParms creates a SELECT statement based on the query values.
// Some keys are columns in the database row, but most of them are in the JSON object in the 'content' column
// For TMF objects with same id, selects the one with the latest version.
func BuildSelectFromParms(tmfResource string, queryValues url.Values) (query string, arguments []any, qlimit int, qoffset int, theerr error) {

	// Default values if the user did not specify them. -1 is equivalent to no values provided.
	var limit = -1
	var offset = -1

	var buf StringRenderer
	var args []any

	// The main select
	buf.Render(
		`SELECT id, type, max(version) AS version, api_version, seller, buyer, last_update, content, created_at, updated_at FROM tmf_object`,
	)

	// WHERE: normally we expect the resource name of object to be specified, but we support a query for all object types
	if len(tmfResource) > 0 {
		buf.Render(" WHERE type = ?")
		args = append(args, tmfResource)
	}

	// Build the WHERE by processing the query values specified by the user
	for key, values := range queryValues {

		// // We here detect and parse expressions like 'arrayName[*].keyName=values'
		// // It is true if the object contains an array 'arrayName' where one element has name 'keyname' with a value in the set of 'values'
		// selector := "[*]."
		// index := strings.Index(key, selector)
		// if index == 0 {
		// 	// index must be -1 (not found) or >0 (an array name of minimum 1 character before the left '[')
		// 	// index == 0 is a syntax error of the array expression
		// 	err := errl.Errorf("array name in array selector is empty")
		// 	return "", nil, 0, 0, err
		// }
		// if index > 0 {
		// 	// index is the length of the array name (the number of chars before the left bracket)

		// 	arrayName := key[:index]

		// 	// keyname is the string after the dot until the end
		// 	keyName := key[index+len(selector):]
		// 	slog.Debug("array selector", "arrayName", arrayName, "keyName", keyName)

		// 	if len(keyName) == 0 {
		// 		err := errl.Errorf("key name in array selector is empty")
		// 		return "", nil, 0, 0, err
		// 	}

		// 	vals := processValues(values)

		// 	// For a single value, use the SQL '=' to express equality.
		// 	// For and array of values, use the SQL 'IN' to express inclusion in a set.
		// 	if len(vals) == 1 {
		// 		buf.Render(
		// 			" AND EXISTS (SELECT 1 FROM json_each(tmf_object.content, '$.", arrayName, "') WHERE json_extract(value, '$.", keyName, "') = ?)",
		// 		)
		// 	} else if len(vals) > 1 {
		// 		buf.Render(
		// 			" AND EXISTS (SELECT 1 FROM json_each(tmf_object.content, '$.", arrayName, "') WHERE json_extract(value, '$", keyName, "') IN ").RenderSQLList(vals...).Render(")")
		// 	}
		// 	args = append(args, vals...)

		// 	continue

		// }

		// Create additional parts of the SELECT, with some special processing
		switch key {
		case "sort", "fields":
			// TODO: implement processing for these parameters. For the moment, they must be implemented by the caller.
			continue

		case "limit":
			// Just extract the value for later, it will not be used in the SELECT
			limitStr := queryValues.Get("limit")
			if limitStr != "" {
				if l, err := strconv.Atoi(limitStr); err == nil {
					limit = l
				}
			}

		case "offset":
			// Just extract the value for later, it will not be used in the SELECT
			offsetStr := queryValues.Get("offset")
			if offsetStr != "" {
				if l, err := strconv.Atoi(offsetStr); err == nil {
					offset = l
				}
			}

		case "seller", "buyer":
			// A shortcut for DOME and ISBE, to simplify life to applications (but can be also done in a TMF-compliant way).
			// Special processing to allow specifying multiple values in the form 'seller=id1,id2,id3'.
			// We also support the standard HTTP query strings like 'seller=id1,id2&seller=id3'
			vals := processValues(values)

			// Use either an equality (when one element) or an inclusion expression (when several)
			if len(vals) == 1 {
				buf.Render(" AND ", key, " = ?")
			} else if len(vals) > 1 {
				buf.Render(" AND ", key, " IN ").RenderSQLList(vals...)
			}
			args = append(args, vals...)

		case "category.id", "productSpecification.id":
			// Simplification of the query in the category array for common fields

			object := strings.TrimSuffix(key, ".id")

			// Special processing because TMForum allows to specify multiple values
			// in the form 'lifecycleStatus=Launched,Active'
			vals := processValues(values)

			if len(vals) == 1 {
				buf.Render(
					" AND EXISTS (SELECT 1 FROM json_each(tmf_object.content, '$.", object, "') WHERE json_extract(value, '$.id') = ?)",
				)
			} else if len(vals) > 1 {
				buf.Render(
					" AND EXISTS (SELECT 1 FROM json_each(tmf_object.content, '$.", object, "') WHERE json_extract(value, '$.id') IN ").RenderSQLList(vals...).Render(")")
			}
			args = append(args, vals...)

		case "organizationIdentification[*].identificationId", "organizationIdentification.identificationId",
			"individualIdentification[*].identificationId", "individualIdentification.identificationId", "individualIdentification.id":
			// Simplification of the query in the identification arrays for the special case of organization identification data
			arrayName := "organizationIdentification"
			keyName := "identificationId"

			if strings.HasPrefix("individualIdentification", key) {
				arrayName = "individualIdentification"
				keyName = "identificationId"
			}

			// Special processing because TMForum allows to specify multiple values
			// in the form 'lifecycleStatus=Launched,Active'
			vals := processValues(values)

			if len(vals) == 1 {
				buf.Render(
					" AND EXISTS (SELECT 1 FROM json_each(tmf_object.content, '$.", arrayName, "') WHERE json_extract(value, '$.", keyName, "') = ?)",
				)
			} else if len(vals) > 1 {
				buf.Render(
					" AND EXISTS (SELECT 1 FROM json_each(tmf_object.content, '$.", arrayName, "') WHERE json_extract(value, '$.", keyName, "') IN ").RenderSQLList(vals...).Render(")")
			}
			args = append(args, vals...)

		default:

			// Special processing because TMForum allows to specify multiple values
			// in the form 'lifecycleStatus=Launched,Active'
			vals := processValues(values)

			// We perform special processin when the key is simple (no dots), to use a simple and more efficient SQL expression.
			pathParts := strings.Split(key, ".")
			if len(pathParts) == 1 {

				if len(vals) == 1 {
					buf.Render(" AND content->>'$.", key, "' = ?")
				} else {
					buf.Render(" AND content->>'$.", key, "' IN ").RenderSQLList(vals...)
				}
				args = append(args, vals...)
			} else {
				subSql, _, err := GenerateRecursiveJSONQuery("tmf_object", key, vals)
				if err != nil {
					return "", nil, 0, 0, err
				}
				buf.Render(" AND ", subSql)
				args = append(args, vals...)
			}

		}
	}

	// We need to GROUP by id, so we can SELECT the record with the latest version from each group
	buf.Render(" GROUP BY id")

	// // For fairness of presenting results to customers, we want a random ordering, which is consistent and fair with the providers.
	// // Ordering by the hash of the content of the TMF object complies with the requirements, as it is consistent across paginations
	// // and nobody can predict the final ordering a-priory.
	// // For a stable catalog, the ordering is the same for all users and at any time.
	// // When a provider creates or modifies a product, it will be inserted at an unpredictable position in the catalog.
	// //
	// // TODO: we can consider a more advanced variation, where we add to the hash a random number which is
	// // generated each day or week, and keeps the same until a new one is generated.
	// // In this way, ordering is efficient, random, and changes every week (or whatever period is chosen)
	// bu.OrderBy("hash")

	// Build the query, with the statement and the arguments to be used
	sql := buf.String()

	return sql, args, limit, offset, nil
}

// GenerateRecursiveJSONQuery generates a SQL query to search for a value in a JSON object, given a JSON path where some elements may be arrays.
// It uses the SQLite json_tree function to recursively search for the value in the JSON object.
// It returns the SQL query and the value to be used in the query.
func GenerateRecursiveJSONQuery(tableName string, pathInput string, values any) (string, string, error) {

	v := reflect.ValueOf(values)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return "", "", fmt.Errorf("invalid values: must be a slice or array")
	}

	if v.Len() == 0 {
		return "", "", fmt.Errorf("invalid values: no values provided for recursive JSON queries")
	}

	targetValue, ok := v.Index(0).Interface().(string)
	if !ok {
		return "", "", fmt.Errorf("invalid value type for recursive query: %T", v.Index(0).Interface())
	}

	// 2. Transform the path into a GLOB pattern to be used in the SQL query
	// Example: field1[1].field2.field3 -> $.field1[1].field2.field3
	// Example: field1.field2.field3    -> $.field1*.field2*.field3*
	// Example: field1[*].field2.field3    -> $.field1*.field2*.field3*

	pathParts := strings.Split(pathInput, ".")
	var globPath strings.Builder
	globPath.WriteString("$")

	// Regex to check if a field already has an index like field1[1]
	// If the index is '[*]', it means any array index and this will be handled by another part of the code
	hasIndex := regexp.MustCompile(`^(.+)\[(\d+)\]$`)

	for _, p := range pathParts {
		matches := hasIndex.FindStringSubmatch(p)
		if len(matches) > 0 {
			field := matches[1]
			index := matches[2]
			// In GLOB, [[] matches a literal '['
			// We format it as .field[[]index] which means .field[index]
			fmt.Fprintf(&globPath, ".%s[[]%s]", field, index)
		} else {
			// No index: match the key directly OR any array index for that key
			// As a special case, also match the string "[*]", meaning 'any array index'
			// Pattern: .field* matches .field (object) or .field[0] or .field[*]  (array)

			p = strings.TrimSuffix(p, "[*]")
			fmt.Fprintf(&globPath, ".%s*", p)
		}
	}

	// 3. Construct the SQL using json_tree for recursive scanning
	if v.Len() == 1 {
		sql := `EXISTS (SELECT 1 FROM json_tree(` + tableName + `.content)
				WHERE json_tree.fullkey GLOB '` + globPath.String() + `' AND json_tree.value = ?)`
		return strings.TrimSpace(sql), targetValue, nil
	} else {
		var buf StringRenderer
		buf.Render("EXISTS (SELECT 1 FROM json_tree(" + tableName + ".content) WHERE json_tree.fullkey GLOB '" + globPath.String() + "' AND json_tree.value IN ")
		buf.RenderSQLList(values)
		buf.Render(")")
		return strings.TrimSpace(buf.String()), "", nil
	}

}

// StringRenderer is a utility for efficiently building strings by rendering values to a buffer.
// It embeds strings.Builder and adds some convenience methods.
type StringRenderer struct {
	strings.Builder
}

// Render renders the given inputs to the buffer.
// It supports strings, byte slices, integers, bytes and runes.
// It returns the StringRenderer for chaining.
func (r *StringRenderer) Render(inputs ...any) *StringRenderer {
	for _, s := range inputs {
		switch v := s.(type) {
		case string:
			r.WriteString(v)
		case []byte:
			r.Write(v)
		case int:
			r.WriteString(strconv.FormatInt(int64(v), 10))
		case byte:
			r.WriteByte(v)
		case rune:
			r.WriteRune(v)
		default:
			slog.Error("attemping to write something not a string, int, rune, []byte or byte: %T", s)
		}
	}
	return r
}

// Renderln renders the given inputs to the buffer, followed by a newline.
// It supports strings, byte slices, integers, bytes and runes.
// It returns the StringRenderer for chaining.
func (r *StringRenderer) Renderln(inputs ...any) *StringRenderer {
	r.Render(inputs...)
	r.Render('\n')
	return r
}

// RenderSQLList renders an SQL argument list like '(?, ?, ?)' from a slice of arguments.
// The number of '?' is the same as the number of arguments
func (r *StringRenderer) RenderSQLList(inputs ...any) *StringRenderer {
	r.Render("(")

	if len(inputs) == 1 {
		v := reflect.ValueOf(inputs[0])
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			for i := 0; i < v.Len(); i++ {
				if i > 0 {
					r.Render(",")
				}
				r.Render("?")
			}
			r.Render(")")
			return r
		}
	}

	for i := range inputs {
		if i > 0 {
			r.Render(",")
		}
		r.Render("?")
	}
	r.Render(")")
	return r
}

// processValues converts from a slice of strings to a slice of any, handling the case where the values are comma separated
// in the form 'value1,value2,value3' as is possible with the url values in an HTTP request
// For example, the url value 'lifecycleStatus=Launched,Active' will be converted to a slice of two strings "Launched" and "Active"
func processValues(values []string) []any {
	var vals []any
	for _, v := range values {
		parts := strings.Split(v, ",")
		// Allow for whitespace surrounding the elements
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			vals = append(vals, parts[i])
		}
	}

	return vals
}
