import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import { MailQueueWidget } from "@/app/(dashboard)/mail-queue-widget";
import type { MailQueueStatus } from "@/types/api";

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

function mockFetchImplementation(handler: (url: string) => unknown) {
  (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(
    (url: string) =>
      Promise.resolve({
        ok: true,
        status: 200,
        json: () => Promise.resolve(handler(url)),
      }),
  );
}

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("MailQueueWidget", () => {
  it("renders loading skeleton initially", () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise(() => {}),
    );

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    const skeletons = document.querySelectorAll("[data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders data-testid after data loads", async () => {
    const queueData: MailQueueStatus = {
      total: 5,
      deferred: 0,
      active: 3,
      oldestAge: "2m",
    };

    mockFetchImplementation(() => queueData);

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    const widget = await screen.findByTestId("mail-queue-widget");
    expect(widget).toBeInTheDocument();
  });

  it("shows green indicator when deferred is 0", async () => {
    const queueData: MailQueueStatus = {
      total: 3,
      deferred: 0,
      active: 3,
      oldestAge: "1m",
    };

    mockFetchImplementation(() => queueData);

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    await screen.findByTestId("mail-queue-widget");

    expect(screen.getByText("Healthy")).toBeInTheDocument();
  });

  it("shows amber indicator when deferred is between 1 and 10", async () => {
    const queueData: MailQueueStatus = {
      total: 15,
      deferred: 5,
      active: 10,
      oldestAge: "5m",
    };

    mockFetchImplementation(() => queueData);

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    await screen.findByTestId("mail-queue-widget");

    const deferredElements = screen.getAllByText("Deferred");
    expect(deferredElements.length).toBeGreaterThanOrEqual(2);
  });

  it("shows red indicator when deferred is greater than 10", async () => {
    const queueData: MailQueueStatus = {
      total: 30,
      deferred: 15,
      active: 15,
      oldestAge: "20m",
    };

    mockFetchImplementation(() => queueData);

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    await screen.findByTestId("mail-queue-widget");

    expect(screen.getByText("Backed up")).toBeInTheDocument();
  });

  it("displays total, active, and deferred counts", async () => {
    const queueData: MailQueueStatus = {
      total: 8,
      deferred: 2,
      active: 6,
      oldestAge: "3m",
    };

    mockFetchImplementation(() => queueData);

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    await screen.findByTestId("mail-queue-widget");

    expect(screen.getByText("8")).toBeInTheDocument();
    expect(screen.getByText("6")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getAllByText("Deferred").length).toBeGreaterThanOrEqual(2);
  });

  it("renders refresh button with correct data-testid", async () => {
    const queueData: MailQueueStatus = {
      total: 0,
      deferred: 0,
      active: 0,
      oldestAge: "",
    };

    mockFetchImplementation(() => queueData);

    render(<MailQueueWidget />, { wrapper: createWrapper() });

    await screen.findByTestId("mail-queue-widget");

    const refreshBtn = screen.getByTestId("mail-queue-refresh");
    expect(refreshBtn).toBeInTheDocument();
    expect(refreshBtn).toHaveAttribute("aria-label", "Refresh mail queue");
  });
});
