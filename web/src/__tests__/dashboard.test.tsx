import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import { KpiCards } from "@/app/(dashboard)/kpi-cards";
import { ServiceHealthGrid } from "@/app/(dashboard)/service-health-grid";
import { RecentActivity } from "@/app/(dashboard)/recent-activity";
import type { PaginatedResponse, HealthStatus, LogEntry, Domain } from "@/types/api";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
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
    { id: 2, name: "test.com", active: true, createdAt: "2024-01-01T00:00:00Z", updatedAt: "2024-01-01T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 1, total: 2 },
};

const mockHealth: HealthStatus = {
  status: "healthy",
  services: {
    postfix: { status: "up", latencyMs: 5 },
    dovecot: { status: "up", latencyMs: 3 },
    rspamd: { status: "up", latencyMs: 10 },
    redis: { status: "up", latencyMs: 1 },
  },
};

const mockLogs: PaginatedResponse<LogEntry> = {
  data: [
    { id: 1, timestamp: "2024-01-01T10:00:00Z", service: "postfix", level: "info", message: "Mail delivered to user@example.com" },
    { id: 2, timestamp: "2024-01-01T10:01:00Z", service: "dovecot", level: "info", message: "IMAP login successful" },
    { id: 3, timestamp: "2024-01-01T10:02:00Z", service: "rspamd", level: "warn", message: "Spam score high for incoming message" },
  ],
  pagination: { page: 1, limit: 10, total: 3 },
};

function mockFetchResponses(responses: Record<string, unknown>) {
  (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(
    (url: string) => {
      for (const [pattern, responseData] of Object.entries(responses)) {
        if (url.includes(pattern)) {
          return Promise.resolve({
            ok: true,
            status: 200,
            json: () => Promise.resolve(responseData),
          });
        }
      }
      return Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ data: [], pagination: { page: 1, limit: 20, total: 0 } }),
      });
    },
  );
}

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("KpiCards", () => {
  it("renders loading skeletons initially", () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<KpiCards />, { wrapper: createWrapper() });

    const skeletons = document.querySelectorAll("[data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders KPI values after data loads", async () => {
    mockFetchResponses({
      "/api/domains": mockDomains,
      "/api/logs": mockLogs,
    });

    render(<KpiCards />, { wrapper: createWrapper() });

    expect(await screen.findByText("Total Domains")).toBeInTheDocument();
    expect(await screen.findByText("2")).toBeInTheDocument();
    expect(screen.getByText("Emails Today")).toBeInTheDocument();
    expect(screen.getByText("Total Users")).toBeInTheDocument();
    expect(screen.getByText("Uptime")).toBeInTheDocument();
  });
});

describe("ServiceHealthGrid", () => {
  it("renders loading skeletons initially", () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<ServiceHealthGrid />, { wrapper: createWrapper() });

    const skeletons = document.querySelectorAll("[data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders 5 service cards with data-testid", async () => {
    mockFetchResponses({
      "/api/health": mockHealth,
    });

    render(<ServiceHealthGrid />, { wrapper: createWrapper() });

    const cards = await screen.findAllByTestId("service-card");
    expect(cards).toHaveLength(5);
  });

  it("displays service names and status badges", async () => {
    mockFetchResponses({
      "/api/health": mockHealth,
    });

    render(<ServiceHealthGrid />, { wrapper: createWrapper() });

    expect(await screen.findByText("Postfix")).toBeInTheDocument();
    expect(screen.getByText("Dovecot")).toBeInTheDocument();
    expect(screen.getByText("Rspamd")).toBeInTheDocument();
    expect(screen.getByText("Redis")).toBeInTheDocument();
    expect(screen.getByText("API")).toBeInTheDocument();
  });

  it("shows latency values", async () => {
    mockFetchResponses({
      "/api/health": mockHealth,
    });

    render(<ServiceHealthGrid />, { wrapper: createWrapper() });

    expect(await screen.findByText("5ms")).toBeInTheDocument();
    expect(screen.getByText("3ms")).toBeInTheDocument();
    expect(screen.getByText("10ms")).toBeInTheDocument();
    expect(screen.getByText("1ms")).toBeInTheDocument();
  });
});

describe("RecentActivity", () => {
  it("renders loading skeletons initially", () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<RecentActivity />, { wrapper: createWrapper() });

    const skeletons = document.querySelectorAll("[data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders log entries after data loads", async () => {
    mockFetchResponses({
      "/api/logs": mockLogs,
    });

    render(<RecentActivity />, { wrapper: createWrapper() });

    expect(await screen.findByText("Recent Activity")).toBeInTheDocument();
    expect(await screen.findByText(/Mail delivered/)).toBeInTheDocument();
    expect(screen.getByText(/IMAP login/)).toBeInTheDocument();
    expect(screen.getByText(/Spam score/)).toBeInTheDocument();
  });

  it("shows service names in log entries", async () => {
    mockFetchResponses({
      "/api/logs": mockLogs,
    });

    render(<RecentActivity />, { wrapper: createWrapper() });

    expect(await screen.findByText("postfix")).toBeInTheDocument();
    expect(screen.getByText("dovecot")).toBeInTheDocument();
    expect(screen.getByText("rspamd")).toBeInTheDocument();
  });

  it("shows empty state when no logs", async () => {
    mockFetchResponses({
      "/api/logs": { data: [], pagination: { page: 1, limit: 10, total: 0 } },
    });

    render(<RecentActivity />, { wrapper: createWrapper() });

    expect(await screen.findByText("No recent activity")).toBeInTheDocument();
  });
});
