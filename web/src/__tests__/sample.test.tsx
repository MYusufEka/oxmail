import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";

// Mock next/navigation
vi.mock("next/navigation", () => ({
  usePathname: () => "/",
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

import DashboardPage from "@/app/page";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };
}

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn(() => new Promise(() => {})));
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("Dashboard page", () => {
  it("renders the Dashboard heading", () => {
    render(<DashboardPage />, { wrapper: createWrapper() });
    expect(screen.getByRole("heading", { level: 2 })).toHaveTextContent(
      "Dashboard"
    );
  });

  it("renders the Service Health section", () => {
    render(<DashboardPage />, { wrapper: createWrapper() });
    expect(screen.getByText("Service Health")).toBeInTheDocument();
  });
});
