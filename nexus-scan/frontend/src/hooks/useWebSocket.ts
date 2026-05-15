import { useEffect, useRef, useState, useCallback } from "react";
import type { WsEvent, WsStatus } from "../types";

const WS_URL = "ws://127.0.0.1:9999/ws";
const RECONNECT_DELAY_MS = 3000;
const MAX_RECONNECT_ATTEMPTS = 20;

type EventHandler = (event: WsEvent) => void;

export function useWebSocket() {
  const [status, setStatus] = useState<WsStatus>("connecting");
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const handlersRef = useRef<Map<string, Set<EventHandler>>>(new Map());
  const unmountedRef = useRef(false);

  const send = useCallback((type: string, payload: Record<string, unknown> = {}) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }));
    }
  }, []);

  const on = useCallback((eventType: string, handler: EventHandler) => {
    if (!handlersRef.current.has(eventType)) {
      handlersRef.current.set(eventType, new Set());
    }
    handlersRef.current.get(eventType)!.add(handler);
    return () => {
      handlersRef.current.get(eventType)?.delete(handler);
    };
  }, []);

  const connect = useCallback(() => {
    if (unmountedRef.current) return;

    setStatus("connecting");
    const ws = new WebSocket(WS_URL);
    wsRef.current = ws;

    ws.onopen = () => {
      if (unmountedRef.current) { ws.close(); return; }
      reconnectAttempts.current = 0;
      setStatus("connected");
    };

    ws.onmessage = (event) => {
      if (unmountedRef.current) return;
      try {
        const parsed: WsEvent = JSON.parse(event.data as string);
        const handlers = handlersRef.current.get(parsed.type);
        if (handlers) {
          handlers.forEach((h) => h(parsed));
        }
        const wildcards = handlersRef.current.get("*");
        if (wildcards) {
          wildcards.forEach((h) => h(parsed));
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onclose = () => {
      if (unmountedRef.current) return;
      setStatus("disconnected");
      if (reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts.current++;
        const delay = Math.min(RECONNECT_DELAY_MS * reconnectAttempts.current, 30000);
        reconnectTimer.current = setTimeout(connect, delay);
      } else {
        setStatus("error");
      }
    };

    ws.onerror = () => {
      if (unmountedRef.current) return;
      setStatus("error");
    };
  }, []);

  useEffect(() => {
    unmountedRef.current = false;
    connect();
    return () => {
      unmountedRef.current = true;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [connect]);

  return { status, send, on };
}
