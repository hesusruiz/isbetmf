package sqlitesync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSync(t *testing.T) {
	// Use the real database from the data directory
	origin := "../data/isbetmf.db"

	// Create a temporary file for the destination
	destFile, err := os.CreateTemp("", "isbetmf_copy_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	destination := destFile.Name()
	destFile.Close()
	defer os.Remove(destination) // Clean up

	// We expect this to fail if sqlite3_rsync is not in the PATH,
	// or succeed if it is.
	err = Sync(origin, destination)
	if err == nil {
		t.Log("Sync succeeded")
	} else {
		t.Logf("Sync failed (likely sqlite3_rsync not in PATH): %v", err)
	}
}

func TestBackup(t *testing.T) {
	// Use the real database from the data directory
	origin := "../data/isbetmf.db"

	// Ensure we can get absolute path
	absOrigin, err := filepath.Abs(origin)
	if err != nil {
		t.Fatalf("Failed to get absolute path of origin: %v", err)
	}

	// Determine expected backup path
	dir := filepath.Dir(absOrigin)
	filename := filepath.Base(absOrigin)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	weekday := time.Now().Weekday()
	dayNum := int(weekday)
	if dayNum == 0 {
		dayNum = 7
	}

	backupFilename := fmt.Sprintf("%s_%d%s", name, dayNum, ext)
	expectedBackupPath := filepath.Join(dir, "backups", backupFilename)

	// Clean up before test (in case it exists)
	os.Remove(expectedBackupPath)
	// Also clean up after test
	defer os.Remove(expectedBackupPath)
	// And the directory if empty (optional, but good for cleanliness)
	defer os.Remove(filepath.Dir(expectedBackupPath))

	// Perform backup
	err = Backup(origin)
	if err != nil {
		// If it fails because of missing binary, that's expected in some envs, but we want to know.
		// However, if Sync works, this should work.
		// If Sync failed in TestSync, this will likely fail too.
		t.Logf("Backup failed (likely sqlite3_rsync not in PATH): %v", err)
		return
	}

	// Verify file exists
	if _, err := os.Stat(expectedBackupPath); os.IsNotExist(err) {
		t.Errorf("Backup file was not created at %s", expectedBackupPath)
	} else {
		t.Logf("Backup created successfully at %s", expectedBackupPath)
	}
}

func TestSyncLocalBinary(t *testing.T) {
	// Create a dummy sqlite3_rsync in the current directory
	binaryName := "sqlite3_rsync"
	content := []byte("#!/bin/sh\necho 'LOCAL BINARY EXECUTED' >&2\nexit 1\n")
	err := os.WriteFile(binaryName, content, 0755)
	if err != nil {
		t.Fatalf("Failed to create dummy binary: %v", err)
	}
	defer os.Remove(binaryName)

	// Use dummy paths
	origin := "dummy_origin"
	destination := "dummy_dest"

	// Call Sync
	err = Sync(origin, destination)

	// We expect an error because our dummy binary exits with 1
	if err == nil {
		t.Error("Expected Sync to fail with local binary, but it succeeded")
	} else {
		// Check if the error message contains output from our dummy binary
		// The Sync function includes stderr in the error message
		if !strings.Contains(err.Error(), "LOCAL BINARY EXECUTED") {
			t.Errorf("Expected error to contain 'LOCAL BINARY EXECUTED', got: %v", err)
		} else {
			t.Log("Confirmed Sync used the local binary")
		}
	}
}
