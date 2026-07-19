import { render, screen, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";

const signatureApiMocks = vi.hoisted(() => ({
  getSignature: vi.fn(),
  upsertSignature: vi.fn(),
  deleteSignature: vi.fn(),
}));

vi.stubGlobal("fetch", vi.fn());

vi.mock("@/contexts/auth", () => ({
  useAuth: () => ({
    user: { email: "alice@local.test" },
    email: "alice@local.test",
    isAuthenticated: true,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    refresh: vi.fn(),
  }),
}));

vi.mock("@/hooks/use-mail", () => ({
  useMailFolders: () => ({ data: { folders: [] }, isLoading: false }),
  useFolderMessages: () => ({ data: { messages: [] }, isLoading: false }),
  useMessage: () => ({ data: undefined, isLoading: false }),
  useSendMail: () => ({ mutate: vi.fn(), isPending: false }),
  useCreateFolder: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteFolder: () => ({ mutate: vi.fn(), isPending: false }),
  useRenameFolder: () => ({ mutate: vi.fn(), isPending: false }),
  useMoveMessage: () => ({ mutate: vi.fn(), isPending: false }),
  useThreads: () => ({ data: { threads: [] }, isLoading: false }),
  hasDraftForEmail: () => false,
}));

vi.mock("@/hooks/use-domains", () => ({
  useDomains: () => ({
    data: { data: [{ id: 1, name: "local.test", active: true, createdAt: "", updatedAt: "" }], pagination: { page: 1, limit: 100, total: 1 } },
  }),
}));

vi.mock("@/hooks/use-users", () => ({
  useUsers: () => ({
    data: { data: [{ id: 1, email: "alice@local.test", domainId: 1, displayName: "Alice", quota: 1024, active: true, createdAt: "", updatedAt: "" }], pagination: { page: 1, limit: 50, total: 1 } },
  }),
}));

vi.mock("@/hooks/use-contacts", () => ({
  useContacts: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getSignature: signatureApiMocks.getSignature,
    upsertSignature: signatureApiMocks.upsertSignature,
    deleteSignature: signatureApiMocks.deleteSignature,
    markAsRead: vi.fn().mockResolvedValue(undefined),
    toggleRead: vi.fn().mockResolvedValue(undefined),
    trashMessage: vi.fn().mockResolvedValue(undefined),
  },
}));

vi.mock("next/navigation", () => ({
  redirect: vi.fn(),
  useSearchParams: () => ({ get: () => null }),
  useRouter: () => ({ replace: vi.fn(), push: vi.fn() }),
}));

import SignaturePage from "@/app/mail/signature/page";
import WebmailPage from "@/app/mail/page";

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

function renderSignaturePage() {
  const Wrapper = createWrapper();
  return render(
    <Wrapper>
      <SignaturePage />
    </Wrapper>,
  );
}

beforeEach(() => {
  signatureApiMocks.getSignature.mockResolvedValue({
    email: "alice@local.test",
    content: "",
    enabled: false,
  });
  signatureApiMocks.upsertSignature.mockResolvedValue({
    email: "alice@local.test",
    content: "--\nAlice",
    enabled: true,
  });
  signatureApiMocks.deleteSignature.mockResolvedValue(undefined);
});

afterEach(() => {
  vi.clearAllMocks();
  cleanup();
});

describe("SignaturePage", () => {
  it("renders page header after API load", async () => {
    renderSignaturePage();

    expect(await screen.findByText("Signature settings")).toBeInTheDocument();
    expect(screen.getByText("Email Signature")).toBeInTheDocument();
    expect(signatureApiMocks.getSignature).toHaveBeenCalledWith("alice@local.test");
  });

  it("renders toggle disabled from API response", async () => {
    renderSignaturePage();

    const toggle = await screen.findByTestId("signature-toggle");
    expect(toggle).toBeInTheDocument();
    expect(screen.getByText("Disabled")).toBeInTheDocument();
  });

  it("toggles enabled state", async () => {
    const user = userEvent.setup();
    renderSignaturePage();

    const toggle = await screen.findByTestId("signature-toggle");
    await user.click(toggle);

    expect(screen.getByText("Enabled")).toBeInTheDocument();
  });

  it("disables textarea when toggle off", async () => {
    renderSignaturePage();

    const textarea = await screen.findByTestId("signature-textarea");
    expect(textarea).toBeDisabled();
  });

  it("enables textarea when toggle on", async () => {
    const user = userEvent.setup();
    renderSignaturePage();

    await user.click(await screen.findByTestId("signature-toggle"));

    const textarea = screen.getByTestId("signature-textarea");
    expect(textarea).not.toBeDisabled();
  });

  it("saves signature through API on Save click", async () => {
    const user = userEvent.setup();
    renderSignaturePage();

    await user.click(await screen.findByTestId("signature-toggle"));
    const textarea = screen.getByTestId("signature-textarea");
    await user.clear(textarea);
    await user.type(textarea, "--\nAlice");

    await user.click(screen.getByTestId("signature-save-btn"));

    await waitFor(() => {
      expect(signatureApiMocks.upsertSignature).toHaveBeenCalledWith("alice@local.test", {
        enabled: true,
        content: "--\nAlice",
      });
    });
  });

  it("renders persisted state from API", async () => {
    signatureApiMocks.getSignature.mockResolvedValue({
      email: "alice@local.test",
      content: "--\nPersisted",
      enabled: true,
    });

    renderSignaturePage();

    expect(await screen.findByText("Enabled")).toBeInTheDocument();
    expect(screen.getByTestId("signature-textarea")).toHaveValue("--\nPersisted");
  });

  it("deletes signature through API when saved disabled", async () => {
    signatureApiMocks.getSignature.mockResolvedValue({
      email: "alice@local.test",
      content: "--\nPersisted",
      enabled: true,
    });
    const user = userEvent.setup();
    renderSignaturePage();

    await user.click(await screen.findByTestId("signature-toggle"));
    await user.click(screen.getByTestId("signature-save-btn"));

    await waitFor(() => {
      expect(signatureApiMocks.deleteSignature).toHaveBeenCalledWith("alice@local.test");
    });
  });

  it("renders save button with data-testid", async () => {
    renderSignaturePage();
    expect(await screen.findByTestId("signature-save-btn")).toBeInTheDocument();
  });
});

describe("WebmailPage signature sidebar link", () => {
  it("renders Signature link in sidebar", () => {
    const Wrapper = createWrapper();
    render(
      <Wrapper>
        <WebmailPage />
      </Wrapper>,
    );
    const signatureLink = document.querySelector('a[href="/mail/signature"]');
    expect(signatureLink).not.toBeNull();
  });
});
