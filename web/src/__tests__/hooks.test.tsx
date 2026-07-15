import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import { useDomains, useCreateDomain, useDeleteDomain } from "@/hooks/use-domains";
import { useUsers, useCreateUser } from "@/hooks/use-users";
import { useAliases } from "@/hooks/use-aliases";
import { useDkim, useGenerateDkim } from "@/hooks/use-dkim";
import { useHealth } from "@/hooks/use-health";
import { useLogs } from "@/hooks/use-logs";
import { useInbox, useSendMail } from "@/hooks/use-mail";
import type { PaginatedResponse, Domain, User, Alias, DKIMKey, HealthStatus, LogEntry, MailMessage } from "@/types/api";

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
    { id: 1, name: "example.com", active: true, createdAt: "2024-01-01T00:00:00Z", updatedAt: "2024-01-01T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const mockUsers: PaginatedResponse<User> = {
  data: [
    { id: 1, email: "admin@example.com", domainId: 1, quota: 1024, active: true, createdAt: "2024-01-01T00:00:00Z", updatedAt: "2024-01-01T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const mockAliases: PaginatedResponse<Alias> = {
  data: [
    { id: 1, sourceAddress: "info@example.com", destinationAddress: "admin@example.com", active: true, createdAt: "2024-01-01T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const mockDkim: DKIMKey = {
  domain: "example.com",
  selector: "default",
  publicKey: "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
  dnsRecord: "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...",
  createdAt: "2024-01-01T00:00:00Z",
};

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

const mockLogs: PaginatedResponse<LogEntry> = {
  data: [
    { id: 1, timestamp: "2024-01-01T00:00:00Z", service: "postfix", level: "info", message: "Mail delivered" },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

const mockInbox: PaginatedResponse<MailMessage> = {
  data: [
    { id: 1, from: "sender@test.com", to: ["admin@example.com"], subject: "Hello", read: false, receivedAt: "2024-01-01T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 1 },
};

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.restoreAllMocks();
});

function mockFetchSuccess(responseData: unknown, status = 200) {
  (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
    ok: true,
    status,
    json: () => Promise.resolve(responseData),
  });
}

function mockFetchError(status: number, errorBody: { error: { code: string; message: string } }) {
  (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
    ok: false,
    status,
    json: () => Promise.resolve(errorBody),
  });
}

describe("useDomains", () => {
  it("fetches domains successfully", async () => {
    mockFetchSuccess(mockDomains);

    const { result } = renderHook(() => useDomains(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDomains);
  });

  it("handles fetch error", async () => {
    mockFetchError(500, { error: { code: "INTERNAL_ERROR", message: "Server error" } });

    const { result } = renderHook(() => useDomains(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeDefined();
  });
});

describe("useCreateDomain", () => {
  it("creates a domain and invalidates cache", async () => {
    const newDomain: Domain = { id: 2, name: "new.com", active: true, createdAt: "2024-01-02T00:00:00Z", updatedAt: "2024-01-02T00:00:00Z" };
    mockFetchSuccess(newDomain);

    const { result } = renderHook(() => useCreateDomain(), { wrapper: createWrapper() });

    result.current.mutate({ name: "new.com" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(newDomain);
  });
});

describe("useDeleteDomain", () => {
  it("deletes a domain", async () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: () => Promise.resolve(undefined),
    });

    const { result } = renderHook(() => useDeleteDomain(), { wrapper: createWrapper() });

    result.current.mutate(1);

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("useUsers", () => {
  it("fetches users for a domain", async () => {
    mockFetchSuccess(mockUsers);

    const { result } = renderHook(() => useUsers(1), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockUsers);
  });

  it("does not fetch when domainId is 0", () => {
    const { result } = renderHook(() => useUsers(0), { wrapper: createWrapper() });

    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useCreateUser", () => {
  it("creates a user", async () => {
    const newUser: User = { id: 2, email: "new@example.com", domainId: 1, quota: 512, active: true, createdAt: "2024-01-02T00:00:00Z", updatedAt: "2024-01-02T00:00:00Z" };
    mockFetchSuccess(newUser);

    const { result } = renderHook(() => useCreateUser(1), { wrapper: createWrapper() });

    result.current.mutate({ email: "new@example.com", password: "securepass" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(newUser);
  });
});

describe("useAliases", () => {
  it("fetches aliases for a domain", async () => {
    mockFetchSuccess(mockAliases);

    const { result } = renderHook(() => useAliases(1), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockAliases);
  });
});

describe("useDkim", () => {
  it("fetches DKIM key for a domain", async () => {
    mockFetchSuccess(mockDkim);

    const { result } = renderHook(() => useDkim("example.com"), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDkim);
  });

  it("does not fetch when domain is empty", () => {
    const { result } = renderHook(() => useDkim(""), { wrapper: createWrapper() });

    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useGenerateDkim", () => {
  it("generates DKIM key", async () => {
    mockFetchSuccess(mockDkim);

    const { result } = renderHook(() => useGenerateDkim(), { wrapper: createWrapper() });

    result.current.mutate("example.com");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockDkim);
  });
});

describe("useHealth", () => {
  it("fetches health status", async () => {
    mockFetchSuccess(mockHealth);

    const { result } = renderHook(() => useHealth(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockHealth);
  });
});

describe("useLogs", () => {
  it("fetches logs", async () => {
    mockFetchSuccess(mockLogs);

    const { result } = renderHook(() => useLogs(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockLogs);
  });
});

describe("useInbox", () => {
  it("fetches inbox for a user", async () => {
    mockFetchSuccess(mockInbox);

    const { result } = renderHook(() => useInbox(1), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockInbox);
  });

  it("does not fetch when userId is 0", () => {
    const { result } = renderHook(() => useInbox(0), { wrapper: createWrapper() });

    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useSendMail", () => {
  it("sends mail successfully", async () => {
    const sendResponse = { messageId: "msg-123", status: "queued" as const };
    mockFetchSuccess(sendResponse);

    const { result } = renderHook(() => useSendMail(), { wrapper: createWrapper() });

    result.current.mutate({
      from: "admin@example.com",
      to: ["user@test.com"],
      subject: "Test",
      bodyText: "Hello",
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(sendResponse);
  });
});
