import React, { useEffect, useReducer, useCallback, useRef } from "react";
import { useWebSocket } from "./hooks/useWebSocket";
import { DevicePanel } from "./components/DevicePanel";
import { ScanStatusPanel } from "./components/ScanStatusPanel";
import { ThreatTable } from "./components/ThreatTable";
import { LogConsole } from "./components/LogConsole";
import { StatusBar } from "./components/StatusBar";
import type {
  Device,
  ScanResult,
  LogEntry,
  ScanSession,
  WsEvent,
} from "./types";

interface AppState {
  devices: Device[];
  selectedDevice: string | null;
  threats: ScanResult[];
  logs: LogEntry[];
  scanSession: ScanSession | null;
}

type Action =
  | { type: "DEVICE_DETECTED"; device: Device }
  | { type: "DEVICE_DISCONNECTED"; device_id: string }
  | { type: "DEVICE_LIST"; devices: Device[] }
  | { type: "SELECT_DEVICE"; device_id: string }
  | { type: "SCAN_STARTED"; device_id: string }
  | { type: "SCAN_PROGRESS"; device_id: string; total_packages: number; scanned: number }
  | { type: "PACKAGE_SCANNED"; device_id: string; scanned: number; total: number }
  | { type: "THREAT_DETECTED"; result: ScanResult }
  | { type: "SCAN_FINISHED"; device_id: string; total_scanned: number; threats_found: number }
  | { type: "SCAN_HISTORY"; history: ScanResult[] }
  | { type: "ADD_LOG"; entry: LogEntry }
  | { type: "REMEDIATION_COMPLETED"; device_id: string; package_name: string; success: boolean }
  | { type: "CLEAR_THREATS" };

function reducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case "DEVICE_DETECTED": {
      const exists = state.devices.find((d) => d.id === action.device.id);
      if (exists) {
        return {
          ...state,
          devices: state.devices.map((d) =>
            d.id === action.device.id ? action.device : d
          ),
        };
      }
      return { ...state, devices: [...state.devices, action.device] };
    }
    case "DEVICE_DISCONNECTED":
      return {
        ...state,
        devices: state.devices.filter((d) => d.id !== action.device_id),
        selectedDevice:
          state.selectedDevice === action.device_id ? null : state.selectedDevice,
      };
    case "DEVICE_LIST":
      return { ...state, devices: action.devices };
    case "SELECT_DEVICE":
      return { ...state, selectedDevice: action.device_id };
    case "SCAN_STARTED":
      return {
        ...state,
        scanSession: {
          device_id: action.device_id,
          total_packages: 0,
          scanned_count: 0,
          threats_found: 0,
          status: "scanning",
        },
      };
    case "SCAN_PROGRESS":
      return {
        ...state,
        scanSession: state.scanSession
          ? {
              ...state.scanSession,
              total_packages: action.total_packages,
              scanned_count: action.scanned,
            }
          : null,
      };
    case "PACKAGE_SCANNED":
      return {
        ...state,
        scanSession: state.scanSession
          ? {
              ...state.scanSession,
              scanned_count: action.scanned,
              total_packages: action.total,
            }
          : null,
      };
    case "THREAT_DETECTED":
      return {
        ...state,
        threats: [action.result, ...state.threats].slice(0, 500),
        scanSession: state.scanSession
          ? {
              ...state.scanSession,
              threats_found: state.scanSession.threats_found + 1,
            }
          : null,
      };
    case "SCAN_FINISHED":
      return {
        ...state,
        scanSession: state.scanSession
          ? {
              ...state.scanSession,
              status: "finished",
              total_packages: action.total_scanned,
              scanned_count: action.total_scanned,
              threats_found: action.threats_found,
            }
          : null,
      };
    case "SCAN_HISTORY":
      return { ...state, threats: action.history };
    case "ADD_LOG":
      return {
        ...state,
        logs: [...state.logs, action.entry].slice(-500),
      };
    case "REMEDIATION_COMPLETED": {
      if (action.success) {
        return {
          ...state,
          threats: state.threats.filter(
            (t) =>
              !(t.device_id === action.device_id && t.package_name === action.package_name)
          ),
        };
      }
      return state;
    }
    case "CLEAR_THREATS":
      return { ...state, threats: [] };
    default:
      return state;
  }
}

const initialState: AppState = {
  devices: [],
  selectedDevice: null,
  threats: [],
  logs: [],
  scanSession: null,
};

