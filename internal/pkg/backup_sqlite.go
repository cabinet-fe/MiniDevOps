package pkg

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/sqlite"
)

// PrepareSlimSQLiteBackup copies the database at srcPath into a new temporary file using SQLite
// VACUUM INTO (consistent snapshot), then removes audit logs, build history, and notifications
// so backup archives stay small and omit high-churn operational tables.
func PrepareSlimSQLiteBackup(srcPath string) (outPath string, cleanup func(), err error) {
	absSrc, err := filepath.Abs(srcPath)
	if err != nil {
		return "", nil, fmt.Errorf("abs src path: %w", err)
	}
	if _, err := os.Stat(absSrc); err != nil {
		return "", nil, err
	}

	snap, err := os.CreateTemp("", "buildflow-backup-snap-*.sqlite")
	if err != nil {
		return "", nil, fmt.Errorf("create snap temp: %w", err)
	}
	snapPath := snap.Name()
	_ = snap.Close()

	cleanup = func() { _ = os.Remove(snapPath) }

	roDSN := fmt.Sprintf("file:%s?mode=ro&_busy_timeout=60000", filepath.ToSlash(absSrc))
	roDB, err := sql.Open("sqlite", roDSN)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("open ro db: %w", err)
	}
	defer func() { _ = roDB.Close() }()

	if err := roDB.Ping(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("ping ro db: %w", err)
	}

	quotedSnap := strings.ReplaceAll(filepath.ToSlash(snapPath), "'", "''")
	if _, err := roDB.Exec("VACUUM INTO '" + quotedSnap + "'"); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("vacuum into: %w", err)
	}

	rwDSN := fmt.Sprintf("file:%s?_busy_timeout=60000", filepath.ToSlash(snapPath))
	rwDB, err := sql.Open("sqlite", rwDSN)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("open snap rw: %w", err)
	}

	if _, err := rwDB.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		_ = rwDB.Close()
		cleanup()
		return "", nil, fmt.Errorf("pragma foreign_keys: %w", err)
	}

	for _, q := range []string{
		"DELETE FROM notifications",
		"DELETE FROM builds",
		"DELETE FROM audit_logs",
	} {
		if _, err := rwDB.Exec(q); err != nil {
			_ = rwDB.Close()
			cleanup()
			return "", nil, fmt.Errorf("%s: %w", q, err)
		}
	}

	if _, err := rwDB.Exec("VACUUM"); err != nil {
		_ = rwDB.Close()
		cleanup()
		return "", nil, fmt.Errorf("vacuum: %w", err)
	}
	if err := rwDB.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("close snap: %w", err)
	}

	return snapPath, cleanup, nil
}
