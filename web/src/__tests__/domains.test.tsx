import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import type { PaginatedResponse, Domain } from "@/types/api";

// Mock hooks
const mockUseDomains = vi.fn();
const mockUseCreateDomain = vi.fn();
const mockUseDeleteDomain = vi.fn();

vi.mock("@/hooks/use-domains", () => ({
  useDomains: () => mockUseDomains(),
  useCreateDomain: () => mockUseCreateDomain(),
  useDeleteDomain: () => mockUseDeleteDomain(),
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import DomainsPage from "@/app/domains/page";
import { DomainTable } from "@/app/domains/domain-table";
import { AddDomainDialog } from "@/app/domains/add-domain-dialog";
import { DeleteDomainDialog } from "@/app/domains/delete-domain-dialog";

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

const mockDomains: Domain[] = [
  {
    id: 1,
    name: "example.com",
    active: true,
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-01-01T00:00:00Z",
  },
  {
    id: 2,
    name: "test.org",
    active: false,
    createdAt: "2024-02-15T00:00:00Z",
    updatedAt: "2024-02-15T00:00:00Z",
  },
];

const mockPaginatedDomains: PaginatedResponse<Domain> = {
  data: mockDomains,
  pagination: { page: 1, limit: 20, total: 2 },
};

function setupDefaultMocks(overrides?: {
  domainsData?: PaginatedResponse<Domain> | undefined;
  isLoading?: boolean;
  isError?: boolean;
}) {
  mockUseDomains.mockReturnValue({
    data: overrides?.domainsData ?? mockPaginatedDomains,
    isLoading: overrides?.isLoading ?? false,
    isError: overrides?.isError ?? false,
    refetch: vi.fn(),
  });

  mockUseCreateDomain.mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  });

  mockUseDeleteDomain.mockReturnValue({
    mutate: vi.fn(),
    isPending: false,
  });
}

beforeEach(() => {
  setupDefaultMocks();
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("DomainsPage", () => {
  it("renders loading skeleton when loading", () => {
    setupDefaultMocks({ isLoading: true, domainsData: undefined });

    render(<DomainsPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("domain-table-skeleton")).toBeInTheDocument();
  });

  it("renders error state with retry button", () => {
    setupDefaultMocks({ isError: true, domainsData: undefined });

    render(<DomainsPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("domain-error-state")).toBeInTheDocument();
    expect(screen.getByTestId("retry-domains")).toBeInTheDocument();
  });

  it("renders empty state when no domains", () => {
    setupDefaultMocks({
      domainsData: { data: [], pagination: { page: 1, limit: 20, total: 0 } },
    });

    render(<DomainsPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("domain-empty-state")).toBeInTheDocument();
    expect(screen.getByText("Add your first domain to start receiving mail.")).toBeInTheDocument();
  });

  it("renders domain table with data", () => {
    render(<DomainsPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("domain-table")).toBeInTheDocument();
    expect(screen.getByText("example.com")).toBeInTheDocument();
    expect(screen.getByText("test.org")).toBeInTheDocument();
  });

  it("shows add domain button when domains exist", () => {
    render(<DomainsPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("add-domain-button")).toBeInTheDocument();
  });

  it("opens add domain dialog on button click", async () => {
    const user = userEvent.setup();
    render(<DomainsPage />, { wrapper: createWrapper() });

    await user.click(screen.getByTestId("add-domain-button"));

    expect(screen.getByTestId("add-domain-dialog")).toBeInTheDocument();
  });

  it("opens delete dialog when delete button clicked", async () => {
    const user = userEvent.setup();
    render(<DomainsPage />, { wrapper: createWrapper() });

    await user.click(screen.getByTestId("delete-domain-1"));

    expect(screen.getByTestId("delete-domain-dialog")).toBeInTheDocument();
    expect(screen.getAllByText("example.com").length).toBeGreaterThanOrEqual(1);
  });

  it("calls refetch on retry click", async () => {
    const mockRefetch = vi.fn();
    mockUseDomains.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      refetch: mockRefetch,
    });

    const user = userEvent.setup();
    render(<DomainsPage />, { wrapper: createWrapper() });

    await user.click(screen.getByTestId("retry-domains"));

    expect(mockRefetch).toHaveBeenCalled();
  });
});

describe("DomainTable", () => {
  it("renders all domain columns", () => {
    const onDelete = vi.fn();
    render(<DomainTable domains={mockDomains} onDelete={onDelete} />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText("Domain")).toBeInTheDocument();
    expect(screen.getByText("Users")).toBeInTheDocument();
    expect(screen.getByText("Aliases")).toBeInTheDocument();
    expect(screen.getByText("DKIM")).toBeInTheDocument();
    expect(screen.getByText("Created")).toBeInTheDocument();
  });

  it("renders domain names", () => {
    const onDelete = vi.fn();
    render(<DomainTable domains={mockDomains} onDelete={onDelete} />, {
      wrapper: createWrapper(),
    });

    expect(screen.getByText("example.com")).toBeInTheDocument();
    expect(screen.getByText("test.org")).toBeInTheDocument();
  });

  it("renders DKIM status badges", () => {
    const onDelete = vi.fn();
    render(<DomainTable domains={mockDomains} onDelete={onDelete} />, {
      wrapper: createWrapper(),
    });

    const badges = screen.getAllByTestId("dkim-status");
    expect(badges[0]).toHaveTextContent("Active");
    expect(badges[1]).toHaveTextContent("Inactive");
  });

  it("calls onDelete when trash icon clicked", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(<DomainTable domains={mockDomains} onDelete={onDelete} />, {
      wrapper: createWrapper(),
    });

    await user.click(screen.getByTestId("delete-domain-1"));

    expect(onDelete).toHaveBeenCalledWith(mockDomains[0]);
  });
});

