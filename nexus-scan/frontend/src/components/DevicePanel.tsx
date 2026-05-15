import React from "react";
import type { Device, ScanSession } from "../types";

interface Props {
  devices: Device[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onStartScan: (id: string) => void;
  onStopScan: (id: string) => void;
  scanSession: ScanSession | null;
}

export function DevicePanel({ devices, selectedId, onSelect, onStartScan, onStopScan, scanSession }: Props) {
  const isScanning = scanSession?.status === "scanning";

  return (
    <div style={s.panel}>
      <div style={s.header}>
        <span style={s.title}>CONNECTED DEVICES</span>
        <span style={{ ...s.badge, background: devices.length > 0 ? "var(--green-subtle)" : "transparent", color: devices.length > 0 ? "var(--green)" : "var(--text-muted)" }}>
          {devices.length}
        </span>
      </div>

      <div style={s.list}>
        {devices.length === 0 ? (
          <div style={s.empty}>
            <div style={s.emptyIcon}>&#x25CB;</div>
            <div style={s.emptyText}>No devices detected</div>
            <div style={s.emptyHint}>Connect via USB and ensure adb/libimobiledevice is running</div>
          </div>
        ) : (
          devices.map((device) => (
            <DeviceCard
              key={device.id}
              device={device}
              selected={device.id === selectedId}
              scanning={isScanning && scanSession?.device_id === device.id}
              onSelect={() => onSelect(device.id)}
              onStartScan={() => onStartScan(device.id)}
              onStopScan={() => onStopScan(device.id)}
              scanSession={scanSession?.device_id === device.id ? scanSession : null}
            />
          ))
        )}
      </div>
    </div>
  );
}

function DeviceCard({
  device,
  selected,
  scanning,
  onSelect,
  onStartScan,
  onStopScan,
  scanSession,
}: {
  device: Device;
  selected: boolean;
  scanning: boolean;
  onSelect: () => void;
  onStartScan: () => void;
  onStopScan: () => void;
  scanSession: ScanSession | null;
}) {
  const isAndroid = device.type === "android";

  return (
    <div
      style={{
        ...s.card,
        borderColor: selected ? "var(--green-dim)" : "var(--border)",
        background: selected ? "var(--green-subtle)" : "var(--bg-card)",
      }}
      onClick={onSelect}
    >
      <div style={s.cardHeader}>
        <span style={{ ...s.deviceIcon, color: isAndroid ? "#3ddc84" : "#aaaaff" }}>
          {isAndroid ? "⬡" : ""}
        </span>
        <div style={s.deviceInfo}>
          <div style={s.deviceName}>{device.name || device.model}</div>
          <div style={s.deviceMeta}>{device.os_version}</div>
        </div>
        <StatusDot status={device.status} scanning={scanning} />
      </div>

      <div style={s.deviceSerial}>
        <span style={s.label}>SERIAL</span>
        <span style={s.value}>{device.serial_number}</span>
      </div>
      <div style={s.deviceSerial}>
        <span style={s.label}>TYPE</span>
        <span style={{ ...s.value, color: isAndroid ? "#3ddc84" : "#aaaaff", textTransform: "uppercase" }}>
          {device.type}
        </span>
      </div>

      {scanSession && scanning && (
        <div style={s.progressWrap}>
          <div style={s.progressLabel}>
            <span>{scanSession.scanned_count} / {scanSession.total_packages || "?"}</span>
            <span style={{ color: "var(--red)" }}>{scanSession.threats_found} threats</span>
          </div>
          <div style={s.progressBar}>
            <div
              style={{
                ...s.progressFill,
                width: scanSession.total_packages > 0
                  ? `${Math.min(100, (scanSession.scanned_count / scanSession.total_packages) * 100)}%`
                  : "0%",
              }}
            />
          </div>
        </div>
      )}

      <div style={s.actions} onClick={(e) => e.stopPropagation()}>
        {!scanning ? (
          <button
            style={s.btnScan}
            onClick={onStartScan}
            disabled={device.status !== "connected"}
          >
            &#x25B6; SCAN
          </button>
        ) : (
          <button style={s.btnStop} onClick={onStopScan}>
            &#x25A0; STOP
          </button>
        )}
      </div>
    </div>
  );
}

function StatusDot({ status, scanning }: { status: string; scanning: boolean }) {
  let color = "var(--text-muted)";
  if (scanning) color = "var(--orange)";
  else if (status === "connected") color = "var(--green)";
  else if (status === "error") color = "var(--red)";

  return (
    <div style={{ width: 8, height: 8, borderRadius: "50%", background: color, flexShrink: 0, boxShadow: `0 0 4px ${color}` }} />
  );
}

const s: Record<string, React.CSSProperties> = {
  panel: { display: "flex", flexDirection: "column", overflow: "hidden" },
  header: {
    display: "flex", alignItems: "center", justifyContent: "space-between",
    padding: "10px 12px 8px",
    borderBottom: "1px solid var(--border)",
    background: "var(--bg-secondary)",
    flexShrink: 0,
  },
  title: { color: "var(--green)", fontSize: 10, letterSpacing: "0.12em", fontWeight: "bold" },
  badge: {
    fontSize: 11, padding: "1px 6px", borderRadius: 10,
    border: "1px solid currentColor",
  },
  list: { flex: 1, overflowY: "auto", padding: 8, display: "flex", flexDirection: "column", gap: 8 },
  empty: {
    display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center",
    padding: "32px 16px", gap: 8, color: "var(--text-muted)",
  },
  emptyIcon: { fontSize: 32 },
  emptyText: { fontSize: 12, color: "var(--text-secondary)" },
  emptyHint: { fontSize: 10, textAlign: "center", lineHeight: 1.6 },
  card: {
    border: "1px solid", borderRadius: "var(--radius)",
    padding: "10px 12px", cursor: "pointer",
    transition: "border-color 0.15s ease, background 0.15s ease",
    display: "flex", flexDirection: "column", gap: 6,
  },
  cardHeader: { display: "flex", alignItems: "center", gap: 8 },
  deviceIcon: { fontSize: 18, flexShrink: 0 },
  deviceInfo: { flex: 1, minWidth: 0 },
  deviceName: { fontSize: 12, fontWeight: "bold", color: "var(--text-primary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" },
  deviceMeta: { fontSize: 10, color: "var(--text-secondary)" },
  deviceSerial: { display: "flex", gap: 8, fontSize: 10 },
  label: { color: "var(--text-muted)", minWidth: 44 },
  value: { color: "var(--text-secondary)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" },
  progressWrap: { display: "flex", flexDirection: "column", gap: 3 },
  progressLabel: { display: "flex", justifyContent: "space-between", fontSize: 10, color: "var(--text-secondary)" },
  progressBar: { height: 3, background: "var(--bg-tertiary)", borderRadius: 2, overflow: "hidden" },
  progressFill: { height: "100%", background: "var(--green)", transition: "width 0.2s ease", borderRadius: 2 },
  actions: { display: "flex", gap: 6, marginTop: 2 },
  btnScan: {
    flex: 1, padding: "5px 0", fontSize: 10, letterSpacing: "0.1em",
    background: "transparent", border: "1px solid var(--green-dim)",
    color: "var(--green)", borderRadius: "var(--radius-sm)",
    transition: "background 0.15s",
  },
  btnStop: {
    flex: 1, padding: "5px 0", fontSize: 10, letterSpacing: "0.1em",
    background: "transparent", border: "1px solid var(--red)",
    color: "var(--red)", borderRadius: "var(--radius-sm)",
  },
};
