import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, afterEach } from "vitest";
import { MailStatsCard } from "./mail-stats-card";
import type { DailyStat } from "@/types/api";

const mockStats: DailyStat[] = [
  { date: "2026-07-10", sent: 5, received: 12, bounced: 1, spamCaught: 3 },
  { date: "2026-07-11", sent: 8, received: 15, bounced: 2, spamCaught: 1 },
  { date: "2026-07-12", sent: 3, received: 7, bounced: 0, spamCaught: 0 },
  { date: "2026-07-13", sent: 0, received: 2, bounced: 0, spamCaught: 0 },
  { date: "2026-07-14", sent: 10, received: 20, bounced: 3, spamCaught: 5 },
  { date: "2026-07-15", sent: 1, received: 4, bounced: 0, spamCaught: 0 },
  { date: "2026-07-16", sent: 6, received: 9, bounced: 1, spamCaught: 2 },
];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}

describe("MailStatsCard", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders totals from API data", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue({
      ok: true,
      json: async () => mockStats,
    } as Response);

    render(<MailStatsCard />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("Sent")).toBeInTheDocument();
      expect(screen.getByText("Received")).toBeInTheDocument();
      expect(screen.getByText("Bounced")).toBeInTheDocument();
      expect(screen.getByText("Spam")).toBeInTheDocument();
    });

    expect(screen.getByText("33")).toBeInTheDocument();
    expect(screen.getByText("69")).toBeInTheDocument();
    expect(screen.getByText("7")).toBeInTheDocument();
    expect(screen.getByText("11")).toBeInTheDocument();
  });

  it("renders 7 day columns in chart", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue({
      ok: true,
      json: async () => mockStats,
    } as Response);

    render(<MailStatsCard />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("Jul 10")).toBeInTheDocument();
    });

    const dateLabels = ["Jul 11", "Jul 12", "Jul 13", "Jul 14", "Jul 15", "Jul 16"];
    for (const label of dateLabels) {
      expect(screen.getByText(label)).toBeInTheDocument();
    }
  });

  it("shows empty state when all stats are zero", async () => {
    vi.spyOn(global, "fetch").mockResolvedValue({
      ok: true,
      json: async () => [],
    } as Response);

    render(<MailStatsCard />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("No mail activity in the last 7 days")).toBeInTheDocument();
    });
  });
});
