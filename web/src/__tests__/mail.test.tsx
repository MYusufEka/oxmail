import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { MailMessage, PaginatedResponse } from "@/types/api";

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

let mockInboxData: PaginatedResponse<MailMessage> | undefined;
let mockInboxLoading = false;
let mockMessageData: MailMessage | undefined;
let mockMessageLoading = false;

vi.mock("@/hooks/use-mail", () => ({
  useInbox: () => ({
    data: mockInboxData,
    isLoading: mockInboxLoading,
  }),
  useMessage: () => ({
    data: mockMessageData,
    isLoading: mockMessageLoading,
  }),
  useSendMail: () => ({
    mutate: vi.fn(),
  }),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    markAsRead: vi.fn().mockResolvedValue(undefined),
    toggleRead: vi.fn().mockResolvedValue(undefined),
    trashMessage: vi.fn().mockResolvedValue(undefined),
  },
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
      data: mockMessages,
      pagination: { page: 1, limit: 50, total: 3 },
    };
    mockInboxLoading = false;
    mockMessageData = undefined;
    mockMessageLoading = false;
  });

  it("renders webmail header", () => {
    render(<WebmailPage />);
    expect(screen.getByText("Webmail")).toBeInTheDocument();
  });

  it("renders inbox folder in sidebar", () => {
    render(<WebmailPage />);
    expect(screen.getByText("Inbox")).toBeInTheDocument();
  });

  it("renders message list", () => {
    render(<WebmailPage />);
    expect(screen.getByTestId("message-list")).toBeInTheDocument();
  });

  it("shows unread count in sidebar", () => {
    render(<WebmailPage />);
    // 2 unread messages (id 1 and 3)
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("shows empty preview when no message selected", () => {
    render(<WebmailPage />);
    expect(screen.getByTestId("message-preview-empty")).toBeInTheDocument();
  });

  it("selects message on click", () => {
    mockMessageData = mockMessages[0];
    render(<WebmailPage />);
    const rows = screen.getAllByTestId("message-row");
    fireEvent.click(rows[0]);
    // After click, the preview should show (re-render needed in real app)
    // We verify the row gets selected styling
    expect(rows[0].getAttribute("aria-selected")).toBe("true");
  });

  it("navigates messages with j/k keys", () => {
    render(<WebmailPage />);
    // Press j to select first message
    fireEvent.keyDown(document, { key: "j" });
    const rows = screen.getAllByTestId("message-row");
    // First message should be selected after j
    expect(rows[0].getAttribute("aria-selected")).toBe("true");
  });
});
