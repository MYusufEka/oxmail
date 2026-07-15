import { render, screen, fireEvent, waitFor, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";

vi.stubGlobal("fetch", vi.fn());

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
    markAsRead: vi.fn().mockResolvedValue(undefined),
    toggleRead: vi.fn().mockResolvedValue(undefined),
    trashMessage: vi.fn().mockResolvedValue(undefined),
  },
}));

vi.mock("next/navigation", () => ({
  useSearchParams: () => ({ get: () => null }),
  useRouter: () => ({ replace: vi.fn(), push: vi.fn() }),
}));

import SignaturePage, { getSignatureForEmail } from "@/app/mail/signature/page";
import { ComposeDialog } from "@/app/mail/compose-dialog";
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

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
  cleanup();
  localStorage.clear();
});

describe("SignaturePage", () => {
  it("renders page header", () => {
    render(<SignaturePage />);
    expect(screen.getByText("Email Signature")).toBeInTheDocument();
    expect(screen.getByText("Signature settings")).toBeInTheDocument();
  });

  it("renders toggle disabled by default", () => {
    render(<SignaturePage />);
    const toggle = screen.getByTestId("signature-toggle");
    expect(toggle).toBeInTheDocument();
    expect(screen.getByText("Disabled")).toBeInTheDocument();
  });

  it("toggles enabled state", async () => {
    const user = userEvent.setup();
    render(<SignaturePage />);

    const toggle = screen.getByTestId("signature-toggle");
    await user.click(toggle);

    expect(screen.getByText("Enabled")).toBeInTheDocument();
  });

  it("disables textarea when toggle off", () => {
    render(<SignaturePage />);
    const textarea = screen.getByTestId("signature-textarea");
    expect(textarea).toBeDisabled();
  });

  it("enables textarea when toggle on", async () => {
    const user = userEvent.setup();
    render(<SignaturePage />);

    await user.click(screen.getByTestId("signature-toggle"));

    const textarea = screen.getByTestId("signature-textarea");
    expect(textarea).not.toBeDisabled();
  });

  it("saves to localStorage on Save click", async () => {
    const user = userEvent.setup();
    render(<SignaturePage />);

    await user.click(screen.getByTestId("signature-toggle"));
    const textarea = screen.getByTestId("signature-textarea");
    await user.clear(textarea);
    await user.type(textarea, "--\nAlice");

    await user.click(screen.getByTestId("signature-save-btn"));

    const stored = JSON.parse(localStorage.getItem("signature:alice@local.test") ?? "{}");
    expect(stored.enabled).toBe(true);
    expect(stored.content).toBe("--\nAlice");
  });

  it("renders persisted state from localStorage", () => {
    localStorage.setItem(
      "signature:alice@local.test",
      JSON.stringify({ enabled: true, content: "--\nPersisted" }),
    );
    render(<SignaturePage />);
    expect(screen.getByText("Enabled")).toBeInTheDocument();
    expect(screen.getByTestId("signature-textarea")).toHaveValue("--\nPersisted");
  });

  it("renders save button with data-testid", () => {
    render(<SignaturePage />);
    expect(screen.getByTestId("signature-save-btn")).toBeInTheDocument();
  });
});

describe("getSignatureForEmail", () => {
  it("returns null when no signature stored", () => {
    expect(getSignatureForEmail("bob@test.com")).toBeNull();
  });

  it("returns null when signature is disabled", () => {
    localStorage.setItem(
      "signature:bob@test.com",
      JSON.stringify({ enabled: false, content: "--\nBob" }),
    );
    expect(getSignatureForEmail("bob@test.com")).toBeNull();
  });

  it("returns null when content is empty", () => {
    localStorage.setItem(
      "signature:bob@test.com",
      JSON.stringify({ enabled: true, content: "" }),
    );
    expect(getSignatureForEmail("bob@test.com")).toBeNull();
  });

  it("returns signature when enabled with content", () => {
    localStorage.setItem(
      "signature:bob@test.com",
      JSON.stringify({ enabled: true, content: "--\nBob" }),
    );
    const result = getSignatureForEmail("bob@test.com");
    expect(result).toEqual({ enabled: true, content: "--\nBob" });
  });
});

describe("ComposeDialog signature integration", () => {
  it("getSignatureForEmail returns enabled signature for compose", () => {
    localStorage.setItem(
      "signature:test@local.test",
      JSON.stringify({ enabled: true, content: "--\nTest Sig" }),
    );
    const sig = getSignatureForEmail("test@local.test");
    expect(sig).not.toBeNull();
    expect(sig?.enabled).toBe(true);
    expect(sig?.content).toBe("--\nTest Sig");
  });

  it("does not append signature when disabled", () => {
    localStorage.setItem(
      "signature:test@local.test",
      JSON.stringify({ enabled: false, content: "--\nTest Sig" }),
    );
    expect(getSignatureForEmail("test@local.test")).toBeNull();
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
