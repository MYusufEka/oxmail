import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { MailMessage, InboxResponse, MailFolder, FoldersResponse } from "@/types/api";

const mockMessages: MailMessage[] = [
  {
    id: 1,
    from: "Alice Smith <alice@local.test>",
    to: ["bob@local.test"],
    subject: "Hello World",
    bodyText: "Hi Bob, how are you?",
    bodyHtml: "<p>Hi Bob, how are you?</p>",
    read: false,
    receivedAt: "2024-01-15T10:30:00.000Z",
  },
  {
    id: 2,
    from: "Charlie <charlie@local.test>",
    to: ["bob@local.test"],
    subject: "Re: Hello World",
    bodyText: "I'm good thanks!",
    read: true,
    receivedAt: "2024-01-15T11:00:00.000Z",
  },
  {
    id: 3,
    from: "Dave <dave@local.test>",
    to: ["bob@local.test"],
    cc: ["alice@local.test"],
    subject: "Meeting Tomorrow",
    bodyText: "Let's meet at 3pm.",
    bodyHtml: "<p>Let's meet at 3pm.</p>",
    read: false,
    receivedAt: "2024-01-15T12:00:00.000Z",
  },
];

const mockFolders: MailFolder[] = [
  { name: "INBOX", delimiter: "/", unread: 2, total: 3 },
  { name: "Sent", delimiter: "/", unread: 0, total: 0 },
  { name: "Drafts", delimiter: "/", unread: 0, total: 0 },
  { name: "Trash", delimiter: "/", unread: 0, total: 0 },
];

let mockInboxData: InboxResponse | undefined;
let mockInboxLoading = false;
let mockMessageData: MailMessage | undefined;
let mockMessageLoading = false;
let mockFolderData: InboxResponse | undefined;
let mockFolderLoading = false;

vi.mock("@/hooks/use-mail", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-mail")>();
  return {
    ...actual,
    useMailFolders: () => ({
      data: { folders: mockFolders } as FoldersResponse,
      isLoading: false,
    }),
    useInbox: () => ({
      data: mockInboxData,
      isLoading: mockInboxLoading,
    }),
    useFolderMessages: () => ({
      data: mockFolderData ?? mockInboxData,
      isLoading: mockFolderLoading,
    }),
    useMessage: () => ({
      data: mockMessageData,
      isLoading: mockMessageLoading,
    }),
    useSendMail: () => ({ mutate: vi.fn() }),
    useCreateFolder: () => ({ mutate: vi.fn(), isPending: false }),
    useDeleteFolder: () => ({ mutate: vi.fn(), isPending: false }),
    useRenameFolder: () => ({ mutate: vi.fn(), isPending: false }),
    useMoveMessage: () => ({ mutate: vi.fn(), isPending: false }),
  };
});

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

import { MessageRow } from "@/app/mail/message-row";
import { MessageList } from "@/app/mail/message-list";
import { MessagePreview } from "@/app/mail/message-preview";
import WebmailPage from "@/app/mail/page";

describe("MessageRow", () => {
  const defaultProps = {
    message: mockMessages[0],
    selected: false,
    onSelect: vi.fn(),
  };

  it("renders with data-testid", () => {
    render(<MessageRow {...defaultProps} />);
    expect(screen.getByTestId("message-row")).toBeInTheDocument();
  });

  it("displays sender name extracted from email", () => {
    render(<MessageRow {...defaultProps} />);
    expect(screen.getByText("Alice Smith")).toBeInTheDocument();
  });

  it("displays subject", () => {
    render(<MessageRow {...defaultProps} />);
    expect(screen.getByText("Hello World")).toBeInTheDocument();
  });

  it("shows unread indicator for unread messages", () => {
    const { container } = render(<MessageRow {...defaultProps} />);
    const dot = container.querySelector(".bg-primary");
    expect(dot).toBeInTheDocument();
  });

  it("hides unread indicator for read messages", () => {
    render(
      <MessageRow {...defaultProps} message={mockMessages[1]} />,
    );
    const { container } = render(
      <MessageRow {...defaultProps} message={mockMessages[1]} />,
    );
    const dot = container.querySelector(".bg-primary");
    expect(dot).not.toBeInTheDocument();
  });

  it("calls onSelect when clicked", () => {
    const onSelect = vi.fn();
    render(<MessageRow {...defaultProps} onSelect={onSelect} />);
    fireEvent.click(screen.getAllByTestId("message-row")[0]);
    expect(onSelect).toHaveBeenCalledWith(1);
  });

  it("applies selected styling when selected", () => {
    render(<MessageRow {...defaultProps} selected={true} />);
    const row = screen.getByTestId("message-row");
    expect(row.className).toContain("bg-accent");
  });

  it("renders indented when indented prop is true", () => {
    render(<MessageRow {...defaultProps} indented={true} />);
    const row = screen.getByTestId("message-row");
    expect(row.className).toContain("pl-8");
  });
});

