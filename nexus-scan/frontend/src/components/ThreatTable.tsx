import React, { useState } from "react";
import type { ScanResult, SeverityLevel } from "../types";

interface Props {
  threats: ScanResult[];
  selectedDeviceId: string | null;
  onRemediate: (deviceId: string, packageName: string) => void;
}

const SEVERITY_ORDER: Record<SeverityLevel, number> = {
  CRITICAL: 0, HIGH: 1, MEDIUM: 2, LOW: 3,
};

export function ThreatTable({ threats, selectedDeviceId, onRemediate }: Props) {
  const [filter, setFilter] = useState<SeverityLevel | "ALL">("ALL");
  const [confirmPurge, setConfirmPurge] = useState<{ deviceId: string; pkg: string } | null>(null);

  const filtered = threats
    .filter((t) => {
      if (selectedDeviceId && t.device_id !== selectedDeviceId) return false;
      if (filter !== "ALL" && t.severity !== filter) return false;
      return true;
    })
    .sort((a, b) => (SEVERITY_ORDER[a.severity] ?? 4) - (SEVERITY_ORDER[b.severity] ?? 4));

  const severityCounts = threats.reduce(
    (acc, t) => { acc[t.severity] = (acc[t.severity] ?? 0) + 1; return acc; },
    {} as Record<SeverityLevel, number>
  );

  function handlePurgeClick(deviceId: string, pkg: string) {
    setConfirmPurge({ deviceId, pkg });
  }

  function handleConfirm() {
    if (confirmPurge) {
      onRemediate(confirmPurge.deviceId, confirmPurge.pkg);
      setConfirmPurge(null);
    }
  }

  return (
    <div style={s.container}>
      <div style={s.header}>
        <span style={s.title}>THREAT INTELLIGENCE</span>
        <div style={s.headerRight}>
          {(["ALL", "CRITICAL", "HIGH", "MEDIUM", "LOW"] as const).map((sev) => (
            <button
              key={sev}
              style={{
                ...s.filterBtn,
                ...(filter === sev ? s.filterBtnActive : {}),
                ...(sev !== "ALL" ? { color: severityColor(sev) } : {}),
              }}
              onClick={() => setFilter(sev)}
            >
              {sev}
              {sev !== "ALL" && severityCounts[sev] ? (
                <span style={{ ...s.filterCount, color: severityColor(sev) }}>
                  {severityCounts[sev]}
                </span>
              ) : null}
            </button>
          ))}
          <span style={s.totalCount}>{filtered.length} entries</span>
        </div>
      </div>

      <div style={s.tableWrapper}>
        {filtered.length === 0 ? (
          <div style={s.empty}>
            <div style={s.emptyIcon}>&#x25A1;</div>
            <div style={s.emptyText}>
              {threats.length === 0
                ? "No threats detected — run a scan to begin"
                : "No threats match the current filter"}
            </div>
          </div>
        ) : (
          <table style={s.table}>
            <thead>
              <tr style={s.theadRow}>
                {["SEVERITY", "THREAT NAME", "PACKAGE NAME", "FILE PATH", "SHA256", "VECTOR", "ACTION"].map((col) => (
                  <th key={col} style={s.th}>{col}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {filtered.map((threat) => (
                <ThreatRow
                  key={threat.id}
                  threat={threat}
                  onPurge={() => handlePurgeClick(threat.device_id, threat.package_name)}
                />
              ))}
            </tbody>
          </table>
        )}
      </div>

      {confirmPurge && (
        <ConfirmDialog
          packageName={confirmPurge.pkg}
          onConfirm={handleConfirm}
          onCancel={() => setConfirmPurge(null)}
        />
      )}
    </div>
  );
}

function ThreatRow({ threat, onPurge }: { threat: ScanResult; onPurge: () => void }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <>
      <tr
        style={{
          ...s.tr,
          background: threat.severity === "CRITICAL"
            ? "rgba(255,45,45,0.04)"
            : threat.severity === "HIGH"
            ? "rgba(255,140,0,0.03)"
            : undefined,
        }}
        onClick={() => setExpanded((e) => !e)}
      >
        <td style={s.td}>
          <SeverityBadge severity={threat.severity} />
        </td>
        <td style={{ ...s.td, color: "var(--text-primary)", fontWeight: "bold" }}>
          {threat.threat_name || "Suspicious.Package"}
        </td>
        <td style={{ ...s.td, color: "var(--text-secondary)", maxWidth: 180, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
          {threat.package_name}
        </td>
        <td style={{ ...s.td, color: "var(--text-muted)", maxWidth: 140, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", fontSize: 10 }}>
          {threat.file_path}
        </td>
        <td style={{ ...s.td, fontFamily: "var(--font-mono)", fontSize: 9, color: "var(--text-muted)", maxWidth: 90, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
          {threat.sha256_hash.slice(0, 16)}…
        </td>
        <td style={{ ...s.td, color: "var(--yellow)", fontSize: 10 }}>
          {threat.infection_vector || "Unknown"}
        </td>
        <td style={s.td} onClick={(e) => e.stopPropagation()}>
          <button style={s.purgeBtn} onClick={onPurge}>
            &#x2717; PURGE
          </button>
        </td>
      </tr>
      {expanded && (
        <tr style={s.expandedRow}>
          <td colSpan={7} style={s.expandedCell}>
            <div style={s.expandedContent}>
              <div><span style={s.expLabel}>FULL HASH</span> <span style={s.expValue}>{threat.sha256_hash}</span></div>
              {threat.malware_family && (
                <div><span style={s.expLabel}>FAMILY</span> <span style={s.expValue}>{threat.malware_family}</span></div>
              )}
              {threat.description && (
                <div><span style={s.expLabel}>DESCRIPTION</span> <span style={s.expValue}>{threat.description}</span></div>
              )}
              <div><span style={s.expLabel}>DEVICE ID</span> <span style={s.expValue}>{threat.device_id}</span></div>
              <div><span style={s.expLabel}>SCANNED AT</span> <span style={s.expValue}>{new Date(threat.scan_timestamp).toLocaleString()}</span></div>
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

function SeverityBadge({ severity }: { severity: SeverityLevel }) {
  return (
    <span style={{ ...s.severityBadge, color: severityColor(severity), borderColor: severityColor(severity) }}>
      {severity}
    </span>
  );
}

function ConfirmDialog({ packageName, onConfirm, onCancel }: { packageName: string; onConfirm: () => void; onCancel: () => void }) {
  return (
    <div style={s.dialogOverlay}>
      <div style={s.dialog}>
        <div style={s.dialogTitle}>&#x26A0; CONFIRM REMEDIATION</div>
        <div style={s.dialogBody}>
          <div style={s.dialogText}>Permanently uninstall:</div>
          <div style={s.dialogPkg}>{packageName}</div>
          <div style={s.dialogWarn}>This action cannot be undone. The app will be removed from the device.</div>
        </div>
        <div style={s.dialogActions}>
          <button style={s.dialogCancel} onClick={onCancel}>CANCEL</button>
          <button style={s.dialogConfirm} onClick={onConfirm}>PURGE</button>
        </div>
      </div>
    </div>
  );
}

function severityColor(sev: SeverityLevel | "ALL"): string {
  switch (sev) {
    case "CRITICAL": return "var(--red)";
    case "HIGH":     return "var(--orange)";
    case "MEDIUM":   return "var(--yellow)";
    case "LOW":      return "var(--text-secondary)";
    default:         return "var(--text-primary)";
  }
}

const s: Record<string, React.CSSProperties> = {
  container: { flex: 1, display: "flex", flexDirection: "column", overflow: "hidden", borderBottom: "1px solid var(--border)" },
  header: {
    display: "flex", alignItems: "center", justifyContent: "space-between",
    padding: "8px 14px", borderBottom: "1px solid var(--border)",
    background: "var(--bg-secondary)", flexShrink: 0,
  },
  title: { color: "var(--green)", fontSize: 10, letterSpacing: "0.12em", fontWeight: "bold" },
  headerRight: { display: "flex", alignItems: "center", gap: 6 },
  filterBtn: {
    background: "transparent", border: "1px solid var(--border)", borderRadius: "var(--radius-sm)",
    color: "var(--text-muted)", fontSize: 9, padding: "2px 7px", letterSpacing: "0.07em",
    cursor: "pointer", display: "flex", alignItems: "center", gap: 4,
  },
  filterBtnActive: { background: "var(--bg-tertiary)", borderColor: "var(--border-bright)", color: "var(--text-primary)" },
  filterCount: { fontWeight: "bold" },
  totalCount: { fontSize: 10, color: "var(--text-muted)", marginLeft: 4 },
  tableWrapper: { flex: 1, overflowY: "auto", overflowX: "auto" },
  empty: { display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", height: "100%", gap: 10, color: "var(--text-muted)" },
  emptyIcon: { fontSize: 36 },
  emptyText: { fontSize: 12, color: "var(--text-secondary)", textAlign: "center", maxWidth: 300 },
  table: { width: "100%", borderCollapse: "collapse", fontSize: 11 },
  theadRow: { background: "var(--bg-secondary)", position: "sticky", top: 0, zIndex: 1 },
  th: { padding: "6px 10px", textAlign: "left", color: "var(--text-muted)", fontSize: 9, letterSpacing: "0.1em", borderBottom: "1px solid var(--border)", fontWeight: "normal", whiteSpace: "nowrap" },
  tr: { borderBottom: "1px solid var(--border)", cursor: "pointer", transition: "background 0.1s" },
  td: { padding: "7px 10px", verticalAlign: "middle" },
  severityBadge: { fontSize: 9, padding: "1px 5px", border: "1px solid", borderRadius: 2, letterSpacing: "0.05em", fontWeight: "bold" },
  purgeBtn: {
    background: "transparent", border: "1px solid var(--red)", color: "var(--red)",
    borderRadius: "var(--radius-sm)", padding: "3px 8px", fontSize: 10, letterSpacing: "0.07em",
    cursor: "pointer", whiteSpace: "nowrap",
  },
  expandedRow: { background: "var(--bg-card)" },
  expandedCell: { padding: "0 10px 10px 36px" },
  expandedContent: { display: "flex", flexDirection: "column", gap: 4, paddingTop: 6 },
  expLabel: { color: "var(--text-muted)", fontSize: 9, letterSpacing: "0.1em", marginRight: 8 },
  expValue: { color: "var(--text-secondary)", fontSize: 10 },
  dialogOverlay: {
    position: "fixed", inset: 0, background: "rgba(0,0,0,0.75)",
    display: "flex", alignItems: "center", justifyContent: "center", zIndex: 100,
  },
  dialog: {
    background: "var(--bg-secondary)", border: "1px solid var(--red)",
    borderRadius: "var(--radius)", padding: "24px", minWidth: 340, maxWidth: 420,
    boxShadow: "0 0 30px rgba(255,45,45,0.2)",
  },
  dialogTitle: { color: "var(--red)", fontSize: 13, letterSpacing: "0.1em", marginBottom: 16, fontWeight: "bold" },
  dialogBody: { display: "flex", flexDirection: "column", gap: 8, marginBottom: 20 },
  dialogText: { color: "var(--text-secondary)", fontSize: 12 },
  dialogPkg: { color: "var(--text-primary)", fontWeight: "bold", fontSize: 13, padding: "6px 10px", background: "var(--bg-tertiary)", borderRadius: "var(--radius-sm)", wordBreak: "break-all" },
  dialogWarn: { color: "var(--text-muted)", fontSize: 11 },
  dialogActions: { display: "flex", gap: 10, justifyContent: "flex-end" },
  dialogCancel: {
    background: "transparent", border: "1px solid var(--border)", color: "var(--text-secondary)",
    borderRadius: "var(--radius-sm)", padding: "7px 18px", fontSize: 11, cursor: "pointer",
  },
  dialogConfirm: {
    background: "var(--red)", border: "1px solid var(--red)", color: "#fff",
    borderRadius: "var(--radius-sm)", padding: "7px 18px", fontSize: 11, cursor: "pointer",
    fontWeight: "bold", letterSpacing: "0.05em",
  },
};
