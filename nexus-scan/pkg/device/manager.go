package device

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/nexus-scan/nexus-scan/pkg/ipc"
	"github.com/nexus-scan/nexus-scan/pkg/logger"
	"github.com/nexus-scan/nexus-scan/pkg/models"
	"github.com/nexus-scan/nexus-scan/pkg/scanner"
)

type Manager struct {
	mu          sync.RWMutex
	devices     map[string]*models.Device
	broadcaster *ipc.Broadcaster
	scannerSvc  *scanner.Scanner
	log         *logger.Logger
}

func NewManager(broadcaster *ipc.Broadcaster, scannerSvc *scanner.Scanner, log *logger.Logger) *Manager {
	return &Manager{
		devices:     make(map[string]*models.Device),
		broadcaster: broadcaster,
		scannerSvc:  scannerSvc,
		log:         log.WithModule("device-manager"),
	}
}

func (m *Manager) StartPolling(ctx context.Context) {
	m.log.Info("device polling started")
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.log.Info("device polling stopped")
			return
		case <-ticker.C:
			m.pollAndroid()
			m.pollIOS()
		}
	}
}

func (m *Manager) ListDevices() []*models.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*models.Device, 0, len(m.devices))
	for _, d := range m.devices {
		list = append(list, d)
	}
	return list
}

func (m *Manager) pollAndroid() {
	cmd := exec.Command("adb", "devices", "-l")
	cmd.Env = append(cmd.Environ(), "ADB_VENDOR_KEYS=")
	out, err := cmd.Output()
	if err != nil {
		m.log.Warn("adb devices failed — adb not available or daemon not running", "error", err)
		m.markAllOffline(models.DeviceTypeAndroid)
		return
	}

	seen := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "List of") || strings.HasPrefix(line, "*") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		serial := fields[0]
		state := fields[1]
		if state != "device" {
			continue
		}

		seen[serial] = true
		m.mu.RLock()
		existing, exists := m.devices[serial]
		m.mu.RUnlock()

		if !exists {
			dev := m.queryAndroidDevice(serial)
			m.mu.Lock()
			m.devices[serial] = dev
			m.mu.Unlock()

			m.log.Info("android device connected", "serial", serial, "model", dev.Model)
			m.broadcaster.Broadcast(ipc.Event{
				Type:      "device_detected",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Payload: map[string]interface{}{
					"device": dev,
				},
			})
		} else if existing.Status != models.DeviceStatusConnected {
			m.mu.Lock()
			existing.Status = models.DeviceStatusConnected
			m.mu.Unlock()
			m.broadcaster.Broadcast(ipc.Event{
				Type:      "device_detected",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Payload:   map[string]interface{}{"device": existing},
			})
		}
	}

	m.mu.Lock()
	for serial, dev := range m.devices {
		if dev.Type == models.DeviceTypeAndroid && !seen[serial] {
			dev.Status = models.DeviceStatusDisconnected
			m.log.Info("android device disconnected", "serial", serial)
			m.broadcaster.Broadcast(ipc.Event{
				Type:      "device_disconnected",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Payload:   map[string]interface{}{"device_id": serial},
			})
			delete(m.devices, serial)
		}
	}
	m.mu.Unlock()
}

func (m *Manager) queryAndroidDevice(serial string) *models.Device {
	dev := &models.Device{
		ID:           serial,
		SerialNumber: serial,
		Type:         models.DeviceTypeAndroid,
		Status:       models.DeviceStatusConnected,
		ConnectedAt:  time.Now(),
	}

	props := map[string]string{
		"ro.product.model":         "",
		"ro.product.manufacturer":  "",
		"ro.build.version.release": "",
		"ro.product.name":          "",
	}

	for prop := range props {
		cmd := exec.Command("adb", "-s", serial, "shell", "getprop", sanitizeShellArg(prop))
		out, err := cmd.Output()
		if err == nil {
			props[prop] = strings.TrimSpace(string(out))
		}
	}

	manufacturer := props["ro.product.manufacturer"]
	model := props["ro.product.model"]
	if manufacturer != "" && model != "" {
		dev.Model = manufacturer + " " + model
	} else if model != "" {
		dev.Model = model
	} else {
		dev.Model = "Android Device"
	}
	dev.Name = dev.Model
	dev.OSVersion = "Android " + props["ro.build.version.release"]
	return dev
}

func (m *Manager) pollIOS() {
	cmd := exec.Command("idevice_id", "-l")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	seen := map[string]bool{}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		serial := strings.TrimSpace(line)
		if serial == "" {
			continue
		}
		seen[serial] = true
		m.mu.RLock()
		_, exists := m.devices[serial]
		m.mu.RUnlock()

		if !exists {
			dev := m.queryIOSDevice(serial)
			m.mu.Lock()
			m.devices[serial] = dev
			m.mu.Unlock()

			m.log.Info("ios device connected", "serial", serial)
			m.broadcaster.Broadcast(ipc.Event{
				Type:      "device_detected",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Payload:   map[string]interface{}{"device": dev},
			})
		}
	}

	m.mu.Lock()
	for serial, dev := range m.devices {
		if dev.Type == models.DeviceTypeIOS && !seen[serial] {
			dev.Status = models.DeviceStatusDisconnected
			m.broadcaster.Broadcast(ipc.Event{
				Type:      "device_disconnected",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Payload:   map[string]interface{}{"device_id": serial},
			})
			delete(m.devices, serial)
		}
	}
	m.mu.Unlock()
}

func (m *Manager) queryIOSDevice(serial string) *models.Device {
	dev := &models.Device{
		ID:           serial,
		SerialNumber: serial,
		Type:         models.DeviceTypeIOS,
		Status:       models.DeviceStatusConnected,
		ConnectedAt:  time.Now(),
		Model:        "iOS Device",
		Name:         "iOS Device",
		OSVersion:    "iOS",
	}

	cmd := exec.Command("ideviceinfo", "-u", sanitizeShellArg(serial), "-k", "ProductVersion")
	if out, err := cmd.Output(); err == nil {
		dev.OSVersion = "iOS " + strings.TrimSpace(string(out))
	}

	cmd2 := exec.Command("ideviceinfo", "-u", sanitizeShellArg(serial), "-k", "ProductType")
	if out, err := cmd2.Output(); err == nil {
		dev.Model = strings.TrimSpace(string(out))
		dev.Name = dev.Model
	}

	cmd3 := exec.Command("ideviceinfo", "-u", sanitizeShellArg(serial), "-k", "DeviceName")
	if out, err := cmd3.Output(); err == nil {
		dev.Name = strings.TrimSpace(string(out))
	}

	return dev
}

func (m *Manager) markAllOffline(deviceType models.DeviceType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for serial, dev := range m.devices {
		if dev.Type == deviceType {
			dev.Status = models.DeviceStatusDisconnected
			delete(m.devices, serial)
		}
	}
}

// sanitizeShellArg prevents shell injection by rejecting dangerous characters.
func sanitizeShellArg(arg string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == ':' {
			return r
		}
		return -1
	}, arg)
	return safe
}
