import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import type { LogsResponse, LogEntry } from "@/types/api";

const { mockGetLogs, mockWsInstance, MockWebSocketCtor } = vi.hoisted(() => {
  const mockWsInstance = {
    addEventListener: vi.fn(),
    close: vi.fn(),
    readyState: 0,
  };
  const MockWebSocketCtor = vi.fn(function MockWebSocket() {
    return mockWsInstance;
  });
  Object.assign(MockWebSocketCtor, { OPEN: 1 });
  return { mockGetLogs: vi.fn(), mockWsInstance, MockWebSocketCtor };
});

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getLogs: mockGetLogs,
  },
}));

import { useLogs, useLogStream } from "@/hooks/use-logs";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

const mockLogResponse: LogsResponse = {
  entries: [
    {
      id: 1,
      timestamp: "2024-01-01T00:00:00Z",
      service: "postfix",
      level: "info",
      message: "Mail delivered",
    },
  ],
  total: 1,
  limit: 20,
  offset: 0,
};

const emptyLogResponse: LogsResponse = {
  entries: [],
  total: 0,
  limit: 20,
  offset: 0,
};

describe("useLogs", () => {
  beforeEach(() => {
    mockGetLogs.mockReset();
  });

  it("fetches logs successfully", async () => {
    mockGetLogs.mockResolvedValue(mockLogResponse);

    const { result } = renderHook(() => useLogs(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockLogResponse);
  });

  it("returns empty logs when none exist", async () => {
    mockGetLogs.mockResolvedValue(emptyLogResponse);

    const { result } = renderHook(() => useLogs(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.entries).toHaveLength(0);
  });

  it("handles fetch error", async () => {
    mockGetLogs.mockRejectedValue(new Error("Failed to fetch logs"));

    const { result } = renderHook(() => useLogs(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useLogStream", () => {
  let originalWs: unknown;

  beforeEach(() => {
    MockWebSocketCtor.mockClear();
    mockWsInstance.addEventListener.mockReset();
    mockWsInstance.close.mockReset();
    mockWsInstance.readyState = 0;
    originalWs = globalThis.WebSocket;
    globalThis.WebSocket = MockWebSocketCtor as unknown as typeof WebSocket;
    window.WebSocket = MockWebSocketCtor as unknown as typeof WebSocket;
  });

  afterEach(() => {
    globalThis.WebSocket = originalWs as typeof WebSocket;
    window.WebSocket = originalWs as typeof WebSocket;
  });

  it("starts with empty entries and not connected", () => {
    const { result } = renderHook(() => useLogStream(), {
      wrapper: createWrapper(),
    });

    expect(result.current.entries).toHaveLength(0);
    expect(result.current.connected).toBe(false);
  });

  it("updates connected state on WebSocket open", async () => {
    const { result } = renderHook(() => useLogStream(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(mockWsInstance.addEventListener).toHaveBeenCalled());
    const openHandler = mockWsInstance.addEventListener.mock.calls.find(
      ([event]: [string]) => event === "open",
    );
    expect(openHandler).toBeDefined();
    act(() => {
      openHandler?.[1]();
    });

    expect(result.current.connected).toBe(true);
  });

  it("adds entry on WebSocket message", async () => {
    const { result } = renderHook(() => useLogStream(), {
      wrapper: createWrapper(),
    });

    const logEntry: LogEntry = {
      id: 1,
      timestamp: "2024-01-01T00:00:00Z",
      service: "postfix",
      level: "info",
      message: "Test log entry",
    };

    await waitFor(() => expect(mockWsInstance.addEventListener).toHaveBeenCalled());
    const messageHandler = mockWsInstance.addEventListener.mock.calls.find(
      ([event]: [string]) => event === "message",
    );
    expect(messageHandler).toBeDefined();
    act(() => {
      messageHandler?.[1]({ data: JSON.stringify(logEntry) });
    });

    expect(result.current.entries).toHaveLength(1);
    expect(result.current.entries[0]).toEqual(logEntry);
  });

  it("calls onMessage callback when provided", async () => {
    const onMessageCallback = vi.fn();
    const { result } = renderHook(() => useLogStream({ onMessage: onMessageCallback }), {
      wrapper: createWrapper(),
    });

    const logEntry: LogEntry = {
      id: 1,
      timestamp: "2024-01-01T00:00:00Z",
      service: "postfix",
      level: "info",
      message: "Test log entry",
    };

    await waitFor(() => expect(mockWsInstance.addEventListener).toHaveBeenCalled());
    const messageHandler = mockWsInstance.addEventListener.mock.calls.find(
      ([event]: [string]) => event === "message",
    );
    expect(messageHandler).toBeDefined();
    act(() => {
      messageHandler?.[1]({ data: JSON.stringify(logEntry) });
    });

    expect(onMessageCallback).toHaveBeenCalledWith(logEntry);
  });

  it("clears entries when clearEntries is called", async () => {
    const { result } = renderHook(() => useLogStream(), {
      wrapper: createWrapper(),
    });

    const logEntry: LogEntry = {
      id: 1,
      timestamp: "2024-01-01T00:00:00Z",
      service: "postfix",
      level: "info",
      message: "Test",
    };

    await waitFor(() => expect(mockWsInstance.addEventListener).toHaveBeenCalled());
    const messageHandler = mockWsInstance.addEventListener.mock.calls.find(
      ([event]: [string]) => event === "message",
    );
    expect(messageHandler).toBeDefined();
    act(() => {
      messageHandler?.[1]({ data: JSON.stringify(logEntry) });
    });

    expect(result.current.entries).toHaveLength(1);

    act(() => {
      result.current.clearEntries();
    });

    expect(result.current.entries).toHaveLength(0);
  });

  it("does not connect when disabled", () => {
    renderHook(() => useLogStream({ enabled: false }), {
      wrapper: createWrapper(),
    });

    expect(MockWebSocketCtor).not.toHaveBeenCalled();
  });
});
