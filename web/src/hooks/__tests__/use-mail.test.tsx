import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { ReactNode } from "react";
import type {
  InboxResponse,
  FoldersResponse,
  ThreadsResponse,
  MailMessage,
  MailFolder,
  MailThread,
  SendMailResponse,
  MailQueueStatus,
} from "@/types/api";

const {
  mockGetMailFolders,
  mockGetInbox,
  mockGetFolderMessages,
  mockGetThreads,
  mockGetMessage,
  mockSendMail,
  mockGetMailQueue,
  mockCreateFolder,
  mockDeleteFolder,
  mockRenameFolder,
  mockMoveMessage,
} = vi.hoisted(() => ({
  mockGetMailFolders: vi.fn(),
  mockGetInbox: vi.fn(),
  mockGetFolderMessages: vi.fn(),
  mockGetThreads: vi.fn(),
  mockGetMessage: vi.fn(),
  mockSendMail: vi.fn(),
  mockGetMailQueue: vi.fn(),
  mockCreateFolder: vi.fn(),
  mockDeleteFolder: vi.fn(),
  mockRenameFolder: vi.fn(),
  mockMoveMessage: vi.fn(),
}));

vi.mock("@/lib/api-client", () => ({
  apiClient: {
    getMailFolders: mockGetMailFolders,
    getInbox: mockGetInbox,
    getFolderMessages: mockGetFolderMessages,
    getThreads: mockGetThreads,
    getMessage: mockGetMessage,
    sendMail: mockSendMail,
    getMailQueue: mockGetMailQueue,
    createFolder: mockCreateFolder,
    deleteFolder: mockDeleteFolder,
    renameFolder: mockRenameFolder,
    moveMessage: mockMoveMessage,
  },
}));

import {
  useMailFolders,
  useInbox,
  useFolderMessages,
  useThreads,
  useMessage,
  useSendMail,
  useMailQueue,
  useCreateFolder,
  useDeleteFolder,
  useRenameFolder,
  useMoveMessage,
} from "@/hooks/use-mail";

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

const mockFolders: FoldersResponse = {
  folders: [
    { name: "INBOX", delimiter: "/", unread: 2, total: 10 },
    { name: "Sent", delimiter: "/", unread: 0, total: 5 },
  ],
};

const mockMessages: MailMessage[] = [
  {
    id: 1,
    from: "sender@test.com",
    to: ["recipient@test.com"],
    subject: "Hello",
    read: false,
    receivedAt: "2024-01-01T00:00:00Z",
  },
];

const mockInbox: InboxResponse = {
  messages: mockMessages,
  pagination: { page: 1, limit: 20, total: 1 },
};

const mockThreads: ThreadsResponse = {
  threads: [
    {
      threadId: "thread-1",
      subject: "Re: Hello",
      messages: mockMessages,
      lastDate: "2024-01-01T00:00:00Z",
      participantCount: 2,
      unreadCount: 1,
    },
  ],
};

const mockMessage: MailMessage = {
  id: 1,
  from: "sender@test.com",
  to: ["recipient@test.com"],
  subject: "Hello",
  bodyText: "Test body",
  read: true,
  receivedAt: "2024-01-01T00:00:00Z",
};

const mockSendResponse: SendMailResponse = {
  messageId: "msg-123",
  status: "queued",
};

