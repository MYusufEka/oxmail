import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { PaginatedResponse, Domain, DomainHealthResult } from "@/types/api";

const { mockGetDomains, mockGetDomainHealth, mockCreateDomain, mockDeleteDomain } = vi.hoisted(() => ({
  mockGetDomains: vi.fn(),
  mockGetDomainHealth: vi.fn(),
  mockCreateDomain: vi.fn(),
  mockDeleteDomain: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getDomains: mockGetDomains,
    getDomainHealth: mockGetDomainHealth,
    createDomain: mockCreateDomain,
    deleteDomain: mockDeleteDomain,
  },
}));

import {
  useDomains,
  useDomainHealth,
  useCreateDomain,
  useDeleteDomain,
} from "@/hooks/use-domains";

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

const mockDomains: PaginatedResponse<Domain> = {
  data: [
    {
      id: 1,
      name: "example.com",
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const emptyResponse: PaginatedResponse<Domain> = {
  data: [],
  pagination: { page: 1, limit: 20, total: 0 },
};

const mockDomainHealth: DomainHealthResult = {
  domain: "example.com",
  status: "healthy",
  checks: [
    { name: "mx", status: "pass", detail: "MX record found" },
  ],
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useDomains", () => {
  it("fetches domains successfully", async () => {
    mockGetDomains.mockResolvedValue(mockDomains);

    const { result } = renderHook(() => useDomains(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDomains);
  });

  it("returns empty data when no domains exist", async () => {
    mockGetDomains.mockResolvedValue(emptyResponse);

    const { result } = renderHook(() => useDomains(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(0);
  });

  it("handles fetch error", async () => {
    mockGetDomains.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useDomains(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useDomainHealth", () => {
  it("fetches domain health successfully", async () => {
    mockGetDomainHealth.mockResolvedValue(mockDomainHealth);

    const { result } = renderHook(() => useDomainHealth("example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDomainHealth);
  });

  it("handles health check error", async () => {
    mockGetDomainHealth.mockRejectedValue(new Error("Health check failed"));

    const { result } = renderHook(() => useDomainHealth("example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useCreateDomain", () => {
  it("creates domain successfully", async () => {
    const newDomain: Domain = {
      id: 2,
      name: "new.com",
      active: true,
      createdAt: "2024-01-02T00:00:00Z",
      updatedAt: "2024-01-02T00:00:00Z",
    };
    mockCreateDomain.mockResolvedValue(newDomain);

    const { result } = renderHook(() => useCreateDomain(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: "new.com" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(newDomain);
  });

  it("handles create error", async () => {
    mockCreateDomain.mockRejectedValue(new Error("Creation failed"));

    const { result } = renderHook(() => useCreateDomain(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: "bad.com" });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useDeleteDomain", () => {
  it("deletes domain successfully", async () => {
    mockDeleteDomain.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteDomain(), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDeleteDomain).toHaveBeenCalledWith(1);
  });

  it("handles delete error", async () => {
    mockDeleteDomain.mockRejectedValue(new Error("Delete failed"));

    const { result } = renderHook(() => useDeleteDomain(), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
