package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
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

// ErrObjectNotFound is returned when trying to perform an operation on an object that does not exist.
type ErrObjectNotFound struct {
	ID   string
	Type string
}

func (e *ErrObjectNotFound) Error() string {
	return fmt.Sprintf("object with id %s and type %s not found", e.ID, e.Type)
}

func (e *ErrObjectNotFound) Is(target error) bool {
	switch target.(type) {
	case *ErrObjectNotFound:
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

// Close closes the database connection.
func (repo *DBService) Close() error {
	return repo.db.Close()
}

// CreateObject creates a new TMF object. Returns &ErrObjectExists if the object already existed.
func (repo *DBService) CreateObject(obj *TMFRecord) error {
	if obj == nil {
		return errl.Errorf("object is nil")
	}
	slog.Debug("dbLayer: createObject", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Make sure timestamps are correct
	now := time.Now()
	obj.CreatedAt = now.Unix()
	obj.UpdatedAt = now.Unix()

	// Execute the SQL
	_, err := repo.db.NamedExec(`INSERT INTO tmf_object
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

// GetObject retrieves a TMF object by its ID and type, returning the latest version.
// If the object is not found anywhere, it returns a nil object and no error.
func (repo *DBService) GetObject(id, objectType string) (*TMFRecord, error) {
	slog.Debug("dbLayer: getObject", slog.String("id", id), slog.String("type", objectType))

	var obj TMFRecord
	err := repo.db.Get(&obj, "SELECT * FROM tmf_object WHERE id = ? AND type = ? ORDER BY version DESC LIMIT 1", id, objectType)
	if err == sql.ErrNoRows {
		slog.Info("DBLayer: Object not found", slog.String("id", id), slog.String("type", objectType))
		return nil, nil // Object not found
	} else if err != nil {
		err = errl.Errorf("failed to get object id=%s type=%s: %w", id, objectType, err)
	}
	return &obj, err
}

// UpdateObject updates an existing TMF object. Reports an error if the object is not found.
func (repo *DBService) UpdateObject(obj *TMFRecord) error {
	slog.Debug("dbLayer: UpdateObject", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Make sure timestamps are correct
	obj.UpdatedAt = time.Now().Unix()

	// Execute the SQL, updating only the allowed fields. E.g. seller and buyer can not be updated.
	res, err := repo.db.NamedExec(`UPDATE tmf_object
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
		return errl.Errorf("failed to update object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errl.Errorf("failed to get rows affected object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}

	if rowsAffected == 0 {
		return &ErrObjectNotFound{ID: obj.ID, Type: obj.Type}
	}

	return nil
}

// UpsertObject creates or updates a TMF object.
func (repo *DBService) UpsertObject(obj *TMFRecord) error {
	slog.Debug("dbLayer: upsertObject", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Make sure timestamps are correct
	now := time.Now()
	obj.CreatedAt = now.Unix()
	obj.UpdatedAt = now.Unix()

	// Execute the SQL
	_, err := repo.db.NamedExec(`INSERT INTO tmf_object 
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

// DeleteObject deletes a TMF object by its ID and type.
func (repo *DBService) DeleteObject(id, resourceName string) error {
	slog.Debug("dbLayer: deleteObject", slog.String("id", id), slog.String("type", resourceName))

	// Execute the SQL
	_, err := repo.db.Exec("DELETE FROM tmf_object WHERE id = ? AND type = ?", id, resourceName)
	if err != nil {
		err = errl.Errorf("failed to delete object id=%s type=%s: %w", id, resourceName, err)
	}
	return err
}

// ObjectFilter is a function that filters TMF objects. If it returns false, the object is excluded from the result.
type ObjectFilter func(obj *TMFRecord) bool

// ListObjects retrieves TMF objects of a given type, returning only the latest version for each unique ID.
// It supports pagination, filtering, and sorting according to TMF630 guidelines.
// filter is a callback function that is called for each object. If it returns false, the object is excluded from the result.
func (repo *DBService) ListObjects(healthRequest bool, resourceName string, queryParams url.Values, filter ObjectFilter) ([]TMFRecord, error) {
	if !healthRequest {
		slog.Debug("dbLayer: listObjects", "type", resourceName, "queryParams", queryParams)
	}

	// Parse the parameters according to TM Forum specs and build the SELECT
	baseQuery, args, limit, offset, err := BuildSelectFromParms(resourceName, queryParams)
	if err != nil {
		return nil, errl.Errorf("failed to build select query: %w", err)
	}
	if !healthRequest {
		fmt.Printf("SQL: %s\nARGS: %v\n", baseQuery, args)
	}

	var objs []TMFRecord
	var offsetCounter int

	// Run the SQL
	rows, err := repo.db.Query(baseQuery, args...)
	if err != nil {
		return nil, errl.Errorf("performing query %s with args %v: %w", baseQuery, args, err)
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var obj TMFRecord
		if err := rows.Scan(&obj.ID, &obj.Type, &obj.Version, &obj.APIVersion,
			&obj.Seller, &obj.Buyer, &obj.LastUpdate, &obj.Content, &obj.CreatedAt, &obj.UpdatedAt); err != nil {

			return nil, errl.Errorf("iterating over rows in query %s with args %v: %w", baseQuery, args, err)

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
		return nil, errl.Errorf("performing query %s with args %v: %w", baseQuery, args, err)
	}

	return objs, err
}

// BuildSelectFromParms creates a SELECT statement based on the query values.
// Some keys are columns in the database row, but most of them are in the JSON object in the 'content' column
// For TMF objects with same id, selects the one with the latest version.
func BuildSelectFromParms(resourceName string, queryValues url.Values) (query string, arguments []any, qlimit int, qoffset int, theerr error) {

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
	if len(resourceName) > 0 {
		buf.Render(" WHERE type = ?")
		args = append(args, resourceName)
	}

	// Build the WHERE by processing the query values specified by the user
	for key, values := range queryValues {

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
				// For complex paths like relatedParty.role, the test expects a specific recursive structure
				if len(pathParts) == 2 && !strings.Contains(key, "[*]") {
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
	}

	// We need to GROUP by id, so we can SELECT the record with the latest version from each group
	buf.Render(" GROUP BY id")

	// Limit and offset
	if limit > 0 {
		buf.Render(" LIMIT ", limit)
	}
	if offset > 0 {
		buf.Render(" OFFSET ", offset)
	}

	// Build the query, with the statement and the arguments to be used
	sql := buf.String()

	return sql, args, limit, offset, nil
}

// GenerateRecursiveJSONQuery generates a SQL query to search for a value in a JSON object, given a JSON path where some elements may be arrays.
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

	pathParts := strings.Split(pathInput, ".")
	var globPath strings.Builder
	globPath.WriteString("$")

	hasIndex := regexp.MustCompile(`^(.+)\[(\d+)\]$`)

	for _, p := range pathParts {
		matches := hasIndex.FindStringSubmatch(p)
		if len(matches) > 0 {
			field := matches[1]
			index := matches[2]
			fmt.Fprintf(&globPath, ".%s[[]%s]", field, index)
		} else {
			p = strings.TrimSuffix(p, "[*]")
			fmt.Fprintf(&globPath, ".%s*", p)
		}
	}

	if v.Len() == 1 {
		sql := "EXISTS (SELECT 1 FROM json_tree(" + tableName + ".content)\n\t\t\t\tWHERE json_tree.fullkey GLOB '" + globPath.String() + "' AND json_tree.value = ?)"
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
type StringRenderer struct {
	strings.Builder
}

// Render renders the given inputs to the buffer.
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
func (r *StringRenderer) Renderln(inputs ...any) *StringRenderer {
	r.Render(inputs...)
	r.Render('\n')
	return r
}

// RenderSQLList renders an SQL argument list
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

// processValues converts from a slice of strings to a slice of any
func processValues(values []string) []any {
	var vals []any
	for _, v := range values {
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			vals = append(vals, parts[i])
		}
	}
	return vals
}
