import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { DNSRecord, DNSCheckResult } from "@/types/api";

const { mockGetDnsRecords, mockGetDnsCheck } = vi.hoisted(() => ({
  mockGetDnsRecords: vi.fn(),
  mockGetDnsCheck: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getDnsRecords: mockGetDnsRecords,
    getDnsCheck: mockGetDnsCheck,
  },
}));

import { useDnsRecords, useDnsCheck } from "@/hooks/use-dns";

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

const mockRecords: DNSRecord[] = [
  {
    domain: "example.com",
    type: "MX",
    name: "example.com",
    value: "mail.example.com",
    priority: 10,
  },
];

const mockCheckResults: DNSCheckResult[] = [
  {
    domain: "example.com",
    record: "MX",
    expected: "mail.example.com",
    actual: "mail.example.com",
    valid: true,
  },
];

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useDnsRecords", () => {
  it("fetches DNS records successfully", async () => {
    mockGetDnsRecords.mockResolvedValue({ records: mockRecords });

    const { result } = renderHook(() => useDnsRecords(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.records).toEqual(mockRecords);
  });

  it("returns empty records when none exist", async () => {
    mockGetDnsRecords.mockResolvedValue({ records: [] });

    const { result } = renderHook(() => useDnsRecords(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.records).toHaveLength(0);
  });

  it("handles fetch error", async () => {
    mockGetDnsRecords.mockRejectedValue(new Error("DNS lookup failed"));

    const { result } = renderHook(() => useDnsRecords(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useDnsCheck", () => {
  it("starts disabled (does not fetch automatically)", () => {
    const { result } = renderHook(() => useDnsCheck(), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
    expect(mockGetDnsCheck).not.toHaveBeenCalled();
  });

  it("fetches when manually triggered", async () => {
    mockGetDnsCheck.mockResolvedValue({ results: mockCheckResults });

    const { result } = renderHook(() => useDnsCheck(), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");

    result.current.refetch();

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.results).toEqual(mockCheckResults);
  });

  it("fetches automatically when enabled", async () => {
    mockGetDnsCheck.mockResolvedValue({ results: mockCheckResults });

    const { result } = renderHook(() => useDnsCheck({ enabled: true }), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockGetDnsCheck).toHaveBeenCalledTimes(1);
    expect(result.current.data?.results).toEqual(mockCheckResults);
  });

  it("handles DNS check error", async () => {
    mockGetDnsCheck.mockRejectedValue(new Error("Check failed"));

    const { result } = renderHook(() => useDnsCheck(), {
      wrapper: createWrapper(),
    });

    result.current.refetch();

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
