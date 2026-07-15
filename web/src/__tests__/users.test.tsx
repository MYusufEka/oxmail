import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import type { PaginatedResponse, Domain, User } from "@/types/api";
import UsersPage from "@/app/users/page";
import { UserTable } from "@/app/users/user-table";
import { DeleteUserDialog } from "@/app/users/delete-user-dialog";

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
    { id: 2, name: "test.org", active: true, createdAt: "2024-01-02T00:00:00Z", updatedAt: "2024-01-02T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 2 },
};

const mockUsers: PaginatedResponse<User> = {
  data: [
    { id: 1, email: "admin@example.com", domainId: 1, quota: 1073741824, active: true, createdAt: "2024-01-01T00:00:00Z", updatedAt: "2024-01-01T00:00:00Z" },
    { id: 2, email: "user@example.com", domainId: 1, quota: 0, active: false, createdAt: "2024-02-01T00:00:00Z", updatedAt: "2024-02-01T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 2 },
};

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.restoreAllMocks();
});

function mockFetchResponses(responses: unknown[]) {
  const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
  for (const response of responses) {
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(response),
    });
  }
}

describe("UserTable", () => {
  it("renders user rows with correct data", () => {
    const onDelete = vi.fn();
    const onEdit = vi.fn();
    const onMailSetup = vi.fn();
    const onResetPassword = vi.fn();
    const onBulkDelete = vi.fn().mockResolvedValue(undefined);
    render(
      <UserTable users={mockUsers.data} isLoading={false} onDelete={onDelete} onEdit={onEdit} onMailSetup={onMailSetup} onResetPassword={onResetPassword} onBulkDelete={onBulkDelete} />
    );

    expect(screen.getByText("admin@example.com")).toBeInTheDocument();
    expect(screen.getByText("user@example.com")).toBeInTheDocument();
    expect(screen.getByText("1.0 GB")).toBeInTheDocument();
    expect(screen.getByText("Unlimited")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });

  it("renders loading skeleton when isLoading is true", () => {
    const onDelete = vi.fn();
    const onEdit = vi.fn();
    const onMailSetup = vi.fn();
    const onResetPassword = vi.fn();
    const onBulkDelete = vi.fn().mockResolvedValue(undefined);
    const { container } = render(
      <UserTable users={[]} isLoading={true} onDelete={onDelete} onEdit={onEdit} onMailSetup={onMailSetup} onResetPassword={onResetPassword} onBulkDelete={onBulkDelete} />
    );

    const skeletons = container.querySelectorAll("[data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders empty state when no users", () => {
    const onDelete = vi.fn();
    const onEdit = vi.fn();
    const onMailSetup = vi.fn();
    const onResetPassword = vi.fn();
    const onBulkDelete = vi.fn().mockResolvedValue(undefined);
    render(
      <UserTable users={[]} isLoading={false} onDelete={onDelete} onEdit={onEdit} onMailSetup={onMailSetup} onResetPassword={onResetPassword} onBulkDelete={onBulkDelete} />
    );

    expect(screen.getByText(/No users found/)).toBeInTheDocument();
  });

  it("calls onDelete when delete button is clicked", () => {
    const onDelete = vi.fn();
    const onEdit = vi.fn();
    const onMailSetup = vi.fn();
    const onResetPassword = vi.fn();
    const onBulkDelete = vi.fn().mockResolvedValue(undefined);
    render(
      <UserTable users={mockUsers.data} isLoading={false} onDelete={onDelete} onEdit={onEdit} onMailSetup={onMailSetup} onResetPassword={onResetPassword} onBulkDelete={onBulkDelete} />
    );

    const deleteButtons = screen.getAllByRole("button", { name: /delete user/i });
    fireEvent.click(deleteButtons[0]);

    expect(onDelete).toHaveBeenCalledWith(mockUsers.data[0]);
  });
});

describe("DeleteUserDialog", () => {
  it("shows user email in confirmation message", () => {
    const user = mockUsers.data[0];
    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })}>
        <DeleteUserDialog user={user} open={true} onOpenChange={vi.fn()} />
      </QueryClientProvider>
    );

    expect(screen.getByText("admin@example.com")).toBeInTheDocument();
    expect(screen.getByText(/cannot be undone/)).toBeInTheDocument();
  });

  it("calls onOpenChange(false) when cancel is clicked", () => {
    const onOpenChange = vi.fn();
    const user = mockUsers.data[0];
    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })}>
        <DeleteUserDialog user={user} open={true} onOpenChange={onOpenChange} />
      </QueryClientProvider>
    );

    fireEvent.click(screen.getByRole("button", { name: /cancel/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});

describe("UsersPage", () => {
  it("renders loading state initially", () => {
    mockFetchResponses([]);
    const { container } = render(<UsersPage />, { wrapper: createWrapper() });

    const skeletons = container.querySelectorAll("[data-slot='skeleton']");
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it("renders domain selector and users after loading", async () => {
    mockFetchResponses([mockDomains, mockUsers]);

    render(<UsersPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("admin@example.com")).toBeInTheDocument();
    });

    expect(screen.getByText("user@example.com")).toBeInTheDocument();
  });

  it("shows empty domain state when no domains exist", async () => {
    mockFetchResponses([{ data: [], pagination: { page: 1, limit: 20, total: 0 } }]);

    render(<UsersPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText(/No domains configured/)).toBeInTheDocument();
    });
  });

  it("shows error state when user fetch fails", async () => {
    const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
    // First call: domains success
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockDomains),
    });
    // Second call: users error
    fetchMock.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ error: { code: "INTERNAL_ERROR", message: "Server error" } }),
    });

    render(<UsersPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText(/Failed to load users/)).toBeInTheDocument();
    });
  });
});
