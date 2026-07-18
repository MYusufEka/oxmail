import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { DailyStat, StatsSummary } from "@/types/api";

const { mockGetStats, mockGetStatsSummary } = vi.hoisted(() => ({
  mockGetStats: vi.fn(),
  mockGetStatsSummary: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getStats: mockGetStats,
    getStatsSummary: mockGetStatsSummary,
  },
}));

import { useStats, useStatsSummary } from "@/hooks/use-stats";

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

const mockDailyStats: DailyStat[] = [
  { date: "2024-01-01", sent: 10, received: 15, bounced: 1, spamCaught: 3 },
  { date: "2024-01-02", sent: 12, received: 18, bounced: 0, spamCaught: 5 },
];

const mockSummary: StatsSummary = {
  totalSent: 22,
  totalReceived: 33,
  totalBounced: 1,
  totalSpamCaught: 8,
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useStats", () => {
  it("fetches stats successfully with default days", async () => {
    mockGetStats.mockResolvedValue(mockDailyStats);

    const { result } = renderHook(() => useStats(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDailyStats);
    expect(mockGetStats).toHaveBeenCalledWith(7);
  });

  it("fetches stats with custom days parameter", async () => {
    mockGetStats.mockResolvedValue(mockDailyStats);

    const { result } = renderHook(() => useStats(30), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGetStats).toHaveBeenCalledWith(30);
  });

  it("returns empty array when no stats exist", async () => {
    mockGetStats.mockResolvedValue([]);

    const { result } = renderHook(() => useStats(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(0);
  });

  it("handles fetch error", async () => {
    mockGetStats.mockRejectedValue(new Error("Failed to fetch stats"));

    const { result } = renderHook(() => useStats(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useStatsSummary", () => {
  it("fetches stats summary successfully", async () => {
    mockGetStatsSummary.mockResolvedValue(mockSummary);

    const { result } = renderHook(() => useStatsSummary(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockSummary);
  });

  it("handles summary fetch error", async () => {
    mockGetStatsSummary.mockRejectedValue(new Error("Failed to fetch summary"));

    const { result } = renderHook(() => useStatsSummary(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
