package replicate

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
	repo "github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection and operations
type Database struct {
	db *sqlx.DB
}

// NewDatabase creates a new database connection
func NewDatabase(databaseFile string) (*Database, error) {
	// Open SQLite database
	sqldb, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, errl.Errorf("failed to open database: %w", err)
	}

	// Create sqlx wrapper
	db := sqlx.NewDb(sqldb, "sqlite3")

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, errl.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db: db,
	}

	// Initialize database schema
	if err := database.initializeSchema(); err != nil {
		return nil, errl.Errorf("failed to initialize database schema: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// initializeSchema creates the database tables if they don't exist
func (d *Database) initializeSchema() error {
	// Use the same table schema as the existing repository
	_, err := d.db.Exec(repo.CreateTMFTableSQL)
	if err != nil {
		return errl.Errorf("failed to create table: %w", err)
	}

	slog.Info("Database schema initialized successfully")
	return nil
}

// StoreObject stores a TMF object in the database
func (d *Database) StoreObject(obj *repo.TMFRecord) error {
	slog.Debug("Storing object", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Check if object already exists
	existing, err := d.GetObject(obj.ID, obj.Type)
	if err != nil {
		return errl.Errorf("failed to check existing object: %w", err)
	}

	if existing != nil {
		// Object exists, update it
		return d.UpdateObject(obj)
	}

	// Object doesn't exist, create it
	return d.CreateObject(obj)
}

// CreateObject creates a new TMF object in the database
func (d *Database) CreateObject(obj *repo.TMFRecord) error {
	slog.Debug("Creating object", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	query := `INSERT INTO tmf_object (id, type, version, api_version, last_update, content, created_at, updated_at)
		VALUES (:id, :type, :version, :api_version, :last_update, :content, :created_at, :updated_at)`

	_, err := d.db.NamedExec(query, obj)
	if err != nil {
		return errl.Errorf("failed to create object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}

	slog.Debug("Object created successfully", slog.String("id", obj.ID), slog.String("type", obj.Type))
	return nil
}

// UpdateObject updates an existing TMF object in the database
func (d *Database) UpdateObject(obj *repo.TMFRecord) error {
	slog.Debug("Updating object", slog.String("id", obj.ID), slog.String("type", obj.Type), slog.String("version", obj.Version))

	// Update the updated_at timestamp
	obj.UpdatedAt = time.Now().Unix()

	query := `UPDATE tmf_object SET version = :version, last_update = :last_update, content = :content, updated_at = :updated_at 
		WHERE id = :id AND type = :type`

	result, err := d.db.NamedExec(query, obj)
	if err != nil {
		return errl.Errorf("failed to update object id=%s type=%s: %w", obj.ID, obj.Type, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errl.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("object not found: id=%s type=%s", obj.ID, obj.Type)
	}

	slog.Debug("Object updated successfully", slog.String("id", obj.ID), slog.String("type", obj.Type))
	return nil
}

// GetObject retrieves a TMF object by its ID and type
func (d *Database) GetObject(id, objectType string) (*repo.TMFRecord, error) {
	slog.Debug("Getting object", slog.String("id", id), slog.String("type", objectType))

	var obj repo.TMFRecord
	query := "SELECT * FROM tmf_object WHERE id = ? AND type = ? ORDER BY version DESC LIMIT 1"

	err := d.db.Get(&obj, query, id, objectType)
	if err == sql.ErrNoRows {
		return nil, nil // Object not found
	} else if err != nil {
		return nil, errl.Errorf("failed to get object id=%s type=%s: %w", id, objectType, err)
	}

	return &obj, nil
}

// ListObjects retrieves all TMF objects of a given type
func (d *Database) ListObjects(objectType string) ([]repo.TMFRecord, error) {
	slog.Debug("Listing objects", slog.String("type", objectType))

	var objs []repo.TMFRecord
	query := `SELECT t1.*
		FROM tmf_object t1
		INNER JOIN (
			SELECT id, type, MAX(version) AS max_version
			FROM tmf_object
			WHERE type = ?
			GROUP BY id, type
		) AS t2
		ON t1.id = t2.id AND t1.type = t2.type AND t1.version = t2.max_version
		WHERE t1.type = ?
		ORDER BY t1.created_at DESC`

	err := d.db.Select(&objs, query, objectType, objectType)
	if err != nil {
		return nil, errl.Errorf("failed to list objects for type %s: %w", objectType, err)
	}

	return objs, nil
}

// GetObjectCount returns the number of objects of a given type
func (d *Database) GetObjectCount(objectType string) (int, error) {
	slog.Debug("Getting object count", slog.String("type", objectType))

	var count int
	query := `SELECT COUNT(DISTINCT id) FROM tmf_object WHERE type = ?`

	err := d.db.Get(&count, query, objectType)
	if err != nil {
		return 0, errl.Errorf("failed to get object count for type %s: %w", objectType, err)
	}

	return count, nil
}

// GetTotalObjectCount returns the total number of objects in the database
func (d *Database) GetTotalObjectCount() (int, error) {
	slog.Debug("Getting total object count")

	var count int
	query := "SELECT COUNT(*) FROM tmf_object"

	err := d.db.Get(&count, query)
	if err != nil {
		return 0, errl.Errorf("failed to get total object count: %w", err)
	}

	return count, nil
}

// DeleteObject deletes a TMF object by its ID and type
func (d *Database) DeleteObject(id, objectType string) error {
	slog.Debug("Deleting object", slog.String("id", id), slog.String("type", objectType))

	query := "DELETE FROM tmf_object WHERE id = ? AND type = ?"
	result, err := d.db.Exec(query, id, objectType)
	if err != nil {
		return errl.Errorf("failed to delete object id=%s type=%s: %w", id, objectType, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errl.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("object not found: id=%s type=%s", id, objectType)
	}

	slog.Debug("Object deleted successfully", slog.String("id", id), slog.String("type", objectType))
	return nil
}

// ClearDatabase removes all objects from the database
func (d *Database) ClearDatabase() error {
	slog.Info("Clearing database")

	_, err := d.db.Exec("DELETE FROM tmf_object")
	if err != nil {
		return errl.Errorf("failed to clear database: %w", err)
	}

	slog.Info("Database cleared successfully")
	return nil
}
