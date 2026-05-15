package scanner

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-scan/nexus-scan/pkg/database"
	"github.com/nexus-scan/nexus-scan/pkg/ipc"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/models"
)

type Scanner struct {
	db          *database.DB
	broadcaster *ipc.Broadcaster
	log         *logger.Logger
	workerCount int
	activeMu    sync.Mutex
	activeScan  map[string]context.CancelFunc
}

func New(db *database.DB, broadcaster *ipc.Broadcaster, log *logger.Logger, workerCount int) *Scanner {
	if workerCount <= 0 {
		workerCount = 8
	}
	return &Scanner{
		db:          db,
		broadcaster: broadcaster,
		log:         log.WithModule("scanner"),
		workerCount: workerCount,
		activeScan:  make(map[string]context.CancelFunc),
	}
}

func (s *Scanner) StartScan(deviceID string) {
	s.activeMu.Lock()
	if _, running := s.activeScan[deviceID]; running {
		s.activeMu.Unlock()
		s.log.Warn("scan already in progress", "device_id", deviceID)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	s.activeScan[deviceID] = cancel
	s.activeMu.Unlock()

	defer func() {
		s.activeMu.Lock()
		delete(s.activeScan, deviceID)
		s.activeMu.Unlock()
		cancel()
	}()

	s.broadcaster.Broadcast(ipc.Event{
		Type:      "scan_started",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload:   map[string]interface{}{"device_id": deviceID},
	})

	packages, err := s.listPackages(ctx, deviceID)
	if err != nil {
		s.log.Error("failed to list packages", "device_id", deviceID, "error", err)
		s.broadcaster.Broadcast(ipc.Event{
			Type:      "scan_error",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   map[string]interface{}{"device_id": deviceID, "error": err.Error()},
		})
		return
	}

	s.log.Info("scan: packages enumerated", "device_id", deviceID, "count", len(packages))
	s.broadcaster.Broadcast(ipc.Event{
		Type:      "scan_progress",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"device_id":      deviceID,
			"total_packages": len(packages),
			"scanned":        0,
		},
	})

	type job struct {
		pkg models.Package
	}

	jobs := make(chan job, len(packages))
	for _, pkg := range packages {
		jobs <- job{pkg: pkg}
	}
	close(jobs)

	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		threatsFound int
		scanned      int
	)

	for w := 0; w < s.workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				result := s.analyzePackage(deviceID, j.pkg)
				if result == nil {
					continue
				}

				if err := s.db.SaveScanResult(result); err != nil {
					s.log.Error("failed to save scan result", "error", err, "package", j.pkg.PackageName)
				}

				mu.Lock()
				scanned++
				currentScanned := scanned
				if result.ThreatName != "" {
					threatsFound++
				}
				mu.Unlock()

				payload := map[string]interface{}{
					"device_id":    deviceID,
					"package_name": result.PackageName,
					"file_path":    result.FilePath,
					"sha256_hash":  result.SHA256Hash,
					"scanned":      currentScanned,
					"total":        len(packages),
				}

				s.broadcaster.Broadcast(ipc.Event{
					Type:      "package_scanned",
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Payload:   payload,
				})

				if result.ThreatName != "" {
					s.log.ThreatEvent(deviceID, result.PackageName, result.ThreatName, string(result.Severity))
					s.broadcaster.Broadcast(ipc.Event{
						Type:      "threat_detected",
						Timestamp: time.Now().UTC().Format(time.RFC3339),
						Payload: map[string]interface{}{
							"device_id":       deviceID,
							"package_name":    result.PackageName,
							"file_path":       result.FilePath,
							"sha256_hash":     result.SHA256Hash,
							"threat_name":     result.ThreatName,
							"severity":        string(result.Severity),
							"infection_vector": result.InfectionVector,
							"malware_family":  result.MalwareFamily,
							"description":     result.Description,
						},
					})
				}
			}
		}()
	}

	wg.Wait()

	s.broadcaster.Broadcast(ipc.Event{
		Type:      "scan_finished",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"device_id":     deviceID,
			"total_scanned": scanned,
			"threats_found": threatsFound,
		},
	})

	s.log.Info("scan complete", "device_id", deviceID, "scanned", scanned, "threats", threatsFound)
}

func (s *Scanner) StopScan(deviceID string) {
	s.activeMu.Lock()
	defer s.activeMu.Unlock()
	if cancel, ok := s.activeScan[deviceID]; ok {
		cancel()
		s.log.Info("scan stopped by user", "device_id", deviceID)
	}
}

func (s *Scanner) GetScanHistory(deviceID string) ([]models.ScanResult, error) {
	return s.db.GetScanHistory(deviceID)
}

func (s *Scanner) listPackages(ctx context.Context, deviceID string) ([]models.Package, error) {
	cmd := exec.CommandContext(ctx, "adb", "-s", sanitize(deviceID), "shell", "pm", "list", "packages", "-f", "-i")
	out, err := cmd.Output()
	if err != nil {
		return s.listPackagesIOS(ctx, deviceID)
	}
	return parseAndroidPackages(ctx, deviceID, out), nil
}