describe("AddDomainDialog", () => {
  it("renders form when open", () => {
    render(
      <AddDomainDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        isPending={false}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByTestId("add-domain-dialog")).toBeInTheDocument();
    expect(screen.getByTestId("domain-name-input")).toBeInTheDocument();
  });

  it("validates invalid domain name", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(
      <AddDomainDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        isPending={false}
      />,
      { wrapper: createWrapper() }
    );

    const input = screen.getByTestId("domain-name-input");
    await user.type(input, "invalid");
    await user.click(screen.getByTestId("add-domain-submit"));

    await waitFor(() => {
      expect(screen.getByTestId("domain-name-error")).toBeInTheDocument();
    });
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("submits valid domain name", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(
      <AddDomainDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        isPending={false}
      />,
      { wrapper: createWrapper() }
    );

    const input = screen.getByTestId("domain-name-input");
    await user.type(input, "newdomain.com");
    await user.click(screen.getByTestId("add-domain-submit"));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith("newdomain.com");
    });
  });

  it("disables inputs when pending", () => {
    render(
      <AddDomainDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        isPending={true}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByTestId("domain-name-input")).toBeDisabled();
    expect(screen.getByTestId("add-domain-submit")).toBeDisabled();
  });
});

describe("DeleteDomainDialog", () => {
  it("renders confirmation when open", () => {
    render(
      <DeleteDomainDialog
        domain={mockDomains[0]}
        open={true}
        onOpenChange={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByTestId("delete-domain-dialog")).toBeInTheDocument();
    expect(screen.getByText("example.com")).toBeInTheDocument();
  });

  it("calls onConfirm when delete button clicked", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();

    render(
      <DeleteDomainDialog
        domain={mockDomains[0]}
        open={true}
        onOpenChange={vi.fn()}
        onConfirm={onConfirm}
        isPending={false}
      />,
      { wrapper: createWrapper() }
    );

    await user.click(screen.getByTestId("confirm-delete-domain"));

    expect(onConfirm).toHaveBeenCalled();
  });

  it("disables buttons when pending", () => {
    render(
      <DeleteDomainDialog
        domain={mockDomains[0]}
        open={true}
        onOpenChange={vi.fn()}
        onConfirm={vi.fn()}
        isPending={true}
      />,
      { wrapper: createWrapper() }
    );

    expect(screen.getByTestId("confirm-delete-domain")).toBeDisabled();
  });
});
