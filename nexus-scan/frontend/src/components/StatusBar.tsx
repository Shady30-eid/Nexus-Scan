import React from "react";
import type { WsStatus } from "../types";

interface Props {
  wsStatus: WsStatus;
}

const STATUS_CONFIG: Record<WsStatus, { label: string; color: string; pulse: boolean }> = {
  connecting:   { label: "CONNECTING…",   color: "var(--yellow)",        pulse: true  },
  connected:    { label: "CONNECTED",     color: "var(--green)",         pulse: false },
  disconnected: { label: "DISCONNECTED",  color: "var(--text-muted)",    pulse: false },
  error:        { label: "ERROR",         color: "var(--red)",           pulse: false },
};

export function StatusBar({ wsStatus }: Props) {
  const cfg = STATUS_CONFIG[wsStatus];
  const now = new Date().toLocaleString("en-US", {
    hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit",
    year: "numeric", month: "2-digit", day: "2-digit",
  });

  return (
    <div style={s.bar}>
      <div style={s.left}>
        <div style={{ ...s.dot, background: cfg.color, boxShadow: `0 0 5px ${cfg.color}`, animation: cfg.pulse ? "pulse-green 1.4s infinite" : undefined }} />
        <span style={{ color: cfg.color, fontSize: 10, letterSpacing: "0.1em" }}>
          WS {cfg.label}
        </span>
        <span style={s.sep}>|</span>
        <span style={s.addr}>ws://127.0.0.1:9999/ws</span>
      </div>
      <div style={s.right}>
        <span style={s.kali}>KALI LINUX</span>
        <span style={s.sep}>|</span>
        <span style={s.time}>{now}</span>
      </div>
    </div>
  );
}

const s: Record<string, React.CSSProperties> = {
  bar: {
    display: "flex", alignItems: "center", justifyContent: "space-between",
    padding: "3px 14px",
    background: "var(--bg-secondary)",
    borderBottom: "1px solid var(--border)",
    flexShrink: 0,
    fontSize: 10,
  },
  left: { display: "flex", alignItems: "center", gap: 8 },
  dot: { width: 6, height: 6, borderRadius: "50%", flexShrink: 0 },
  sep: { color: "var(--border-bright)" },
  addr: { color: "var(--text-muted)", fontFamily: "var(--font-mono)" },
  right: { display: "flex", alignItems: "center", gap: 8 },
  kali: { color: "var(--green-dim)", letterSpacing: "0.1em" },
  time: { color: "var(--text-muted)", fontFamily: "var(--font-mono)" },
};
