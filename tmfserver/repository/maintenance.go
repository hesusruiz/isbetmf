package repository

import (
	"log/slog"
	"time"

	"github.com/hesusruiz/isbetmf/sqlitesync"
	"github.com/jmoiron/sqlx"
)

// ScheduleMaintenance schedules periodic database maintenance tasks like VACUUM or backups.
// Periodicity is one per day, at the provided time.
func ScheduleMaintenance(db *sqlx.DB, dbPath string, hour, minute int) {
	if hour < 0 {
		return
	}

	// Schedule cleanups every night
	targetHour := hour
	targetMinute := minute
	targetSecond := 0

	go func() {
		for {
			now := time.Now()

			// Calculate the next scheduled time
			nextRun := time.Date(
				now.Year(), now.Month(), now.Day(),
				targetHour, targetMinute, targetSecond, 0, now.Location(),
			)

			// If the next run time is in the past, schedule it for the next day
			if nextRun.Before(now) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			slog.Info("Next database cleanup scheduled", "time", nextRun)

			// Wait until the next run time
			time.Sleep(time.Until(nextRun))

			// Execute the function
			PerformMaintenance(db, dbPath)
		}
	}()
}

// PerformMaintenance performs the actual database maintenance tasks (VACUUM and Backup).
func PerformMaintenance(db *sqlx.DB, dbPath string) {
	slog.Info("Executing scheduled database cleanup...")

	// Perform VACUUM
	if _, err := db.Exec(VacuumSQL); err != nil {
		slog.Error("failed to vacuum database", "error", err)
	} else {
		slog.Info("Database vacuumed successfully")
	}

	// Perform Backup
	if err := sqlitesync.Backup(dbPath); err != nil {
		slog.Error("failed to backup database", "error", err)
	} else {
		slog.Info("Database backup completed successfully")
	}
}
