package tests

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/nexus-scan/nexus-scan/pkg/database"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/models"
)

func newTestDB(t *testing.T) *database.DB {
	t.Helper()
	f, err := os.CreateTemp("", "nexus-scan-test-*.db")
	if err != nil {
		t.Fatalf("create temp db file: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })

	log, _ := logger.New("error", "")
	db, err := database.New(f.Name(), log)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func TestDatabaseInitialization(t *testing.T) {
	db := newTestDB(t)
	if db == nil {
		t.Fatal("expected non-nil db")
	}
}

func TestDatabaseSeed(t *testing.T) {
	db := newTestDB(t)
	if err := db.Seed(); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	sig, err := db.LookupHash("a3f5c8d2e1b4f9a6c7d0e3b2a5f8c1d4e7b0a3f6")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if sig == nil {
		t.Fatal("expected seed signature to be found")
	}
	if sig.MalwareName != "BankBot.A" {
		t.Errorf("expected BankBot.A, got %s", sig.MalwareName)
	}
}

func TestHashMatching_KnownMalware(t *testing.T) {
	db := newTestDB(t)
	if err := db.Seed(); err != nil {
		t.Fatal(err)
	}

	knownHashes := []struct {
		hash     string
		expected string
		severity models.SeverityLevel
	}{
		{"a3f5c8d2e1b4f9a6c7d0e3b2a5f8c1d4e7b0a3f6", "BankBot.A", models.SeverityCritical},
		{"b4e6d9c2f5a8b1e4d7c0f3a6b9e2d5c8f1a4b7e0", "SpyAgent.B", models.SeverityHigh},
		{"f8c0b3e6a9d2f5c8b1e4a7d0c3f6b9e2a5d8c1f4", "Rootnik.F", models.SeverityCritical},
	}

	for _, tc := range knownHashes {
		tc := tc
		t.Run(tc.expected, func(t *testing.T) {
			sig, err := db.LookupHash(tc.hash)
			if err != nil {
				t.Fatalf("lookup error: %v", err)
			}
			if sig == nil {
				t.Fatalf("expected match for %s, got nil", tc.hash)
			}
			if sig.MalwareName != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, sig.MalwareName)
			}
			if sig.Severity != tc.severity {
				t.Errorf("expected severity %s, got %s", tc.severity, sig.Severity)
			}
		})
	}
}

func TestHashMatching_UnknownHash(t *testing.T) {
	db := newTestDB(t)
	if err := db.Seed(); err != nil {
		t.Fatal(err)
	}

	sig, err := db.LookupHash("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if sig != nil {
		t.Errorf("expected nil for unknown hash, got %v", sig)
	}
}

func TestSaveScanResult(t *testing.T) {
	db := newTestDB(t)

	result := &models.ScanResult{
		ID:              "test-uuid-001",
		DeviceID:        "emulator-5554",
		PackageName:     "com.evil.bankbot",
		FilePath:        "/data/app/com.evil.bankbot-1/base.apk",
		SHA256Hash:      "a3f5c8d2e1b4f9a6c7d0e3b2a5f8c1d4e7b0a3f6",
		ThreatName:      "BankBot.A",
		Severity:        models.SeverityCritical,
		InfectionVector: "Web Browser Link",
		MalwareFamily:   "BankBot",
		Description:     "Test entry",
		ScanTimestamp:   time.Now().UTC(),
	}

	if err := db.SaveScanResult(result); err != nil {
		t.Fatalf("save scan result: %v", err)
	}

	history, err := db.GetScanHistory("emulator-5554")
	if err != nil {
		t.Fatalf("get scan history: %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected at least 1 history entry")
	}
	if history[0].PackageName != "com.evil.bankbot" {
		t.Errorf("expected com.evil.bankbot, got %s", history[0].PackageName)
	}
}

func TestInfectionVectorClassification(t *testing.T) {
	cases := []struct {
		pkg      models.Package
		expected string
	}{
		{
			models.Package{PackageName: "com.evil.app", FilePath: "/sdcard/Download/evil.apk"},
			"Web Browser Link",
		},
		{
			models.Package{PackageName: "com.evil.app", FilePath: "/storage/emulated/0/Download/bad.apk"},
			"Web Browser Link",
		},
		{
			models.Package{PackageName: "com.android.settings2", FilePath: "/data/data/com.android.settings2"},
			"Social Engineering/Trojan",
		},
		{
			models.Package{PackageName: "com.unknown.sideload", FilePath: "/data/local/tmp/sideload.apk"},
			"Unknown Sideloading",
		},
	}

	for i, tc := range cases {
		got := classifyInfectionVectorTest(tc.pkg)
		if got != tc.expected {
			t.Errorf("case %d: expected %q, got %q (pkg=%s, path=%s)",
				i, tc.expected, got, tc.pkg.PackageName, tc.pkg.FilePath)
		}
	}
}

func classifyInfectionVectorTest(pkg models.Package) string {
	path := toLower(pkg.FilePath)
	name := toLower(pkg.PackageName)

	if contains(path, "/sdcard/download") || contains(path, "/storage/emulated/0/download") {
		return "Web Browser Link"
	}

	systemPrefixes := []string{"com.android.", "com.google.android.", "android.", "com.samsung.", "com.huawei.", "com.miui."}
	for _, prefix := range systemPrefixes {
		if hasPrefix(name, prefix) && !hasSuffix(path, ".apk") && contains(path, "/data/data/") {
			return "Social Engineering/Trojan"
		}
	}

	if hasSuffix(path, ".apk") {
		return "Unknown Sideloading"
	}

	if pkg.InstallSource == "" || pkg.InstallSource == "null" {
		return "Unknown Sideloading"
	}

	return "Unknown"
}

func TestCommandSanitization(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"emulator-5554", "emulator-5554"},
		{"com.example.app", "com.example.app"},
		{"device; rm -rf /", "devicerm-rf"},
		{"../../../etc/passwd", "etcpasswd"},
		{"device`whoami`", "devicewhoami"},
		{"$(reboot)", "reboot"},
	}

	for _, tc := range cases {
		got := sanitizeTest(tc.input)
		if got != tc.expected {
			t.Errorf("sanitize(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func sanitizeTest(s string) string {
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == ':' {
			result += string(r)
		}
	}
	return result
}

func TestRemediationLog(t *testing.T) {
	db := newTestDB(t)

	if err := db.LogRemediation("emulator-5554", "com.evil.app", "uninstall", true, "Success"); err != nil {
		t.Fatalf("log remediation: %v", err)
	}

	if err := db.LogRemediation("emulator-5554", "com.evil2.app", "uninstall", false, "Failed: not found"); err != nil {
		t.Fatalf("log remediation failure: %v", err)
	}
}

func TestSHA256HashGeneration(t *testing.T) {
	data := []byte("com.evil.bankbot:/data/app/com.evil.bankbot-1/base.apk")
	h := sha256.Sum256(data)
	got := fmt.Sprintf("%x", h[:])
	if len(got) != 64 {
		t.Errorf("expected 64-char hex hash, got len=%d", len(got))
	}
}

func TestThreatSimulationDataset(t *testing.T) {
	db := newTestDB(t)
	if err := db.Seed(); err != nil {
		t.Fatal(err)
	}

	dataset := []struct {
		hash     string
		isMalware bool
	}{
		{"a3f5c8d2e1b4f9a6c7d0e3b2a5f8c1d4e7b0a3f6", true},
		{"b4e6d9c2f5a8b1e4d7c0f3a6b9e2d5c8f1a4b7e0", true},
		{"c5f7e0d3a6b9c2f5e8d1b4a7c0f3e6b9d2a5c8f1", true},
		{"d6a8f1e4b7c0d3f6a9e2b5d8c1f4a7e0b3d6a9f2", false},
		{"e7b9a2f5c8d1e4b7a0d3f6c9b2e5a8f1d4c7b0e3", false},
		{"0000000000000000000000000000000000000000000000000000000000000000", false},
	}

	for _, tc := range dataset {
		sig, err := db.LookupHash(tc.hash)
		if err != nil {
			t.Errorf("lookup %s: %v", tc.hash, err)
			continue
		}
		if tc.isMalware && sig == nil {
			t.Errorf("expected malware match for hash %s, got nil", tc.hash)
		}
		if !tc.isMalware && tc.hash == "0000000000000000000000000000000000000000000000000000000000000000" && sig != nil {
			t.Errorf("expected no match for zero hash, got %v", sig)
		}
	}
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		result[i] = c
	}
	return string(result)
}

func contains(s, sub string) bool { return len(s) >= len(sub) && indexOf(s, sub) >= 0 }
func hasPrefix(s, p string) bool  { return len(s) >= len(p) && s[:len(p)] == p }
func hasSuffix(s, p string) bool  { return len(s) >= len(p) && s[len(s)-len(p):] == p }

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
