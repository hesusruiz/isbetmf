package repository

import (
	"log/slog"
	"time"

	"github.com/hesusruiz/isbetmf/sqlitesync"
	"github.com/jmoiron/sqlx"
)

// ScheduleMaintenance schedules periodic database maintenance tasks like VACUUM or backups.
// It runs the maintenance task every 'interval'.
func ScheduleMaintenance(db *sqlx.DB, dbPath string, hours int) {

	if hours < 1 {
		hours = 1
	}

	interval := time.Duration(hours) * time.Hour

	// Perform maintenance when started
	PerformMaintenance(db, dbPath)

	go func() {
		slog.Info("Database maintenance scheduled", "interval", interval)

		// then, perform maintenance at exactly the time "o-clock" every interval hours
		nextRun := time.Now().Truncate(time.Hour).Add(interval)

		for {
			time.Sleep(time.Until(nextRun))
			PerformMaintenance(db, dbPath)
			nextRun = time.Now().Truncate(time.Hour).Add(interval)
		}
	}()
}

// PerformMaintenance performs the actual database maintenance tasks (VACUUM and Backup).
func PerformMaintenance(db *sqlx.DB, dbPath string) {
	slog.Info("Executing scheduled database cleanup...")

	// Perform a checkpoint to ensure the WAL file is truncated
	start := time.Now()
	if err := forceWalTruncate(db); err != nil {
		slog.Error("failed to truncate WAL file", "error", err, "elapsed", time.Since(start))
	} else {
		slog.Info("WAL file truncated successfully", "elapsed", time.Since(start))
	}

	// Perform VACUUM
	start = time.Now()
	if _, err := db.Exec(VacuumSQL); err != nil {
		slog.Error("failed to vacuum database", "error", err, "elapsed", time.Since(start))
	} else {
		slog.Info("Database vacuumed successfully", "elapsed", time.Since(start))
	}

	// Perform Backup
	start = time.Now()
	if err := sqlitesync.Backup(dbPath); err != nil {
		slog.Error("failed to backup database", "error", err, "elapsed", time.Since(start))
	} else {
		slog.Info("Database backup completed successfully", "elapsed", time.Since(start))
	}
}

func forceWalTruncate(db *sqlx.DB) error {
	// TRUNCATE ensures the checkpoint runs to completion, deleting the WAL file.
	// This call can block briefly if a write is in progress.
	_, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}