export default function App() {
  const [state, dispatch] = useReducer(reducer, initialState);
  const { status: wsStatus, send, on } = useWebSocket();
  const didRequestDevices = useRef(false);

  const addLog = useCallback(
    (level: LogEntry["level"], message: string, extra?: Partial<LogEntry>) => {
      dispatch({
        type: "ADD_LOG",
        entry: {
          timestamp: new Date().toISOString(),
          level,
          message,
          ...extra,
        },
      });
    },
    []
  );

  useEffect(() => {
    const unsub: Array<() => void> = [];

    unsub.push(
      on("connected", () => {
        addLog("INFO", "Connected to Nexus-Scan backend");
        if (!didRequestDevices.current) {
          send("list_devices");
          didRequestDevices.current = true;
        }
      })
    );

    unsub.push(
      on("device_detected", (e: WsEvent) => {
        const device = e.payload["device"] as Device;
        dispatch({ type: "DEVICE_DETECTED", device });
        addLog("INFO", `Device detected: ${device.model} (${device.serial_number})`, {
          device_id: device.id,
        });
      })
    );

    unsub.push(
      on("device_disconnected", (e: WsEvent) => {
        const deviceId = e.payload["device_id"] as string;
        dispatch({ type: "DEVICE_DISCONNECTED", device_id: deviceId });
        addLog("WARNING", `Device disconnected: ${deviceId}`, { device_id: deviceId });
      })
    );

    unsub.push(
      on("device_list", (e: WsEvent) => {
        const devices = (e.payload["devices"] as Device[]) || [];
        dispatch({ type: "DEVICE_LIST", devices });
        addLog("INFO", `${devices.length} device(s) found`);
      })
    );

    unsub.push(
      on("scan_started", (e: WsEvent) => {
        const deviceId = e.payload["device_id"] as string;
        dispatch({ type: "SCAN_STARTED", device_id: deviceId });
        addLog("INFO", "Scan started", { device_id: deviceId });
      })
    );

    unsub.push(
      on("scan_progress", (e: WsEvent) => {
        dispatch({
          type: "SCAN_PROGRESS",
          device_id: e.payload["device_id"] as string,
          total_packages: e.payload["total_packages"] as number,
          scanned: e.payload["scanned"] as number,
        });
      })
    );

    unsub.push(
      on("package_scanned", (e: WsEvent) => {
        dispatch({
          type: "PACKAGE_SCANNED",
          device_id: e.payload["device_id"] as string,
          scanned: e.payload["scanned"] as number,
          total: e.payload["total"] as number,
        });
      })
    );

    unsub.push(
      on("threat_detected", (e: WsEvent) => {
        const result: ScanResult = {
          id: crypto.randomUUID(),
          device_id: e.payload["device_id"] as string,
          package_name: e.payload["package_name"] as string,
          file_path: e.payload["file_path"] as string,
          sha256_hash: e.payload["sha256_hash"] as string,
          threat_name: e.payload["threat_name"] as string,
          severity: e.payload["severity"] as ScanResult["severity"],
          infection_vector: e.payload["infection_vector"] as string,
          scan_timestamp: new Date().toISOString(),
          malware_family: e.payload["malware_family"] as string | undefined,
          description: e.payload["description"] as string | undefined,
        };
        dispatch({ type: "THREAT_DETECTED", result });
        addLog("THREAT", `THREAT: ${result.threat_name} — ${result.package_name}`, {
          device_id: result.device_id,
          package_name: result.package_name,
        });
      })
    );

    unsub.push(
      on("scan_finished", (e: WsEvent) => {
        dispatch({
          type: "SCAN_FINISHED",
          device_id: e.payload["device_id"] as string,
          total_scanned: e.payload["total_scanned"] as number,
          threats_found: e.payload["threats_found"] as number,
        });
        addLog(
          "INFO",
          `Scan complete — ${e.payload["total_scanned"]} packages, ${e.payload["threats_found"]} threats`,
          { device_id: e.payload["device_id"] as string }
        );
      })
    );

    unsub.push(
      on("scan_history", (e: WsEvent) => {
        const history = (e.payload["history"] as ScanResult[]) || [];
        dispatch({ type: "SCAN_HISTORY", history });
        addLog("INFO", `Loaded ${history.length} scan history entries`);
      })
    );

    unsub.push(
      on("remediation_completed", (e: WsEvent) => {
        const deviceId = e.payload["device_id"] as string;
        const packageName = e.payload["package_name"] as string;
        const success = e.payload["success"] as boolean;
        dispatch({ type: "REMEDIATION_COMPLETED", device_id: deviceId, package_name: packageName, success });
        addLog(
          success ? "INFO" : "ERROR",
          success
            ? `Purged: ${packageName}`
            : `Purge failed: ${packageName} — ${e.payload["message"]}`,
          { device_id: deviceId, package_name: packageName }
        );
      })
    );

    unsub.push(
      on("remediation_failed", (e: WsEvent) => {
        addLog("ERROR", `Remediation failed: ${e.payload["message"]}`, {
          device_id: e.payload["device_id"] as string,
          package_name: e.payload["package_name"] as string,
        });
      })
    );

    unsub.push(
      on("error", (e: WsEvent) => {
        addLog("ERROR", `Error [${e.payload["code"]}]: ${e.payload["message"]}`);
      })
    );

    return () => unsub.forEach((u) => u());
  }, [on, send, addLog]);

  useEffect(() => {
    if (wsStatus === "disconnected" || wsStatus === "error") {
      didRequestDevices.current = false;
      addLog(
        wsStatus === "error" ? "ERROR" : "WARNING",
        wsStatus === "error"
          ? "WebSocket error — check backend"
          : "WebSocket disconnected — reconnecting..."
      );
    }
  }, [wsStatus, addLog]);

  const handleStartScan = useCallback(
    (deviceId: string) => {
      dispatch({ type: "CLEAR_THREATS" });
      send("start_scan", { device_id: deviceId });
      addLog("INFO", "Scan initiated", { device_id: deviceId });
    },
    [send, addLog]
  );

  const handleStopScan = useCallback(
    (deviceId: string) => {
      send("stop_scan", { device_id: deviceId });
      addLog("WARNING", "Scan stopped by user", { device_id: deviceId });
    },
    [send, addLog]
  );

  const handleRemediate = useCallback(
    (deviceId: string, packageName: string) => {
      send("remediate", { device_id: deviceId, package_name: packageName });
      addLog("WARNING", `Purge initiated: ${packageName}`, {
        device_id: deviceId,
        package_name: packageName,
      });
    },
    [send, addLog]
  );

  const handleSelectDevice = useCallback((deviceId: string) => {
    dispatch({ type: "SELECT_DEVICE", device_id: deviceId });
    send("get_scan_history", { device_id: deviceId });
  }, [send]);

  const selectedDevice = state.devices.find((d) => d.id === state.selectedDevice) ?? null;

  return (
    <div style={styles.root}>
      <Header />
      <StatusBar wsStatus={wsStatus} />
      <div style={styles.layout}>
        <aside style={styles.sidebar}>
          <DevicePanel
            devices={state.devices}
            selectedId={state.selectedDevice}
            onSelect={handleSelectDevice}
            onStartScan={handleStartScan}
            onStopScan={handleStopScan}
            scanSession={state.scanSession}
          />
          <ScanStatusPanel session={state.scanSession} device={selectedDevice} />
        </aside>
        <main style={styles.main}>
          <ThreatTable
            threats={state.threats}
            selectedDeviceId={state.selectedDevice}
            onRemediate={handleRemediate}
          />
          <LogConsole logs={state.logs} />
        </main>
      </div>
    </div>
  );
}

