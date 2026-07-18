import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { DKIMKey } from "@/types/api";

const { mockGetDkim, mockGenerateDkim, ApiErrorClass } = vi.hoisted(() => {
  class ApiError extends Error {
    status: number;
    code: string;
    constructor(status: number, code: string, message: string) {
      super(message);
      this.name = "ApiError";
      this.status = status;
      this.code = code;
    }
  }
  return {
    mockGetDkim: vi.fn(),
    mockGenerateDkim: vi.fn(),
    ApiErrorClass: ApiError,
  };
});

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getDkim: mockGetDkim,
    generateDkim: mockGenerateDkim,
  },
  ApiError: ApiErrorClass,
}));

import { useDkim, useGenerateDkim } from "@/hooks/use-dkim";

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

const mockDkim: DKIMKey = {
  domain: "example.com",
  selector: "default",
  publicKey: "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
  dnsRecord: "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
  createdAt: "2024-01-01T00:00:00Z",
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useDkim", () => {
  it("fetches DKIM key successfully", async () => {
    mockGetDkim.mockResolvedValue(mockDkim);

    const { result } = renderHook(() => useDkim("example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDkim);
  });

  it("does not fetch when domain is empty", () => {
    const { result } = renderHook(() => useDkim(""), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("handles fetch error", async () => {
    mockGetDkim.mockRejectedValue(new ApiErrorClass(404, "NOT_FOUND", "DKIM key not found"));

    const { result } = renderHook(() => useDkim("example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useGenerateDkim", () => {
  it("generates DKIM key successfully", async () => {
    mockGenerateDkim.mockResolvedValue(mockDkim);

    const { result } = renderHook(() => useGenerateDkim(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("example.com");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDkim);
  });

  it("handles generate error", async () => {
    mockGenerateDkim.mockRejectedValue(new Error("Generation failed"));

    const { result } = renderHook(() => useGenerateDkim(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("example.com");

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
