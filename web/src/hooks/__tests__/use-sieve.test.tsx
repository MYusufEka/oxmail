import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { SieveResponse } from "@/types/api";

const { mockGetSieveScript, mockSetSieveScript, mockDeleteSieveScript } = vi.hoisted(() => ({
  mockGetSieveScript: vi.fn(),
  mockSetSieveScript: vi.fn(),
  mockDeleteSieveScript: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getSieveScript: mockGetSieveScript,
    setSieveScript: mockSetSieveScript,
    deleteSieveScript: mockDeleteSieveScript,
  },
}));

import {
  useSieveScript,
  useSetSieveScript,
  useDeleteSieveScript,
} from "@/hooks/use-sieve";

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

const mockSieve: SieveResponse = {
  email: "alice@example.com",
  script: "require [\"fileinto\"];\nif true { fileinto \"INBOX\"; }",
  active: true,
};

const emptySieve: SieveResponse = {
  email: "alice@example.com",
  active: false,
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useSieveScript", () => {
  it("fetches sieve script successfully", async () => {
    mockGetSieveScript.mockResolvedValue(mockSieve);

    const { result } = renderHook(() => useSieveScript("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockSieve);
  });

  it("returns empty script when none exists", async () => {
    mockGetSieveScript.mockResolvedValue(emptySieve);

    const { result } = renderHook(() => useSieveScript("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.active).toBe(false);
    expect(result.current.data?.script).toBeUndefined();
  });

  it("does not fetch when email is empty", () => {
    const { result } = renderHook(() => useSieveScript(""), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("handles fetch error", async () => {
    mockGetSieveScript.mockRejectedValue(new Error("Failed to fetch script"));

    const { result } = renderHook(() => useSieveScript("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useSetSieveScript", () => {
  it("sets sieve script successfully", async () => {
    mockSetSieveScript.mockResolvedValue(mockSieve);

    const { result } = renderHook(() => useSetSieveScript(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      email: "alice@example.com",
      script: 'require ["fileinto"];\nif true { fileinto "INBOX"; }',
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockSieve);
  });

  it("handles set error", async () => {
    mockSetSieveScript.mockRejectedValue(new Error("Failed to set script"));

    const { result } = renderHook(() => useSetSieveScript(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      email: "alice@example.com",
      script: "invalid script",
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useDeleteSieveScript", () => {
  it("deletes sieve script successfully", async () => {
    mockDeleteSieveScript.mockResolvedValue(emptySieve);

    const { result } = renderHook(() => useDeleteSieveScript(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("alice@example.com");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it("handles delete error", async () => {
    mockDeleteSieveScript.mockRejectedValue(new Error("Failed to delete script"));

    const { result } = renderHook(() => useDeleteSieveScript(), {
      wrapper: createWrapper(),
    });

    result.current.mutate("alice@example.com");

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
