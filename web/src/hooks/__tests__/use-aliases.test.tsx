import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { PaginatedResponse, Alias } from "@/types/api";

const { mockGetAliases, mockCreateAlias, mockUpdateAlias, mockDeleteAlias } = vi.hoisted(() => ({
  mockGetAliases: vi.fn(),
  mockCreateAlias: vi.fn(),
  mockUpdateAlias: vi.fn(),
  mockDeleteAlias: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getAliases: mockGetAliases,
    createAlias: mockCreateAlias,
    updateAlias: mockUpdateAlias,
    deleteAlias: mockDeleteAlias,
  },
}));

import {
  useAliases,
  useCreateAlias,
  useUpdateAlias,
  useDeleteAlias,
} from "@/hooks/use-aliases";

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

const mockAliases: PaginatedResponse<Alias> = {
  data: [
    {
      id: 1,
      sourceAddress: "info@example.com",
      destinationAddress: "admin@example.com",
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const emptyResponse: PaginatedResponse<Alias> = {
  data: [],
  pagination: { page: 1, limit: 20, total: 0 },
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useAliases", () => {
  it("fetches aliases successfully", async () => {
    mockGetAliases.mockResolvedValue(mockAliases);

    const { result } = renderHook(() => useAliases(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockAliases);
    expect(mockGetAliases).toHaveBeenCalledWith(1, undefined);
  });

  it("returns empty data when no aliases exist", async () => {
    mockGetAliases.mockResolvedValue(emptyResponse);

    const { result } = renderHook(() => useAliases(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(0);
  });

  it("does not fetch when domainId is 0", () => {
    mockGetAliases.mockResolvedValue(mockAliases);

    const { result } = renderHook(() => useAliases(0), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("handles fetch error", async () => {
    mockGetAliases.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useAliases(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useCreateAlias", () => {
  it("creates alias successfully", async () => {
    const newAlias: Alias = {
      id: 2,
      sourceAddress: "support@example.com",
      destinationAddress: "admin@example.com",
      active: true,
      createdAt: "2024-01-02T00:00:00Z",
    };
    mockCreateAlias.mockResolvedValue(newAlias);

    const { result } = renderHook(() => useCreateAlias(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      sourceAddress: "support@example.com",
      destinationAddress: "admin@example.com",
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(newAlias);
  });

  it("handles create error", async () => {
    mockCreateAlias.mockRejectedValue(new Error("Creation failed"));

    const { result } = renderHook(() => useCreateAlias(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      sourceAddress: "support@example.com",
      destinationAddress: "admin@example.com",
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useUpdateAlias", () => {
  it("updates alias successfully", async () => {
    const updated: Alias = {
      id: 1,
      sourceAddress: "new@example.com",
      destinationAddress: "admin@example.com",
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    };
    mockUpdateAlias.mockResolvedValue(updated);

    const { result } = renderHook(() => useUpdateAlias(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      aliasId: 1,
      payload: {
        sourceAddress: "new@example.com",
        destinationAddress: "admin@example.com",
      },
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(updated);
  });

  it("handles update error", async () => {
    mockUpdateAlias.mockRejectedValue(new Error("Update failed"));

    const { result } = renderHook(() => useUpdateAlias(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      aliasId: 1,
      payload: {
        sourceAddress: "bad@example.com",
        destinationAddress: "admin@example.com",
      },
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useDeleteAlias", () => {
  it("deletes alias successfully", async () => {
    mockDeleteAlias.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteAlias(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it("handles delete error", async () => {
    mockDeleteAlias.mockRejectedValue(new Error("Delete failed"));

    const { result } = renderHook(() => useDeleteAlias(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
