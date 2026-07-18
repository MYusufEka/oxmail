import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { HealthStatus } from "@/types/api";

const { mockGetHealth } = vi.hoisted(() => ({
  mockGetHealth: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getHealth: mockGetHealth,
  },
}));

import { useHealth } from "@/hooks/use-health";

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

const mockHealth: HealthStatus = {
  status: "healthy",
  version: "1.0.0",
  uptime: "1d 0h 0m",
  services: [
    { name: "postfix", status: "healthy", latencyMs: 5 },
    { name: "dovecot", status: "healthy", latencyMs: 3 },
    { name: "rspamd", status: "healthy", latencyMs: 10 },
    { name: "redis", status: "healthy", latencyMs: 1 },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useHealth", () => {
  it("fetches health status successfully", async () => {
    mockGetHealth.mockResolvedValue(mockHealth);

    const { result } = renderHook(() => useHealth(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockHealth);
  });

  it("handles health check error", async () => {
    mockGetHealth.mockRejectedValue(new Error("Health check failed"));

    const { result } = renderHook(() => useHealth(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });

  it("returns degraded status when a service is down", async () => {
    const degradedHealth: HealthStatus = {
      status: "degraded",
      version: "1.0.0",
      uptime: "1d 0h 0m",
      services: [
        { name: "postfix", status: "unhealthy", latencyMs: 0 },
        { name: "dovecot", status: "healthy", latencyMs: 3 },
      ],
    };
    mockGetHealth.mockResolvedValue(degradedHealth);

    const { result } = renderHook(() => useHealth(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBe("degraded");
  });
});
