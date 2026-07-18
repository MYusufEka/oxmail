import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { PaginatedResponse, User, UserImportResult } from "@/types/api";

const { mockGetUsers, mockCreateUser, mockDeleteUser, mockUpdateUser, mockImportUsers } = vi.hoisted(() => ({
  mockGetUsers: vi.fn(),
  mockCreateUser: vi.fn(),
  mockDeleteUser: vi.fn(),
  mockUpdateUser: vi.fn(),
  mockImportUsers: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getUsers: mockGetUsers,
    createUser: mockCreateUser,
    deleteUser: mockDeleteUser,
    updateUser: mockUpdateUser,
    importUsers: mockImportUsers,
  },
}));

import {
  useUsers,
  useCreateUser,
  useDeleteUser,
  useUpdateUser,
  useImportUsers,
} from "@/hooks/use-users";

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

const mockUsers: PaginatedResponse<User> = {
  data: [
    {
      id: 1,
      email: "admin@example.com",
      domainId: 1,
      quota: 1024,
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const emptyResponse: PaginatedResponse<User> = {
  data: [],
  pagination: { page: 1, limit: 20, total: 0 },
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useUsers", () => {
  it("fetches users successfully", async () => {
    mockGetUsers.mockResolvedValue(mockUsers);

    const { result } = renderHook(() => useUsers(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockUsers);
  });

  it("returns empty data when no users exist", async () => {
    mockGetUsers.mockResolvedValue(emptyResponse);

    const { result } = renderHook(() => useUsers(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.data).toHaveLength(0);
  });

  it("does not fetch when domainId is 0", () => {
    const { result } = renderHook(() => useUsers(0), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("handles fetch error", async () => {
    mockGetUsers.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useUsers(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useCreateUser", () => {
  it("creates user successfully", async () => {
    const newUser: User = {
      id: 2,
      email: "new@example.com",
      domainId: 1,
      quota: 512,
      active: true,
      createdAt: "2024-01-02T00:00:00Z",
      updatedAt: "2024-01-02T00:00:00Z",
    };
    mockCreateUser.mockResolvedValue(newUser);

    const { result } = renderHook(() => useCreateUser(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      email: "new@example.com",
      password: "securepass",
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(newUser);
  });

  it("handles create error", async () => {
    mockCreateUser.mockRejectedValue(new Error("Creation failed"));

    const { result } = renderHook(() => useCreateUser(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      email: "bad@example.com",
      password: "pass",
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useDeleteUser", () => {
  it("deletes user successfully", async () => {
    mockDeleteUser.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteUser(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockDeleteUser).toHaveBeenCalledWith(1, 1);
  });

  it("handles delete error", async () => {
    mockDeleteUser.mockRejectedValue(new Error("Delete failed"));

    const { result } = renderHook(() => useDeleteUser(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useUpdateUser", () => {
  it("updates user successfully", async () => {
    const updated: User = {
      id: 1,
      email: "admin@example.com",
      domainId: 1,
      displayName: "Updated Name",
      quota: 2048,
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
      updatedAt: "2024-01-02T00:00:00Z",
    };
    mockUpdateUser.mockResolvedValue(updated);

    const { result } = renderHook(() => useUpdateUser(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      userId: 1,
      payload: { displayName: "Updated Name", quota: 2048 },
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(updated);
  });

  it("handles update error", async () => {
    mockUpdateUser.mockRejectedValue(new Error("Update failed"));

    const { result } = renderHook(() => useUpdateUser(1), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      userId: 1,
      payload: { quota: -1 },
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useImportUsers", () => {
  it("imports users successfully", async () => {
    const importResult: UserImportResult = {
      created: 5,
      skipped: 1,
      errors: [],
    };
    mockImportUsers.mockResolvedValue(importResult);

    const { result } = renderHook(() => useImportUsers(1), {
      wrapper: createWrapper(),
    });

    const file = new File(["email,password\ntest@example.com,pass"], "users.csv", {
      type: "text/csv",
    });
    result.current.mutate(file);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(importResult);
  });

  it("handles import with errors", async () => {
    const importResultWithErrors: UserImportResult = {
      created: 3,
      skipped: 2,
      errors: [
        { row: 4, email: "bad@example.com", error: "Invalid email" },
      ],
    };
    mockImportUsers.mockResolvedValue(importResultWithErrors);

    const { result } = renderHook(() => useImportUsers(1), {
      wrapper: createWrapper(),
    });

    const file = new File(["data"], "users.csv", { type: "text/csv" });
    result.current.mutate(file);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.errors).toHaveLength(1);
    expect(result.current.data?.errors[0].error).toBe("Invalid email");
  });
});
