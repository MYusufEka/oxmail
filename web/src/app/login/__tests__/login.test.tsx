import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import LoginPage from "../page";
import { AuthProvider } from "@/contexts/auth";

const pushMock = vi.fn();
const replaceMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: pushMock,
    replace: replaceMock,
    prefetch: vi.fn(),
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

function renderLoginPage() {
  return render(
    <AuthProvider>
      <LoginPage />
    </AuthProvider>,
  );
}

function jsonResponse(body: unknown, init?: ResponseInit) {
  return new Response(JSON.stringify(body), {
    status: 200,
    headers: { "Content-Type": "application/json" },
    ...init,
  });
}

describe("LoginPage", () => {
  beforeEach(() => {
    pushMock.mockClear();
    replaceMock.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders login form after unauthenticated session check", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(
        { error: { code: "UNAUTHORIZED", message: "Unauthorized" } },
        { status: 401 },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    renderLoginPage();

    expect(await screen.findByTestId("login-form")).toBeInTheDocument();
    expect(screen.getByTestId("login-email")).toBeInTheDocument();
    expect(screen.getByTestId("login-password")).toBeInTheDocument();
    expect(screen.getByTestId("login-submit")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/auth/me",
      expect.objectContaining({ credentials: "include" }),
    );
  });

  it("successful submit calls login and redirects home", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(
          { error: { code: "UNAUTHORIZED", message: "Unauthorized" } },
          { status: 401 },
        ),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          status: "authenticated",
          email: "admin@local.test",
          mustChangePassword: false,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    renderLoginPage();

    await userEvent.type(await screen.findByTestId("login-email"), "admin@local.test");
    await userEvent.type(screen.getByTestId("login-password"), "secret123");
    await userEvent.click(screen.getByTestId("login-submit"));

    await waitFor(() => expect(pushMock).toHaveBeenCalledWith("/"));
    expect(fetchMock).toHaveBeenLastCalledWith(
      "http://localhost:8080/api/auth/login",
      expect.objectContaining({
        body: JSON.stringify({ email: "admin@local.test", password: "secret123" }),
        credentials: "include",
        method: "POST",
      }),
    );
  });

  it("successful submit redirects forced users to change password", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(
          { error: { code: "UNAUTHORIZED", message: "Unauthorized" } },
          { status: 401 },
        ),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          status: "authenticated",
          email: "admin@local.test",
          mustChangePassword: true,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    renderLoginPage();

    await userEvent.type(await screen.findByTestId("login-email"), "admin@local.test");
    await userEvent.type(screen.getByTestId("login-password"), "secret123");
    await userEvent.click(screen.getByTestId("login-submit"));

    await waitFor(() => expect(pushMock).toHaveBeenCalledWith("/account/change-password"));
  });

  it("invalid credentials show an inline error", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(
          { error: { code: "UNAUTHORIZED", message: "Unauthorized" } },
          { status: 401 },
        ),
      )
      .mockResolvedValueOnce(
        jsonResponse(
          { error: { code: "INVALID_CREDENTIALS", message: "Invalid credentials" } },
          { status: 401 },
        ),
      );
    vi.stubGlobal("fetch", fetchMock);

    renderLoginPage();

    await userEvent.type(await screen.findByTestId("login-email"), "admin@local.test");
    await userEvent.type(screen.getByTestId("login-password"), "wrong-password");
    await userEvent.click(screen.getByTestId("login-submit"));

    expect(await screen.findByTestId("login-error")).toHaveTextContent("Invalid credentials");
    expect(pushMock).not.toHaveBeenCalled();
  });

  it("authenticated session redirects away from login", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        email: "admin@local.test",
        role: "admin",
        mustChangePassword: false,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    renderLoginPage();

    await waitFor(() => expect(replaceMock).toHaveBeenCalledWith("/"));
    expect(screen.queryByTestId("login-form")).not.toBeInTheDocument();
  });
});
