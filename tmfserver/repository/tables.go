package repository

import (
	"log/slog"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/jmoiron/sqlx"
)

// CreateTMFTableSQL is the table 'tmf_object', the one holding all objects of all types
const CreateTMFTableSQL = `CREATE TABLE IF NOT EXISTS tmf_object (
	"id" TEXT NOT NULL,
	"type" TEXT NOT NULL,
	"version" TEXT DEFAULT '',
	"api_version" TEXT DEFAULT '',
	"seller" TEXT DEFAULT '',
	"seller_operator" TEXT DEFAULT '',
	"buyer" TEXT DEFAULT '',
	"buyer_operator" TEXT DEFAULT '',
	"last_update" TEXT DEFAULT '',
	"content" BLOB NOT NULL,
	"random" INTEGER DEFAULT 0,
	"created_at" INTEGER,
	"updated_at" INTEGER,
	PRIMARY KEY ("id", "type", "version")
);`

// Used for migrations and to keep the database file compact
const DeleteTMFTableSQL = `DROP TABLE IF EXISTS tmf_object;`
const VacuumSQL = `VACUUM;`

func NewDBService(configuration *config.Config) (*sqlx.DB, error) {

	// Build the connection string with the parameters we want to use.

	// _journal_mode=WAL Enables Write-Ahead Logging for high concurrency (many readers, one writer).
	dbName := "file:" + configuration.Dbname + "?_journal_mode=WAL"

	// _cache_size=-100000 Sets the cache size to 100000 kilobytes (100MB) instead of the default 2MB,
	// which is a good balance between performance and memory usage.
	// The negative sign indicates that the cache size is in kilobytes, not pages.
	dbName = dbName + "&_cache_size=-100000"

	// _busy_timeout=5000 Sets the 5 seconds timeout that the connection will wait and retry when the database is locked,
	// to mitigate the SQLITE_BUSY errors.
	dbName = dbName + "&_busy_timeout=5000"

	// _foreign_keys=on Enforces foreign key constraints.
	dbName = dbName + "&_foreign_keys=on"

	// _synchronous=NORMAL Offers a good balance between data safety and performance.
	dbName = dbName + "&_synchronous=NORMAL"

	// _txlock=immediate To prevent potential deadlocks or missed busy timeouts that can occur when a read transaction
	// is implicitly upgraded to a write transaction.
	dbName = dbName + "&_txlock=immediate"

	// _cache=shared Enables shared cache mode, which allows multiple connections to share the same cache.
	dbName = dbName + "&_cache=shared"

	// _defer_foreign_keys=on Delays the enforcement of foreign key constraints until the transaction is committed.
	dbName = dbName + "&_defer_foreign_keys=on"

	// Connect to the database.
	db, err := sqlx.Connect("sqlite3", dbName)
	if err != nil {
		return nil, errl.Errorf("failed to connect to database: %w", err)
	}

	// Create tables if they do not exist
	slog.Info("About to create tables if they do not exist")
	err = CreateTables(db)
	if err != nil {
		return nil, errl.Error(err)
	}

	return db, nil
}

// CreateTables creates the tables in the database if they do not exist.
// It also handles automatic schema/data migration when possible.
func CreateTables(db *sqlx.DB) error {

	if _, err := db.Exec(CreateTMFTableSQL); err != nil {
		return errl.Errorf("failed to create tmf_object table: %w", err)
	}

	if err := RunMigrationsUp(db); err != nil {
		return errl.Error(err)
	}

	return nil
}
