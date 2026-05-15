package remediation

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/nexus-scan/nexus-scan/pkg/database"
	"github.com/nexus-scan/nexus-scan/pkg/ipc"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/models"
)

type Service struct {
	db          *database.DB
	broadcaster *ipc.Broadcaster
	log         *logger.Logger
}

func New(db *database.DB, broadcaster *ipc.Broadcaster, log *logger.Logger) *Service {
	return &Service{
		db:          db,
		broadcaster: broadcaster,
		log:         log.WithModule("remediation"),
	}
}

func (s *Service) Uninstall(deviceID, packageName string) {
	s.log.Info("remediation started", "device_id", deviceID, "package", packageName)

	s.broadcaster.Broadcast(ipc.Event{
		Type:      "remediation_started",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Payload: map[string]interface{}{
			"device_id":    deviceID,
			"package_name": packageName,
		},
	})

	result := s.performUninstall(deviceID, packageName)

	if err := s.db.LogRemediation(deviceID, packageName, "uninstall", result.Success, result.Message); err != nil {
		s.log.Error("failed to log remediation", "error", err)
	}

	s.log.RemediationEvent(deviceID, packageName, "uninstall", result.Success)

	eventPayload := map[string]interface{}{
		"device_id":    deviceID,
		"package_name": packageName,
		"success":      result.Success,
		"message":      result.Message,
	}

	if result.Success {
		s.broadcaster.Broadcast(ipc.Event{
			Type:      "remediation_completed",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   eventPayload,
		})
	} else {
		s.broadcaster.Broadcast(ipc.Event{
			Type:      "remediation_failed",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   eventPayload,
		})
	}
}

func (s *Service) performUninstall(deviceID, packageName string) models.RemediationResult {
	safeDevice := sanitize(deviceID)
	safePkg := sanitizePackage(packageName)

	if safeDevice == "" {
		return models.RemediationResult{
			DeviceID:    deviceID,
			PackageName: packageName,
			Success:     false,
			Message:     "invalid device ID",
			Timestamp:   time.Now(),
		}
	}
	if safePkg == "" {
		return models.RemediationResult{
			DeviceID:    deviceID,
			PackageName: packageName,
			Success:     false,
			Message:     "invalid package name",
			Timestamp:   time.Now(),
		}
	}

	result := s.tryAndroidUninstall(safeDevice, safePkg)
	if result != nil {
		return *result
	}

	result = s.tryIOSUninstall(safeDevice, safePkg)
	if result != nil {
		return *result
	}

	return models.RemediationResult{
		DeviceID:    deviceID,
		PackageName: packageName,
		Success:     false,
		Message:     "no suitable uninstall method found — ensure adb or ideviceinstaller is available",
		Timestamp:   time.Now(),
	}
}

func (s *Service) tryAndroidUninstall(deviceID, packageName string) *models.RemediationResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", "-s", deviceID, "uninstall", packageName)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		if isCommandNotFound(err) {
			return nil
		}
		s.log.Error("adb uninstall failed", "device_id", deviceID, "package", packageName, "error", err, "output", output)
		return &models.RemediationResult{
			DeviceID:    deviceID,
			PackageName: packageName,
			Success:     false,
			Message:     fmt.Sprintf("adb uninstall failed: %s", output),
			Timestamp:   time.Now(),
		}
	}

	success := strings.Contains(output, "Success")
	msg := output
	if success {
		msg = fmt.Sprintf("Successfully uninstalled %s", packageName)
	}

	return &models.RemediationResult{
		DeviceID:    deviceID,
		PackageName: packageName,
		Success:     success,
		Message:     msg,
		Timestamp:   time.Now(),
	}
}

func (s *Service) tryIOSUninstall(deviceID, packageName string) *models.RemediationResult {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ideviceinstaller", "-u", deviceID, "-U", packageName)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	if err != nil {
		if isCommandNotFound(err) {
			return nil
		}
		s.log.Error("ideviceinstaller uninstall failed", "device_id", deviceID, "package", packageName, "error", err, "output", output)
		return &models.RemediationResult{
			DeviceID:    deviceID,
			PackageName: packageName,
			Success:     false,
			Message:     fmt.Sprintf("ideviceinstaller uninstall failed: %s", output),
			Timestamp:   time.Now(),
		}
	}

	success := !strings.Contains(strings.ToLower(output), "error") &&
		!strings.Contains(strings.ToLower(output), "fail")

	msg := output
	if success {
		msg = fmt.Sprintf("Successfully uninstalled %s via ideviceinstaller", packageName)
	}

	return &models.RemediationResult{
		DeviceID:    deviceID,
		PackageName: packageName,
		Success:     success,
		Message:     msg,
		Timestamp:   time.Now(),
	}
}

func isCommandNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "no such file") ||
		strings.Contains(err.Error(), "not found")
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == ':' {
			return r
		}
		return -1
	}, s)
}

func sanitizePackage(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' {
			return r
		}
		return -1
	}, s)
}
