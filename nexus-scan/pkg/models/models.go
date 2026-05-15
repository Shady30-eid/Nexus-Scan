package models

import "time"

type DeviceType string
type SeverityLevel string
type DeviceStatus string

const (
	DeviceTypeAndroid DeviceType = "android"
	DeviceTypeIOS     DeviceType = "ios"

	SeverityLow      SeverityLevel = "LOW"
	SeverityMedium   SeverityLevel = "MEDIUM"
	SeverityHigh     SeverityLevel = "HIGH"
	SeverityCritical SeverityLevel = "CRITICAL"

	DeviceStatusConnected    DeviceStatus = "connected"
	DeviceStatusDisconnected DeviceStatus = "disconnected"
	DeviceStatusScanning     DeviceStatus = "scanning"
	DeviceStatusError        DeviceStatus = "error"
)

type Device struct {
	ID           string       `json:"id"`
	SerialNumber string       `json:"serial_number"`
	Name         string       `json:"name"`
	Model        string       `json:"model"`
	OSVersion    string       `json:"os_version"`
	Type         DeviceType   `json:"type"`
	Status       DeviceStatus `json:"status"`
	ConnectedAt  time.Time    `json:"connected_at"`
	VendorID     string       `json:"vendor_id"`
	ProductID    string       `json:"product_id"`
}

type Package struct {
	PackageName       string        `json:"package_name"`
	FilePath          string        `json:"file_path"`
	InstallSource     string        `json:"install_source"`
	Permissions       []string      `json:"permissions"`
	SHA256Hash        string        `json:"sha256_hash"`
	IsSystemApp       bool          `json:"is_system_app"`
	IsSuspicious      bool          `json:"is_suspicious"`
	SuspicionReason   string        `json:"suspicion_reason,omitempty"`
	ThreatName        string        `json:"threat_name,omitempty"`
	Severity          SeverityLevel `json:"severity,omitempty"`
	InfectionVector   string        `json:"infection_vector,omitempty"`
	MalwareFamily     string        `json:"malware_family,omitempty"`
}

type ScanResult struct {
	ID              string        `json:"id"`
	DeviceID        string        `json:"device_id"`
	PackageName     string        `json:"package_name"`
	FilePath        string        `json:"file_path"`
	SHA256Hash      string        `json:"sha256_hash"`
	ThreatName      string        `json:"threat_name"`
	Severity        SeverityLevel `json:"severity"`
	InfectionVector string        `json:"infection_vector"`
	ScanTimestamp   time.Time     `json:"scan_timestamp"`
	MalwareFamily   string        `json:"malware_family,omitempty"`
	Description     string        `json:"description,omitempty"`
}

type MalwareSignature struct {
	ID           int           `json:"id"`
	MalwareName  string        `json:"malware_name"`
	SHA256Hash   string        `json:"sha256_hash"`
	MalwareFamily string       `json:"malware_family"`
	Severity     SeverityLevel `json:"severity"`
	Description  string        `json:"description"`
	CreatedAt    time.Time     `json:"created_at"`
}

type ScanSession struct {
	DeviceID      string    `json:"device_id"`
	StartTime     time.Time `json:"start_time"`
	TotalPackages int       `json:"total_packages"`
	ScannedCount  int       `json:"scanned_count"`
	ThreatsFound  int       `json:"threats_found"`
	Status        string    `json:"status"`
}

type RemediationResult struct {
	DeviceID    string    `json:"device_id"`
	PackageName string    `json:"package_name"`
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}
