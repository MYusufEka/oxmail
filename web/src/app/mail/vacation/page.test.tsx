import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";

const authEmail = "auth-user@local.test";

const { mockGetVacationScript, mockSetVacation, mockDeleteVacationScript } = vi.hoisted(() => ({
  mockGetVacationScript: vi.fn(),
  mockSetVacation: vi.fn(),
  mockDeleteVacationScript: vi.fn(),
}));

vi.mock("@/contexts/auth", () => ({
  useAuth: () => ({
    user: { email: authEmail, role: "user" },
    email: authEmail,
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    refresh: vi.fn(),
  }),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getVacationScript: mockGetVacationScript,
    setVacation: mockSetVacation,
    deleteVacationScript: mockDeleteVacationScript,
  },
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

import VacationPage from "@/app/mail/vacation/page";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

describe("VacationPage auth identity", () => {
  beforeEach(() => {
    mockGetVacationScript.mockResolvedValue({ email: authEmail, active: false, script: "" });
    mockSetVacation.mockResolvedValue({ email: authEmail, active: true, script: "" });
    mockDeleteVacationScript.mockResolvedValue({ email: authEmail, active: false });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("loads vacation script with authenticated email", async () => {
    render(<VacationPage />, { wrapper: createWrapper() });

    await waitFor(() => expect(mockGetVacationScript).toHaveBeenCalledWith(authEmail));
    expect(mockGetVacationScript).not.toHaveBeenCalledWith("alice@local.test");
  });

  it("saves structured vacation settings with authenticated email", async () => {
    render(<VacationPage />, { wrapper: createWrapper() });

    await waitFor(() => expect(screen.getByTestId("save-vacation")).toBeInTheDocument());
    fireEvent.click(screen.getByRole("switch", { name: "Toggle vacation auto-reply" }));
    fireEvent.change(screen.getByTestId("vacation-subject"), { target: { value: "Away" } });
    fireEvent.change(screen.getByTestId("vacation-body"), { target: { value: "Back soon" } });
    fireEvent.click(screen.getByTestId("save-vacation"));

    await waitFor(() =>
      expect(mockSetVacation).toHaveBeenCalledWith(authEmail, {
        subject: "Away",
        body: "Back soon",
        enabled: true,
      }),
    );
    expect(mockSetVacation).not.toHaveBeenCalledWith("alice@local.test", expect.anything());
  });
});
