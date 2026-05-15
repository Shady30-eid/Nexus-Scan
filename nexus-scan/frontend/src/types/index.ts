export type DeviceType = "android" | "ios";
export type SeverityLevel = "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";
export type DeviceStatus = "connected" | "disconnected" | "scanning" | "error";

export interface Device {
  id: string;
  serial_number: string;
  name: string;
  model: string;
  os_version: string;
  type: DeviceType;
  status: DeviceStatus;
  connected_at: string;
  vendor_id?: string;
  product_id?: string;
}

export interface ScanResult {
  id: string;
  device_id: string;
  package_name: string;
  file_path: string;
  sha256_hash: string;
  threat_name: string;
  severity: SeverityLevel;
  infection_vector: string;
  scan_timestamp: string;
  malware_family?: string;
  description?: string;
}

export interface LogEntry {
  timestamp: string;
  level: "INFO" | "WARNING" | "ERROR" | "THREAT";
  message: string;
  device_id?: string;
  package_name?: string;
  module?: string;
}

export interface ScanSession {
  device_id: string;
  total_packages: number;
  scanned_count: number;
  threats_found: number;
  status: "idle" | "scanning" | "finished" | "error";
}

export type WsStatus = "connecting" | "connected" | "disconnected" | "error";

export interface WsEvent {
  type: string;
  timestamp: string;
  payload: Record<string, unknown>;
}