function Header() {
  return (
    <header style={styles.header}>
      <div style={styles.headerLeft}>
        <span style={styles.logo}>&#x25B6; NEXUS-SCAN</span>
        <span style={styles.logoSub}>Mobile Forensic & Malware Remediation Platform</span>
      </div>
      <div style={styles.headerRight}>
        <span style={styles.headerVersion}>v1.0.0</span>
      </div>
    </header>
  );
}

const styles: Record<string, React.CSSProperties> = {
  root: {
    display: "flex",
    flexDirection: "column",
    height: "100vh",
    background: "var(--bg-primary)",
    overflow: "hidden",
  },
  header: {
    display: "flex",
    alignItems: "center",
    justifyContent: "space-between",
    padding: "8px 16px",
    background: "var(--bg-secondary)",
    borderBottom: "1px solid var(--green-dim)",
    flexShrink: 0,
  },
  headerLeft: {
    display: "flex",
    alignItems: "center",
    gap: 12,
  },
  logo: {
    color: "var(--green)",
    fontSize: 16,
    fontWeight: "bold",
    letterSpacing: "0.15em",
    textShadow: "0 0 8px var(--green-glow)",
  },
  logoSub: {
    color: "var(--text-secondary)",
    fontSize: 11,
  },
  headerRight: {
    color: "var(--text-muted)",
    fontSize: 11,
  },
  headerVersion: {},
  layout: {
    display: "flex",
    flex: 1,
    overflow: "hidden",
  },
  sidebar: {
    width: 280,
    flexShrink: 0,
    borderRight: "1px solid var(--border)",
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
  },
  main: {
    flex: 1,
    display: "flex",
    flexDirection: "column",
    overflow: "hidden",
    minWidth: 0,
  },
};