const mockMailQueue: MailQueueStatus = {
  total: 3,
  deferred: 1,
  active: 2,
  oldestAge: "5m",
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe("useMailFolders", () => {
  it("fetches mail folders successfully", async () => {
    mockGetMailFolders.mockResolvedValue(mockFolders);

    const { result } = renderHook(() => useMailFolders("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockFolders);
  });

  it("does not fetch when email is empty", () => {
    const { result } = renderHook(() => useMailFolders(""), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("handles fetch error", async () => {
    mockGetMailFolders.mockRejectedValue(new Error("Failed to fetch folders"));

    const { result } = renderHook(() => useMailFolders("alice@example.com"), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useInbox", () => {
  it("fetches inbox successfully", async () => {
    mockGetInbox.mockResolvedValue(mockInbox);

    const { result } = renderHook(() => useInbox(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockInbox);
  });

  it("does not fetch when userId is 0", () => {
    const { result } = renderHook(() => useInbox(0), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe("idle");
  });

  it("returns empty inbox when no messages exist", async () => {
    const emptyInbox: InboxResponse = {
      messages: [],
      pagination: { page: 1, limit: 20, total: 0 },
    };
    mockGetInbox.mockResolvedValue(emptyInbox);

    const { result } = renderHook(() => useInbox(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.messages).toHaveLength(0);
  });

  it("handles fetch error", async () => {
    mockGetInbox.mockRejectedValue(new Error("Failed to fetch inbox"));

    const { result } = renderHook(() => useInbox(1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useFolderMessages", () => {
  it("fetches folder messages successfully", async () => {
    mockGetFolderMessages.mockResolvedValue(mockInbox);

    const { result } = renderHook(
      () => useFolderMessages("INBOX", "alice@example.com"),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockInbox);
  });

  it("does not fetch when folder or email is empty", () => {
    const { result: r1 } = renderHook(
      () => useFolderMessages("", "alice@example.com"),
      { wrapper: createWrapper() },
    );
    const { result: r2 } = renderHook(
      () => useFolderMessages("INBOX", ""),
      { wrapper: createWrapper() },
    );

    expect(r1.current.fetchStatus).toBe("idle");
    expect(r2.current.fetchStatus).toBe("idle");
  });
});

describe("useThreads", () => {
  it("fetches threads successfully", async () => {
    mockGetThreads.mockResolvedValue(mockThreads);

    const { result } = renderHook(
      () => useThreads("alice@example.com", "INBOX"),
      { wrapper: createWrapper() },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockThreads);
  });

  it("does not fetch when folder or email is empty", () => {
    const { result } = renderHook(
      () => useThreads("", "INBOX"),
      { wrapper: createWrapper() },
    );

    expect(result.current.fetchStatus).toBe("idle");
  });
});

describe("useMessage", () => {
  it("fetches a single message successfully", async () => {
    mockGetMessage.mockResolvedValue(mockMessage);

    const { result } = renderHook(() => useMessage(1, 1), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockMessage);
  });

  it("does not fetch when userId or messageId is 0", () => {
    const { result: r1 } = renderHook(() => useMessage(0, 1), {
      wrapper: createWrapper(),
    });
    const { result: r2 } = renderHook(() => useMessage(1, 0), {
      wrapper: createWrapper(),
    });

    expect(r1.current.fetchStatus).toBe("idle");
    expect(r2.current.fetchStatus).toBe("idle");
  });
});

describe("useSendMail", () => {
  it("sends mail successfully", async () => {
    mockSendMail.mockResolvedValue(mockSendResponse);

    const { result } = renderHook(() => useSendMail(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      from: "alice@example.com",
      to: ["bob@test.com"],
      subject: "Test",
      bodyText: "Hello",
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockSendResponse);
  });

  it("handles send error", async () => {
    mockSendMail.mockRejectedValue(new Error("Send failed"));

    const { result } = renderHook(() => useSendMail(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({
      from: "alice@example.com",
      to: ["bob@test.com"],
      subject: "Test",
      bodyText: "Hello",
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe("useMailQueue", () => {
  it("fetches mail queue status successfully", async () => {
    mockGetMailQueue.mockResolvedValue(mockMailQueue);

    const { result } = renderHook(() => useMailQueue(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(mockMailQueue);
  });
});

describe("useCreateFolder", () => {
  it("creates folder successfully", async () => {
    mockCreateFolder.mockResolvedValue({ status: "created", name: "Custom" });

    const { result } = renderHook(() => useCreateFolder("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate("Custom");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBe("created");
  });
});

describe("useDeleteFolder", () => {
  it("deletes folder successfully", async () => {
    mockDeleteFolder.mockResolvedValue({ status: "deleted" });

    const { result } = renderHook(() => useDeleteFolder("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate("Custom");

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("useRenameFolder", () => {
  it("renames folder successfully", async () => {
    mockRenameFolder.mockResolvedValue({ status: "renamed", name: "NewName" });

    const { result } = renderHook(() => useRenameFolder("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ oldName: "Old", newName: "NewName" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});

describe("useMoveMessage", () => {
  it("moves message successfully", async () => {
    mockMoveMessage.mockResolvedValue({ status: "moved" });

    const { result } = renderHook(() => useMoveMessage("alice@example.com"), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ uid: 1, fromFolder: "INBOX", toFolder: "Archive" });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
  });
});
