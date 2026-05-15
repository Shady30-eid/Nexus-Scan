import React from "react";
import type { Device, ScanSession } from "../types";

interface Props {
  session: ScanSession | null;
  device: Device | null;
}

export function ScanStatusPanel({ session, device }: Props) {
  const progress =
    session && session.total_packages > 0
      ? Math.min(100, Math.round((session.scanned_count / session.total_packages) * 100))
      : 0;

  return (
    <div style={s.panel}>
      <div style={s.header}>
        <span style={s.title}>SCAN STATUS</span>
        {session && (
          <span style={{ ...s.statusBadge, ...statusBadgeStyle(session.status) }}>
            {session.status.toUpperCase()}
          </span>
        )}
      </div>

      <div style={s.body}>
        {!session ? (
          <div style={s.idle}>
            <div style={s.idleIcon}>&#x25C7;</div>
            <div style={s.idleText}>No scan active</div>
          </div>
        ) : (
          <>
            <div style={s.row}>
              <span style={s.label}>DEVICE</span>
              <span style={s.value}>{device?.model ?? session.device_id}</span>
            </div>
            <div style={s.row}>
              <span style={s.label}>PACKAGES</span>
              <span style={s.value}>
                <span style={{ color: "var(--green)" }}>{session.scanned_count}</span>
                <span style={{ color: "var(--text-muted)" }}> / {session.total_packages || "…"}</span>
              </span>
            </div>
            <div style={s.row}>
              <span style={s.label}>THREATS</span>
              <span style={{
                ...s.value,
                color: session.threats_found > 0 ? "var(--red)" : "var(--text-secondary)",
                fontWeight: session.threats_found > 0 ? "bold" : "normal",
              }}>
                {session.threats_found}
              </span>
            </div>

            <div style={s.progressSection}>
              <div style={s.progressHeader}>
                <span style={{ color: "var(--text-secondary)" }}>Progress</span>
                <span style={{ color: "var(--green)" }}>{progress}%</span>
              </div>
              <div style={s.progressBar}>
                <div style={{ ...s.progressFill, width: `${progress}%` }} />
              </div>
            </div>

            {session.status === "scanning" && (
              <div style={s.scanLine}>
                <span style={s.scanDot} className="blink">&#x25A0;</span>
                <span style={{ color: "var(--text-secondary)" }}>Scanning…</span>
              </div>
            )}
            {session.status === "finished" && (
              <div style={s.finishedMsg}>
                &#x2713; Scan complete
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

function statusBadgeStyle(status: string): React.CSSProperties {
  switch (status) {
    case "scanning": return { color: "var(--orange)", borderColor: "var(--orange)" };
    case "finished": return { color: "var(--green)", borderColor: "var(--green-dim)" };
    case "error":    return { color: "var(--red)", borderColor: "var(--red)" };
    default:         return { color: "var(--text-muted)", borderColor: "var(--border)" };
  }
}

const s: Record<string, React.CSSProperties> = {
  panel: {
    borderTop: "1px solid var(--border)",
    display: "flex",
    flexDirection: "column",
    background: "var(--bg-secondary)",
  },
  header: {
    display: "flex", alignItems: "center", justifyContent: "space-between",
    padding: "8px 12px",
    borderBottom: "1px solid var(--border)",
  },
  title: { color: "var(--green)", fontSize: 10, letterSpacing: "0.12em", fontWeight: "bold" },
  statusBadge: {
    fontSize: 9, padding: "1px 6px", border: "1px solid",
    borderRadius: 10, letterSpacing: "0.08em",
  },
  body: { padding: 12, display: "flex", flexDirection: "column", gap: 8 },
  idle: { display: "flex", flexDirection: "column", alignItems: "center", gap: 6, padding: "12px 0", color: "var(--text-muted)" },
  idleIcon: { fontSize: 20 },
  idleText: { fontSize: 11 },
  row: { display: "flex", justifyContent: "space-between", fontSize: 11 },
  label: { color: "var(--text-muted)", letterSpacing: "0.08em" },
  value: { color: "var(--text-secondary)" },
  progressSection: { display: "flex", flexDirection: "column", gap: 4 },
  progressHeader: { display: "flex", justifyContent: "space-between", fontSize: 10 },
  progressBar: { height: 4, background: "var(--bg-tertiary)", borderRadius: 2, overflow: "hidden" },
  progressFill: { height: "100%", background: "linear-gradient(90deg, var(--green-dim), var(--green))", transition: "width 0.3s ease", borderRadius: 2 },
  scanLine: { display: "flex", alignItems: "center", gap: 6, fontSize: 11, color: "var(--orange)" },
  scanDot: { fontSize: 8 },
  finishedMsg: { color: "var(--green)", fontSize: 11, textAlign: "center", padding: "4px 0" },
};