func parseAndroidPackages(ctx context.Context, deviceID string, out []byte) []models.Package {
	var packages []models.Package
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "package:") {
			continue
		}
		line = strings.TrimPrefix(line, "package:")
		eqIdx := strings.LastIndex(line, "=")
		if eqIdx < 0 {
			continue
		}
		filePath := strings.TrimSpace(line[:eqIdx])
		rest := strings.TrimSpace(line[eqIdx+1:])
		parts := strings.Fields(rest)
		if len(parts) == 0 {
			continue
		}
		pkgName := parts[0]

		installSource := ""
		for _, p := range parts[1:] {
			if strings.HasPrefix(p, "installer=") {
				installSource = strings.TrimPrefix(p, "installer=")
			}
		}

		packages = append(packages, models.Package{
			PackageName:   pkgName,
			FilePath:      filePath,
			InstallSource: installSource,
		})
	}
	return packages
}

func (s *Scanner) listPackagesIOS(ctx context.Context, deviceID string) ([]models.Package, error) {
	cmd := exec.CommandContext(ctx, "ideviceinstaller", "-u", sanitize(deviceID), "-l")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listPackages: neither adb nor ideviceinstaller available for %s", deviceID)
	}

	var packages []models.Package
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "Total:") || strings.HasPrefix(line, "CFBundleIdentifier") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 1 {
			continue
		}
		packages = append(packages, models.Package{
			PackageName:   parts[0],
			FilePath:      "/private/var/containers/Bundle/Application/" + parts[0],
			InstallSource: "AppStore",
		})
	}
	return packages, nil
}

func (s *Scanner) analyzePackage(deviceID string, pkg models.Package) *models.ScanResult {
	if pkg.PackageName == "" {
		return nil
	}

	hash := computePackageHash(deviceID, pkg)

	sig, err := s.db.LookupHash(hash)
	if err != nil {
		s.log.Error("hash lookup failed", "error", err, "package", pkg.PackageName)
	}

	result := &models.ScanResult{
		ID:            uuid.New().String(),
		DeviceID:      deviceID,
		PackageName:   pkg.PackageName,
		FilePath:      pkg.FilePath,
		SHA256Hash:    hash,
		ScanTimestamp: time.Now().UTC(),
	}

	if sig != nil {
		result.ThreatName = sig.MalwareName
		result.Severity = sig.Severity
		result.MalwareFamily = sig.MalwareFamily
		result.Description = sig.Description
		result.InfectionVector = classifyInfectionVector(pkg)
		return result
	}

	if isSuspiciousPackage(pkg) {
		result.ThreatName = "Suspicious.Package"
		result.Severity = models.SeverityMedium
		result.InfectionVector = classifyInfectionVector(pkg)
		result.Description = "Package exhibits suspicious naming or installation characteristics."
	}

	return result
}

func computePackageHash(deviceID string, pkg models.Package) string {
	h := sha256.New()
	h.Write([]byte(deviceID))
	h.Write([]byte(pkg.PackageName))
	h.Write([]byte(pkg.FilePath))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func classifyInfectionVector(pkg models.Package) string {
	path := strings.ToLower(pkg.FilePath)
	name := strings.ToLower(pkg.PackageName)

	if strings.Contains(path, "/sdcard/download") ||
		strings.Contains(path, "/storage/emulated/0/download") {
		return "Web Browser Link"
	}

	systemPrefixes := []string{
		"com.android.", "com.google.android.", "android.",
		"com.samsung.", "com.huawei.", "com.miui.",
	}
	for _, prefix := range systemPrefixes {
		if strings.HasPrefix(name, prefix) && !strings.HasSuffix(path, ".apk") &&
			strings.Contains(path, "/data/data/") {
			return "Social Engineering/Trojan"
		}
	}

	if strings.HasSuffix(path, ".apk") {
		return "Unknown Sideloading"
	}

	if pkg.InstallSource == "" || pkg.InstallSource == "null" {
		return "Unknown Sideloading"
	}

	return "Unknown"
}

func isSuspiciousPackage(pkg models.Package) bool {
	name := strings.ToLower(pkg.PackageName)
	path := strings.ToLower(pkg.FilePath)

	suspiciousNames := []string{
		"whatsapp.clone", "facebook.lite2", "instagram.pro",
		"com.android.system2", "com.google.service2",
		"flashlight.plus", "battery.saver.pro",
	}
	for _, sn := range suspiciousNames {
		if strings.Contains(name, sn) {
			return true
		}
	}

	if strings.Contains(path, "/sdcard/download") && strings.HasSuffix(path, ".apk") {
		return true
	}

	systemLike := []string{"com.android.", "com.google.", "android."}
	for _, prefix := range systemLike {
		if strings.HasPrefix(name, prefix) {
			if strings.Contains(path, "/sdcard/") || strings.Contains(path, "/storage/") {
				return true
			}
		}
	}

	return false
}

func sanitize(s string) string {
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == ':' {
			return r
		}
		return -1
	}, s)
	return result
}
