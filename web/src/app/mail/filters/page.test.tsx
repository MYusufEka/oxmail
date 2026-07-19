import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

const authEmail = "auth-user@local.test";

const { mockUseSieveScript, mockUseMailFolders, mockMutateAsync } = vi.hoisted(() => ({
  mockUseSieveScript: vi.fn(),
  mockUseMailFolders: vi.fn(),
  mockMutateAsync: vi.fn(),
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

vi.mock("@/hooks/use-sieve", () => ({
  useSieveScript: mockUseSieveScript,
  useSetSieveScript: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
}));

vi.mock("@/hooks/use-mail", () => ({
  useMailFolders: mockUseMailFolders,
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
}));

import FiltersPage from "@/app/mail/filters/page";

describe("FiltersPage auth identity", () => {
  beforeEach(() => {
    mockUseSieveScript.mockReturnValue({
      data: { email: authEmail, active: false, script: "" },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mockUseMailFolders.mockReturnValue({
      data: { folders: [{ name: "INBOX", delimiter: "/", unread: 0, total: 0 }] },
      isLoading: false,
    });
    mockMutateAsync.mockResolvedValue({ email: authEmail, active: true, script: "" });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("loads filters and folders with authenticated email", () => {
    render(<FiltersPage />);

    expect(mockUseSieveScript).toHaveBeenCalledWith(authEmail);
    expect(mockUseMailFolders).toHaveBeenCalledWith(authEmail);
    expect(mockUseSieveScript).not.toHaveBeenCalledWith("alice@local.test");
    expect(mockUseMailFolders).not.toHaveBeenCalledWith("alice@local.test");
  });

  it("saves filter script with authenticated email", async () => {
    render(<FiltersPage />);

    fireEvent.click(screen.getByTestId("save-filters"));

    await waitFor(() =>
      expect(mockMutateAsync).toHaveBeenCalledWith({
        email: authEmail,
        script: "",
      }),
    );
    expect(mockMutateAsync).not.toHaveBeenCalledWith(
      expect.objectContaining({ email: "alice@local.test" }),
    );
  });
});
