import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import type { PaginatedResponse, Domain, Alias } from "@/types/api";
import AliasesPage from "@/app/aliases/page";
import { CatchallSection } from "@/app/aliases/catchall-section";
import { AliasTable, type GroupedAlias } from "@/app/aliases/alias-table";

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
  pagination: { page: 1, limit: 50, total: 1 },
};

const mockAliasesWithCatchall: PaginatedResponse<Alias> = {
  data: [
    {
      id: 1,
      sourceAddress: "info@example.com",
      destinationAddress: "admin@example.com",
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    },
    {
      id: 2,
      sourceAddress: "@example.com",
      destinationAddress: "admin@example.com",
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    },
  ],
  pagination: { page: 1, limit: 50, total: 2 },
};

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.restoreAllMocks();
});

function mockFetchResponses(responses: unknown[]) {
  const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
  let callIndex = 0;
  fetchMock.mockImplementation(() => {
    const resp = responses[callIndex % responses.length];
    callIndex++;
    return Promise.resolve({
      ok: true,
      status: 200,
      json: () => Promise.resolve(resp),
    });
  });
}

function mockFetchByUrl(urlMap: Record<string, unknown>) {
  const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
  fetchMock.mockImplementation((url: string) => {
    const sortedEntries = Object.entries(urlMap).sort(
      ([a], [b]) => b.length - a.length,
    );
    const matched = sortedEntries.find(([pattern]) => url.includes(pattern));
    const body = matched
      ? matched[1]
      : { data: [], pagination: { page: 1, limit: 20, total: 0 } };
    return Promise.resolve({
      ok: true,
      status: 200,
      json: () => Promise.resolve(body),
    });
  });
}

describe("AliasTable", () => {
  it("renders catch-all badge when isCatchall is true", () => {
    const catchallAlias: GroupedAlias = {
      sourceAddress: "@example.com",
      destinations: ["admin@example.com"],
      ids: [1],
      isCatchall: true,
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    };

    render(
      <AliasTable
        aliases={[catchallAlias]}
        isLoading={false}
        onDelete={vi.fn()}
        onEdit={vi.fn()}
        onBulkDelete={vi.fn()}
      />,
    );

    expect(screen.getByTestId("catchall-badge")).toBeDefined();
    expect(screen.getByText("catch-all")).toBeDefined();
  });

  it("does not render catch-all badge for normal aliases", () => {
    const normalAlias: GroupedAlias = {
      sourceAddress: "info@example.com",
      destinations: ["admin@example.com"],
      ids: [1],
      isCatchall: false,
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    };

    render(
      <AliasTable
        aliases={[normalAlias]}
        isLoading={false}
        onDelete={vi.fn()}
        onEdit={vi.fn()}
        onBulkDelete={vi.fn()}
      />,
    );

    expect(screen.queryByTestId("catchall-badge")).toBeNull();
  });

  it("shows empty state when no aliases", () => {
    render(
      <AliasTable
        aliases={[]}
        isLoading={false}
        onDelete={vi.fn()}
        onEdit={vi.fn()}
        onBulkDelete={vi.fn()}
      />,
    );

    expect(screen.getByTestId("alias-empty-state")).toBeDefined();
    expect(screen.getByText("No aliases yet")).toBeDefined();
  });
});

describe("CatchallSection", () => {
  it("renders disabled state with destination input when no catchall alias", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <CatchallSection
          domainId={1}
          domainName="example.com"
          catchallAlias={null}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTestId("catchall-section")).toBeDefined();
    expect(screen.getByTestId("catchall-toggle")).toBeDefined();
    expect(screen.getByTestId("catchall-destination-input")).toBeDefined();
    expect(screen.getByTestId("catchall-enable-btn")).toBeDefined();
    expect(screen.getByText("Disabled")).toBeDefined();
  });

  it("renders enabled state with destination display when catchall alias exists", () => {
    const catchallAlias: GroupedAlias = {
      sourceAddress: "@example.com",
      destinations: ["admin@example.com"],
      ids: [1],
      isCatchall: true,
      active: true,
      createdAt: "2024-01-01T00:00:00Z",
    };

    render(
      <QueryClientProvider client={new QueryClient()}>
        <CatchallSection
          domainId={1}
          domainName="example.com"
          catchallAlias={catchallAlias}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByTestId("catchall-section")).toBeDefined();
    expect(screen.getByTestId("catchall-toggle")).toBeDefined();
    expect(screen.getByTestId("catchall-destination")).toBeDefined();
    expect(screen.getByText("Enabled")).toBeDefined();
    expect(screen.getByText("admin@example.com")).toBeDefined();
  });

  it("disables enable button when destination is empty", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <CatchallSection
          domainId={1}
          domainName="example.com"
          catchallAlias={null}
        />
      </QueryClientProvider>,
    );

    const enableBtn = screen.getByTestId("catchall-enable-btn");
    expect(enableBtn).toBeDisabled();
  });

  it("enables enable button when destination is entered", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <CatchallSection
          domainId={1}
          domainName="example.com"
          catchallAlias={null}
        />
      </QueryClientProvider>,
    );

    const input = screen.getByTestId("catchall-destination-input");
    fireEvent.change(input, { target: { value: "admin@example.com" } });

    const enableBtn = screen.getByTestId("catchall-enable-btn");
    expect(enableBtn).not.toBeDisabled();
  });

  it("shows domain name in description", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <CatchallSection
          domainId={1}
          domainName="example.com"
          catchallAlias={null}
        />
      </QueryClientProvider>,
    );

    expect(screen.getByText("@example.com")).toBeDefined();
  });
});

describe("AliasesPage", () => {
  it("renders loading state", () => {
    mockFetchResponses([]);
    render(<AliasesPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("aliases-loading")).toBeDefined();
  });

  it("renders no domains state", async () => {
    mockFetchResponses([{ data: [], pagination: { page: 1, limit: 20, total: 0 } }]);
    render(<AliasesPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByTestId("aliases-no-domains")).toBeDefined();
    });
  });

  it("renders aliases table with domain selector", async () => {
    mockFetchResponses([mockDomains, mockAliases]);
    render(<AliasesPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByTestId("domain-selector")).toBeDefined();
    });
  });

  it.skip("renders catch-all section when domains exist", async () => {
    mockFetchByUrl({ "/api/domains/1/aliases": mockAliases, "/api/domains": mockDomains });
    render(<AliasesPage />, { wrapper: createWrapper() });

    await waitFor(
      () => {
        expect(screen.getByTestId("catchall-section")).toBeDefined();
      },
      { timeout: 3000 },
    );
  });

  it.skip("shows catch-all enabled when aliases include catchall", async () => {
    mockFetchByUrl({ "/api/domains/1/aliases": mockAliasesWithCatchall, "/api/domains": mockDomains });
    render(<AliasesPage />, { wrapper: createWrapper() });

    await waitFor(
      () => {
        expect(screen.getByTestId("catchall-destination")).toBeDefined();
      },
      { timeout: 3000 },
    );
  });
});
