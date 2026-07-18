import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type { Contact } from "@/types/api";

const { mockGetContacts, mockCreateContact, mockUpdateContact, mockDeleteContact } = vi.hoisted(() => ({
  mockGetContacts: vi.fn(),
  mockCreateContact: vi.fn(),
  mockUpdateContact: vi.fn(),
  mockDeleteContact: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getContacts: mockGetContacts,
    createContact: mockCreateContact,
    updateContact: mockUpdateContact,
    deleteContact: mockDeleteContact,
  },
}));

import {
  useContacts,
  useCreateContact,
  useUpdateContact,
  useDeleteContact,
} from "@/hooks/use-contacts";

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

const mockContacts: Contact[] = [
  {
    id: 1,
    userEmail: "alice@example.com",
    name: "Alice",
    email: "alice@test.com",
    createdAt: "2024-01-01T00:00:00Z",
  },
];

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useContacts", () => {
  it("fetches contacts successfully", async () => {
    mockGetContacts.mockResolvedValue(mockContacts);

    const { result } = renderHook(() => useContacts("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockContacts);
  });

  it("returns empty array when no contacts exist", async () => {
    mockGetContacts.mockResolvedValue([]);

    const { result } = renderHook(() => useContacts("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(0);
  });

  it("does not fetch when email is empty", () => {
    const { result } = renderHook(() => useContacts(""), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("handles fetch error", async () => {
    mockGetContacts.mockRejectedValue(new Error("Network error"));

    const { result } = renderHook(() => useContacts("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useCreateContact", () => {
  it("creates contact successfully", async () => {
    const newContact: Contact = {
      id: 2,
      userEmail: "alice@example.com",
      name: "Bob",
      email: "bob@test.com",
      createdAt: "2024-01-02T00:00:00Z",
    };
    mockCreateContact.mockResolvedValue(newContact);

    const { result } = renderHook(() => useCreateContact("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: "Bob", email: "bob@test.com" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(newContact);
  });

  it("handles create error", async () => {
    mockCreateContact.mockRejectedValue(new Error("Creation failed"));

    const { result } = renderHook(() => useCreateContact("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ name: "Bob", email: "bob@test.com" });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useUpdateContact", () => {
  it("updates contact successfully", async () => {
    const updated: Contact = {
      id: 1,
      userEmail: "alice@example.com",
      name: "Alice Updated",
      email: "alice@new.com",
      createdAt: "2024-01-01T00:00:00Z",
    };
    mockUpdateContact.mockResolvedValue(updated);

    const { result } = renderHook(() => useUpdateContact("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      contactId: 1,
      payload: { name: "Alice Updated", email: "alice@new.com" },
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(updated);
  });
});

describe("useDeleteContact", () => {
  it("deletes contact successfully", async () => {
    mockDeleteContact.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeleteContact("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });

  it("handles delete error", async () => {
    mockDeleteContact.mockRejectedValue(new Error("Delete failed"));

    const { result } = renderHook(() => useDeleteContact("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});
