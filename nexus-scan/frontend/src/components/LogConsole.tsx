import React, { useEffect, useRef, useState } from "react";
import type { LogEntry } from "../types";

interface Props {
  logs: LogEntry[];
}

const LEVEL_COLORS: Record<LogEntry["level"], string> = {
  INFO:    "var(--text-secondary)",
  WARNING: "var(--yellow)",
  ERROR:   "var(--red)",
  THREAT:  "var(--red)",
};

const LEVEL_PREFIX: Record<LogEntry["level"], string> = {
  INFO:    "[INFO]   ",
  WARNING: "[WARN]   ",
  ERROR:   "[ERROR]  ",
  THREAT:  "[THREAT] ",
};

export function LogConsole({ logs }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [filter, setFilter] = useState<"ALL" | "THREAT" | "ERROR" | "WARNING">("ALL");

  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs, autoScroll]);

  function handleScroll(e: React.UIEvent<HTMLDivElement>) {
    const el = e.currentTarget;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
    setAutoScroll(atBottom);
  }

  const filtered = logs.filter((l) => filter === "ALL" || l.level === filter);

  return (
    <div style={s.container}>
      <div style={s.header}>
        <span style={s.title}>LIVE LOG CONSOLE</span>
        <div style={s.controls}>
          {(["ALL", "THREAT", "ERROR", "WARNING"] as const).map((f) => (
            <button
              key={f}
              style={{
                ...s.filterBtn,
                ...(filter === f ? s.filterActive : {}),
                color: f === "THREAT" || f === "ERROR" ? "var(--red)" : f === "WARNING" ? "var(--yellow)" : "var(--text-secondary)",
              }}
              onClick={() => setFilter(f)}
            >
              {f}
            </button>
          ))}
          <span style={s.scrollIndicator}>
            {autoScroll ? (
              <span style={{ color: "var(--green-dim)" }}>&#x25BC; AUTO</span>
            ) : (
              <button style={s.scrollBtn} onClick={() => { setAutoScroll(true); bottomRef.current?.scrollIntoView(); }}>
                &#x25BC; SCROLL
              </button>
            )}
          </span>
        </div>
      </div>

      <div style={s.body} onScroll={handleScroll}>
        {filtered.length === 0 ? (
          <div style={s.empty}>
            <span style={{ color: "var(--text-muted)" }}>Awaiting events</span>
            <span style={s.cursor} className="blink">_</span>
          </div>
        ) : (
          filtered.map((entry, i) => (
            <LogLine key={i} entry={entry} />
          ))
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}

function LogLine({ entry }: { entry: LogEntry }) {
  const time = new Date(entry.timestamp).toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });

  return (
    <div style={{
      ...s.line,
      background: entry.level === "THREAT" ? "rgba(255,45,45,0.04)" : undefined,
    }}>
      <span style={s.time}>{time}</span>
      <span style={{ ...s.level, color: LEVEL_COLORS[entry.level] }}>
        {LEVEL_PREFIX[entry.level]}
      </span>
      {entry.device_id && (
        <span style={s.deviceTag}>[{entry.device_id.slice(0, 12)}] </span>
      )}
      <span style={{ color: LEVEL_COLORS[entry.level], opacity: entry.level === "INFO" ? 0.85 : 1 }}>
        {entry.message}
      </span>
    </div>
  );
}

const s: Record<string, React.CSSProperties> = {
  container: {
    height: 180,
    flexShrink: 0,
    display: "flex",
    flexDirection: "column",
    background: "var(--bg-primary)",
    borderTop: "1px solid var(--border)",
  },
  header: {
    display: "flex", alignItems: "center", justifyContent: "space-between",
    padding: "5px 12px",
    background: "var(--bg-secondary)",
    borderBottom: "1px solid var(--border)",
    flexShrink: 0,
  },
  title: { color: "var(--green)", fontSize: 10, letterSpacing: "0.12em", fontWeight: "bold" },
  controls: { display: "flex", alignItems: "center", gap: 6 },
  filterBtn: {
    background: "transparent", border: "1px solid var(--border)",
    borderRadius: "var(--radius-sm)", padding: "1px 6px", fontSize: 9,
    letterSpacing: "0.07em", cursor: "pointer", color: "var(--text-muted)",
  },
  filterActive: { background: "var(--bg-tertiary)", borderColor: "var(--border-bright)" },
  scrollIndicator: { fontSize: 9, color: "var(--text-muted)", marginLeft: 4 },
  scrollBtn: {
    background: "transparent", border: "none", color: "var(--text-muted)",
    fontSize: 9, cursor: "pointer", padding: "1px 4px",
  },
  body: {
    flex: 1, overflowY: "auto", padding: "4px 0",
    fontFamily: "var(--font-mono)", fontSize: 11,
  },
  line: {
    display: "flex", padding: "1px 12px", gap: 4,
    lineHeight: 1.7, whiteSpace: "nowrap", overflow: "hidden", textOverflow: "ellipsis",
    alignItems: "baseline",
  },
  time: { color: "var(--text-muted)", marginRight: 4, flexShrink: 0, fontSize: 10 },
  level: { flexShrink: 0, fontWeight: "bold" },
  deviceTag: { color: "var(--blue-dim)", flexShrink: 0 },
  empty: { display: "flex", gap: 4, padding: "8px 12px", color: "var(--text-muted)", alignItems: "center" },
  cursor: { color: "var(--green)" },
};
