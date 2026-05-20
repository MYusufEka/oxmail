"use client";

import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState, useCallback } from "react";
import { apiClient, type PaginationParams } from "@/lib/api-client";
import type { LogEntry } from "@/types/api";

const WS_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL?.replace(/^http/, "ws") ?? "ws://localhost:8080";

export function useLogs(params?: PaginationParams) {
  return useQuery({
    queryKey: ["logs", params],
    queryFn: () => apiClient.getLogs(params),
  });
}

interface UseLogStreamOptions {
  enabled?: boolean;
  onMessage?: (entry: LogEntry) => void;
}

interface LogStreamState {
  entries: LogEntry[];
  connected: boolean;
}

export function useLogStream(options: UseLogStreamOptions = {}) {
  const { enabled = true, onMessage } = options;
  const [state, setState] = useState<LogStreamState>({
    entries: [],
    connected: false,
  });

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(`${WS_BASE_URL}/api/logs/stream`);
    wsRef.current = ws;

    ws.addEventListener("open", () => {
      setState((prev) => ({ ...prev, connected: true }));
    });

    ws.addEventListener("message", (event) => {
      const entry = JSON.parse(event.data as string) as LogEntry;
      setState((prev) => ({
        ...prev,
        entries: [...prev.entries, entry],
      }));
      onMessageRef.current?.(entry);
    });

    ws.addEventListener("close", () => {
      setState((prev) => ({ ...prev, connected: false }));
      wsRef.current = null;

      if (enabled) {
        reconnectTimeoutRef.current = setTimeout(connect, 3000);
      }
    });

    ws.addEventListener("error", () => {
      ws.close();
    });
  }, [enabled]);

  useEffect(() => {
    if (!enabled) {
      wsRef.current?.close();
      return;
    }

    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      wsRef.current?.close();
    };
  }, [enabled, connect]);

  const clearEntries = useCallback(() => {
    setState((prev) => ({ ...prev, entries: [] }));
  }, []);

  return {
    entries: state.entries,
    connected: state.connected,
    clearEntries,
  };
}
