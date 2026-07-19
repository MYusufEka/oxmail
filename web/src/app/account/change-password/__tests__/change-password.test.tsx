import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import ChangePasswordPage from "../page";

const mocks = vi.hoisted(() => ({
  push: vi.fn(),
  refresh: vi.fn(),
  changePassword: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: mocks.push,
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

vi.mock("@/contexts/auth", () => ({
  useAuth: () => ({
    email: "alice@local.test",
    refresh: mocks.refresh,
  }),
}));

vi.mock("@/lib/api-client", () => ({
  ApiError: class ApiError extends Error {
    constructor(
      public readonly status: number,
      public readonly code: string,
      message: string,
    ) {
      super(message);
    }
  },
  apiClient: {
    changePassword: mocks.changePassword,
  },
}));

describe("ChangePasswordPage", () => {
  beforeEach(() => {
    mocks.push.mockClear();
    mocks.refresh.mockClear();
    mocks.changePassword.mockReset();
    mocks.changePassword.mockResolvedValue({ status: "password_changed" });
    mocks.refresh.mockResolvedValue(undefined);
  });

  it("prefills authenticated email and refreshes session after password change", async () => {
    render(<ChangePasswordPage />);

    expect(await screen.findByTestId("email-input")).toHaveValue("alice@local.test");

    await userEvent.type(screen.getByTestId("current-password-input"), "OldPass123!");
    await userEvent.type(screen.getByTestId("new-password-input"), "NewPass456!");
    await userEvent.type(screen.getByTestId("confirm-password-input"), "NewPass456!");
    await userEvent.click(screen.getByTestId("change-password-submit"));

    await waitFor(() => {
      expect(mocks.changePassword).toHaveBeenCalledWith({
        email: "alice@local.test",
        currentPassword: "OldPass123!",
        newPassword: "NewPass456!",
      });
    });
    expect(mocks.refresh).toHaveBeenCalled();
    expect(mocks.push).toHaveBeenCalledWith("/");
  });
});