describe("MessageList", () => {
  it("shows loading skeletons when loading", () => {
    render(
      <MessageList
        messages={[]}
        selectedId={null}
        isLoading={true}
        onSelect={vi.fn()}
      />,
    );
    expect(screen.getByTestId("message-list-loading")).toBeInTheDocument();
  });

  it("shows empty state when no messages", () => {
    render(
      <MessageList
        messages={[]}
        selectedId={null}
        isLoading={false}
        onSelect={vi.fn()}
      />,
    );
    expect(screen.getByTestId("inbox-empty")).toBeInTheDocument();
    expect(screen.getByText("Your inbox is empty")).toBeInTheDocument();
  });

  it("renders message rows", () => {
    render(
      <MessageList
        messages={mockMessages}
        selectedId={null}
        isLoading={false}
        onSelect={vi.fn()}
      />,
    );
    expect(screen.getByTestId("message-list")).toBeInTheDocument();
    const rows = screen.getAllByTestId("message-row");
    expect(rows).toHaveLength(3);
  });

  it("groups thread messages with indentation", () => {
    render(
      <MessageList
        messages={mockMessages}
        selectedId={null}
        isLoading={false}
        onSelect={vi.fn()}
      />,
    );
    const rows = screen.getAllByTestId("message-row");
    // "Re: Hello World" is a reply to "Hello World" — should be indented
    expect(rows[1].className).toContain("pl-8");
  });
});

describe("MessagePreview", () => {
  it("shows empty state when no message selected", () => {
    render(<MessagePreview message={null} isLoading={false} />);
    expect(screen.getByTestId("message-preview-empty")).toBeInTheDocument();
    expect(screen.getByText("Select a message to read")).toBeInTheDocument();
  });

  it("shows loading spinner when loading", () => {
    render(<MessagePreview message={null} isLoading={true} />);
    expect(screen.getByTestId("message-preview-loading")).toBeInTheDocument();
  });

  it("renders message subject and sender", () => {
    render(<MessagePreview message={mockMessages[0]} isLoading={false} />);
    expect(screen.getByTestId("message-preview")).toBeInTheDocument();
    expect(screen.getByText("Hello World")).toBeInTheDocument();
    expect(
      screen.getByText("Alice Smith <alice@local.test>"),
    ).toBeInTheDocument();
  });

  it("renders HTML body when available", () => {
    render(<MessagePreview message={mockMessages[0]} isLoading={false} />);
    expect(screen.getByTestId("message-body-html")).toBeInTheDocument();
  });

  it("renders plain text body as fallback", () => {
    render(<MessagePreview message={mockMessages[1]} isLoading={false} />);
    expect(screen.getByTestId("message-body-text")).toBeInTheDocument();
    expect(screen.getByText("I'm good thanks!")).toBeInTheDocument();
  });

  it("shows CC recipients when present", () => {
    render(<MessagePreview message={mockMessages[2]} isLoading={false} />);
    expect(screen.getByText("CC: alice@local.test")).toBeInTheDocument();
  });
});

describe("WebmailPage", () => {
  beforeEach(() => {
    mockInboxData = {
      messages: mockMessages,
      pagination: { page: 1, limit: 50, total: 3 },
    };
    mockInboxLoading = false;
    mockFolderData = {
      messages: mockMessages,
      pagination: { page: 1, limit: 50, total: 3 },
    };
    mockFolderLoading = false;
    mockMessageData = undefined;
    mockMessageLoading = false;
  });

  const renderWithProvider = (ui: React.ReactNode) => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
      <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>,
    );
  };

  it("renders webmail header", () => {
    renderWithProvider(<WebmailPage />);
    expect(screen.getByText("Webmail")).toBeInTheDocument();
  });

  it("renders inbox folder in sidebar", () => {
    renderWithProvider(<WebmailPage />);
    expect(screen.getByText("INBOX")).toBeInTheDocument();
  });

  it("renders additional folders in sidebar", () => {
    renderWithProvider(<WebmailPage />);
    expect(screen.getByText("Sent")).toBeInTheDocument();
    expect(screen.getByText("Drafts")).toBeInTheDocument();
    expect(screen.getByText("Trash")).toBeInTheDocument();
  });

  it("renders message list", () => {
    renderWithProvider(<WebmailPage />);
    expect(screen.getByTestId("message-list")).toBeInTheDocument();
  });

  it("shows unread count in sidebar", () => {
    renderWithProvider(<WebmailPage />);
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("shows empty preview when no message selected", () => {
    renderWithProvider(<WebmailPage />);
    expect(screen.getByTestId("message-preview-empty")).toBeInTheDocument();
  });

  it("selects message on click", () => {
    mockMessageData = mockMessages[0];
    renderWithProvider(<WebmailPage />);
    const rows = screen.getAllByTestId("message-row");
    fireEvent.click(rows[0]);
    expect(rows[0].getAttribute("aria-selected")).toBe("true");
  });

  it("navigates messages with j/k keys", () => {
    renderWithProvider(<WebmailPage />);
    fireEvent.keyDown(document, { key: "j" });
    const rows = screen.getAllByTestId("message-row");
    expect(rows[0].getAttribute("aria-selected")).toBe("true");
  });
});
