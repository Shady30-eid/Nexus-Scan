package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/models"
)

type DB struct {
	conn *sql.DB
	log  *logger.Logger
}

func New(path string, log *logger.Logger) (*DB, error) {
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000&_synchronous=NORMAL", path)
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return &DB{conn: conn, log: log}, nil
}

func (d *DB) Close() error {
	return d.conn.Close()
}

func (d *DB) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version     INTEGER PRIMARY KEY,
			applied_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS scan_history (
			id               TEXT PRIMARY KEY NOT NULL,
			device_id        TEXT NOT NULL,
			package_name     TEXT NOT NULL,
			file_path        TEXT NOT NULL DEFAULT '',
			sha256_hash      TEXT NOT NULL DEFAULT '',
			threat_name      TEXT NOT NULL DEFAULT '',
			severity         TEXT NOT NULL DEFAULT 'LOW' CHECK(severity IN ('LOW','MEDIUM','HIGH','CRITICAL','')),
			infection_vector TEXT NOT NULL DEFAULT '',
			malware_family   TEXT NOT NULL DEFAULT '',
			description      TEXT NOT NULL DEFAULT '',
			scan_timestamp   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(device_id, package_name, sha256_hash)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_scan_history_device_id ON scan_history(device_id);`,
		`CREATE INDEX IF NOT EXISTS idx_scan_history_severity  ON scan_history(severity);`,
		`CREATE INDEX IF NOT EXISTS idx_scan_history_timestamp ON scan_history(scan_timestamp);`,
		`CREATE INDEX IF NOT EXISTS idx_scan_history_hash      ON scan_history(sha256_hash);`,

		`CREATE TABLE IF NOT EXISTS malware_signatures (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			malware_name   TEXT    NOT NULL,
			sha256_hash    TEXT    NOT NULL UNIQUE,
			malware_family TEXT    NOT NULL DEFAULT '',
			severity       TEXT    NOT NULL DEFAULT 'MEDIUM' CHECK(severity IN ('LOW','MEDIUM','HIGH','CRITICAL')),
			description    TEXT    NOT NULL DEFAULT '',
			created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE INDEX IF NOT EXISTS idx_malware_signatures_hash     ON malware_signatures(sha256_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_malware_signatures_family   ON malware_signatures(malware_family);`,
		`CREATE INDEX IF NOT EXISTS idx_malware_signatures_severity ON malware_signatures(severity);`,

		`CREATE TABLE IF NOT EXISTS remediation_log (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id    TEXT    NOT NULL,
			package_name TEXT    NOT NULL,
			action       TEXT    NOT NULL,
			success      INTEGER NOT NULL DEFAULT 0,
			message      TEXT    NOT NULL DEFAULT '',
			executed_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE INDEX IF NOT EXISTS idx_remediation_log_device_id ON remediation_log(device_id);`,
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback()

	for i, migration := range migrations {
		if _, err := tx.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w\nSQL: %s", i, err, migration)
		}
	}

	if _, err := tx.Exec(`INSERT OR IGNORE INTO schema_migrations(version) VALUES (1)`); err != nil {
		return fmt.Errorf("record migration version: %w", err)
	}

	return tx.Commit()
}

func (d *DB) Seed() error {
	var count int
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM malware_signatures`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	seeds := []struct {
		name, hash, family, severity, desc string
	}{
		{"BankBot.A", "a3f5c8d2e1b4f9a6c7d0e3b2a5f8c1d4e7b0a3f6", "BankBot", "CRITICAL", "Fake banking overlay steals credentials. Intercepts SMS OTP codes."},
		{"SpyAgent.B", "b4e6d9c2f5a8b1e4d7c0f3a6b9e2d5c8f1a4b7e0", "SpyAgent", "HIGH", "SMS/call logger, GPS tracker, uploads device data to C2 server."},
		{"FakeWhatsApp.C", "c5f7e0d3a6b9c2f5e8d1b4a7c0f3e6b9d2a5c8f1", "Trojan", "HIGH", "Cloned WhatsApp UI harvests credentials and contact lists."},
		{"Adware.D", "d6a8f1e4b7c0d3f6a9e2b5d8c1f4a7e0b3d6a9f2", "Adware", "MEDIUM", "Persistent adware displaying full-screen overlays. Survives reboots via service."},
		{"SMSSpam.E", "e7b9a2f5c8d1e4b7a0d3f6c9b2e5a8f1d4c7b0e3", "SMSSpam", "LOW", "Sends premium-rate SMS messages without user consent."},
		{"Rootnik.F", "f8c0b3e6a9d2f5c8b1e4a7d0c3f6b9e2a5d8c1f4", "Rootnik", "CRITICAL", "Exploits local privilege escalation, installs persistent root backdoor."},
		{"GhostPush.G", "a9d1c4f7b0e3a6d9c2f5b8e1d4a7c0f3b6e9d2a5", "GhostPush", "HIGH", "Silently installs arbitrary APKs by abusing accessibility services."},
		{"FakeAV.H", "b0e2d5a8c1f4b7e0d3a6c9f2b5e8d1a4c7f0b3e6", "FakeAV", "MEDIUM", "Scareware claiming to detect viruses, demands payment for 'removal'."},
		{"Lotoor.I", "c1f3e6b9d2a5c8f1e4b7d0a3f6c9e2b5d8a1c4f7", "Lotoor", "CRITICAL", "Exploits CVE-2012-0056 and others to gain root access."},
		{"DroidDream.J", "d2a4f7c0e3b6d9a2f5c8e1b4d7a0f3c6e9b2d5a8", "DroidDream", "CRITICAL", "Early Google Play malware, root exploit + data exfiltration."},
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO malware_signatures(malware_name, sha256_hash, malware_family, severity, description, created_at) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range seeds {
		if _, err := stmt.Exec(s.name, s.hash, s.family, s.severity, s.desc, time.Now()); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *DB) LookupHash(hash string) (*models.MalwareSignature, error) {
	row := d.conn.QueryRow(
		`SELECT id, malware_name, sha256_hash, malware_family, severity, description, created_at
		 FROM malware_signatures WHERE sha256_hash = ? LIMIT 1`, hash)

	var sig models.MalwareSignature
	var createdAt string
	err := row.Scan(&sig.ID, &sig.MalwareName, &sig.SHA256Hash, &sig.MalwareFamily, &sig.Severity, &sig.Description, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lookup hash: %w", err)
	}
	sig.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &sig, nil
}

func (d *DB) SaveScanResult(result *models.ScanResult) error {
	_, err := d.conn.Exec(
		`INSERT OR REPLACE INTO scan_history
		 (id, device_id, package_name, file_path, sha256_hash, threat_name, severity, infection_vector, malware_family, description, scan_timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.ID, result.DeviceID, result.PackageName, result.FilePath,
		result.SHA256Hash, result.ThreatName, string(result.Severity),
		result.InfectionVector, result.MalwareFamily, result.Description,
		result.ScanTimestamp.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("save scan result: %w", err)
	}
	return nil
}

func (d *DB) GetScanHistory(deviceID string) ([]models.ScanResult, error) {
	var query string
	var args []interface{}

	if deviceID != "" {
		query = `SELECT id, device_id, package_name, file_path, sha256_hash, threat_name, severity, infection_vector, malware_family, description, scan_timestamp
		         FROM scan_history WHERE device_id = ? ORDER BY scan_timestamp DESC LIMIT 500`
		args = []interface{}{deviceID}
	} else {
		query = `SELECT id, device_id, package_name, file_path, sha256_hash, threat_name, severity, infection_vector, malware_family, description, scan_timestamp
		         FROM scan_history ORDER BY scan_timestamp DESC LIMIT 500`
	}

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get scan history: %w", err)
	}
	defer rows.Close()

	var results []models.ScanResult
	for rows.Next() {
		var r models.ScanResult
		var ts string
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.PackageName, &r.FilePath, &r.SHA256Hash,
			&r.ThreatName, &r.Severity, &r.InfectionVector, &r.MalwareFamily, &r.Description, &ts); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		r.ScanTimestamp, _ = time.Parse(time.RFC3339, ts)
		results = append(results, r)
	}
	return results, rows.Err()
}

func (d *DB) LogRemediation(deviceID, packageName, action string, success bool, message string) error {
	_, err := d.conn.Exec(
		`INSERT INTO remediation_log(device_id, package_name, action, success, message, executed_at) VALUES(?,?,?,?,?,?)`,
		deviceID, packageName, action, boolToInt(success), message, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
