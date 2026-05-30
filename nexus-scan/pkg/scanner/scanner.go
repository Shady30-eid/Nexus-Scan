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

                                // Only persist results that have an identified threat
                                if result.ThreatName != "" {
                                        if err := s.db.SaveScanResult(result); err != nil {
                                                s.log.Error("failed to save scan result", "error", err, "package", j.pkg.PackageName)
                                        }
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

                // Skip packages with no real name (MIUI overlay APKs report "null")
                if pkgName == "" || pkgName == "null" {
                        continue
                }

                installSource := ""
                for _, p := range parts[1:] {
                        if strings.HasPrefix(p, "installer=") {
                                src := strings.TrimPrefix(p, "installer=")
                                if src != "null" {
                                        installSource = src
                                }
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

        result := &models.ScanResult{
                ID:            uuid.New().String(),
                DeviceID:      deviceID,
                PackageName:   pkg.PackageName,
                FilePath:      pkg.FilePath,
                SHA256Hash:    hash,
                ScanTimestamp: time.Now().UTC(),
        }

        // ── 1. Check exact package name against malware database ─────────────
        pkgSig, err := s.db.LookupPackageName(pkg.PackageName)
        if err != nil {
                s.log.Error("package name lookup failed", "error", err, "package", pkg.PackageName)
        }
        if pkgSig != nil {
                result.ThreatName      = pkgSig.MalwareName
                result.Severity        = pkgSig.Severity
                result.MalwareFamily   = pkgSig.MalwareFamily
                result.Description     = pkgSig.Description
                result.InfectionVector = classifyInfectionVector(pkg)
                return result
        }

        // ── 2. Check APK hash against hash-based signature database ──────────
        hashSig, err := s.db.LookupHash(hash)
        if err != nil {
                s.log.Error("hash lookup failed", "error", err, "package", pkg.PackageName)
        }
        if hashSig != nil {
                result.ThreatName      = hashSig.MalwareName
                result.Severity        = hashSig.Severity
                result.MalwareFamily   = hashSig.MalwareFamily
                result.Description     = hashSig.Description
                result.InfectionVector = classifyInfectionVector(pkg)
                return result
        }

        // ── 3. Heuristic analysis ─────────────────────────────────────────────
        if threat, sev, desc := heuristicAnalysis(pkg); threat != "" {
                result.ThreatName      = threat
                result.Severity        = sev
                result.Description     = desc
                result.InfectionVector = classifyInfectionVector(pkg)
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

// heuristicAnalysis returns (threatName, severity, description) based on
// behavioral and naming patterns when no exact signature match is found.
func heuristicAnalysis(pkg models.Package) (string, models.SeverityLevel, string) {
        name := strings.ToLower(pkg.PackageName)
        path := strings.ToLower(pkg.FilePath)

        // ── High-confidence: sideloaded APK in Download folder ───────────────
        if (strings.Contains(path, "/sdcard/download") || strings.Contains(path, "/storage/emulated/0/download")) &&
                strings.HasSuffix(path, ".apk") {
                return "Suspicious.Sideload", models.SeverityHigh,
                        "APK sideloaded from Downloads folder — not installed via official store."
        }

        // ── Fake system/Google app impersonation ──────────────────────────────
        systemPrefixes := []string{"com.android.", "com.google.", "android.", "com.samsung.", "com.miui."}
        for _, pfx := range systemPrefixes {
                if strings.HasPrefix(name, pfx) {
                        if strings.Contains(path, "/sdcard/") || strings.Contains(path, "/storage/") {
                                return "Suspicious.FakeSystem", models.SeverityCritical,
                                        "App impersonates system package but is installed on external storage — likely trojan."
                        }
                }
        }

        // ── Duplicate suffix impersonation (e.g. com.whatsapp2, org.telegram.update) ──
        suspSuffixes := []string{"2", "3", "update", "pro", "plus", "lite2", "clone", "mod", "hack", "cracked", "patched", "unofficial"}
        for _, sfx := range suspSuffixes {
                if strings.HasSuffix(name, "."+sfx) || strings.HasSuffix(name, sfx) {
                        return "Suspicious.FakeApp", models.SeverityHigh,
                                "Package name suffix suggests fake/modified clone of a legitimate application."
                }
        }

        // ── Known malware keyword patterns ───────────────────────────────────
        malwareKeywords := map[string]string{
                "spyware":    "Keyword.Spyware",
                "keylogger":  "Keyword.Keylogger",
                "rootkit":    "Keyword.Rootkit",
                "botnet":     "Keyword.Botnet",
                "backdoor":   "Keyword.Backdoor",
                "rat.":       "Keyword.RAT",
                ".rat":       "Keyword.RAT",
                "stealthy":   "Keyword.Stealth",
                "covert":     "Keyword.Covert",
                "hiddenspy":  "Keyword.HiddenSpy",
                "silentspy":  "Keyword.SilentSpy",
        }
        for kw, threat := range malwareKeywords {
                if strings.Contains(name, kw) {
                        return threat, models.SeverityHigh,
                                "Package name contains known malware-related keyword: " + kw
                }
        }

        // ── Apps in user-space without a recognised official store ───────────
        if !strings.Contains(path, "/system/") && !strings.Contains(path, "/vendor/") &&
                !strings.Contains(path, "/priv-app/") {

                trustedInstallers := []string{
                        "com.android.vending",       // Google Play
                        "com.amazon.venezia",         // Amazon Appstore
                        "com.sec.android.app.samsungapps", // Galaxy Store
                        "com.huawei.appmarket",       // HUAWEI AppGallery
                        "com.android.packageinstaller", // Stock installer (APKs)
                }
                isTrusted := false
                for _, ti := range trustedInstallers {
                        if pkg.InstallSource == ti {
                                isTrusted = true
                                break
                        }
                }

                if !isTrusted && pkg.InstallSource != "" {
                        // Installed by an unofficial/OEM store (e.g. Xiaomi GetApps)
                        return "Suspicious.UnofficialStore", models.SeverityMedium,
                                "App installed via unofficial store (" + pkg.InstallSource + ") — unverified source."
                }

                if !isTrusted && pkg.InstallSource == "" {
                        // Sideloaded via ADB or manual APK install
                        return "Suspicious.Sideloaded", models.SeverityHigh,
                                "App has no install source record — likely sideloaded via ADB or manual APK."
                }
        }

        return "", "", ""
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
