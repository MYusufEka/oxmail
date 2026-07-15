import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import { RecipientInput } from "@/app/mail/recipient-input";
import {
  ComposeDialog,
  loadDraft,
  hasDraftForEmail,
} from "@/app/mail/compose-dialog";

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
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("RecipientInput", () => {
  it("renders with label and placeholder", () => {
    render(
      <RecipientInput
        label="To"
        recipients={[]}
        onChange={() => {}}
        placeholder="email@example.com"
      />,
    );

    expect(screen.getByText("To")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("email@example.com")).toBeInTheDocument();
  });

  it("adds a valid email on Enter", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(
      <RecipientInput label="To" recipients={[]} onChange={onChange} />,
    );

    const input = screen.getByRole("textbox");
    await user.type(input, "test@example.com{Enter}");

    expect(onChange).toHaveBeenCalledWith(["test@example.com"]);
  });

  it("shows error for invalid email", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(
      <RecipientInput label="To" recipients={[]} onChange={onChange} />,
    );

    const input = screen.getByRole("textbox");
    await user.type(input, "invalid-email{Enter}");

    expect(screen.getByRole("alert")).toHaveTextContent("Invalid email");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("removes a recipient when X is clicked", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(
      <RecipientInput
        label="To"
        recipients={["alice@test.com", "bob@test.com"]}
        onChange={onChange}
      />,
    );

    const removeButtons = screen.getAllByRole("button");
    await user.click(removeButtons[0]);

    expect(onChange).toHaveBeenCalledWith(["bob@test.com"]);
  });

  it("prevents duplicate emails", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    render(
      <RecipientInput
        label="To"
        recipients={["test@example.com"]}
        onChange={onChange}
      />,
    );

    const input = screen.getByRole("textbox");
    await user.type(input, "test@example.com{Enter}");

    expect(screen.getByRole("alert")).toHaveTextContent("Already added");
    expect(onChange).not.toHaveBeenCalled();
  });
});

describe("ComposeDialog", () => {
  it("renders when open", () => {
    const Wrapper = createWrapper();

    render(
      <Wrapper>
        <ComposeDialog open={true} onOpenChange={() => {}} currentUserEmail="test@local.test" />
      </Wrapper>,
    );

    expect(screen.getByText("New Message")).toBeInTheDocument();
    expect(screen.getByText("Send")).toBeInTheDocument();
    expect(screen.getByText("Add Cc/Bcc")).toBeInTheDocument();
  });

  it("shows Cc/Bcc fields when link is clicked", async () => {
    const Wrapper = createWrapper();
    const user = userEvent.setup();

    render(
      <Wrapper>
        <ComposeDialog open={true} onOpenChange={() => {}} currentUserEmail="test@local.test" />
      </Wrapper>,
    );

    await user.click(screen.getByText("Add Cc/Bcc"));

    expect(screen.getByText("Cc")).toBeInTheDocument();
    expect(screen.getByText("Bcc")).toBeInTheDocument();
  });

  it("renders with initial To and Subject for reply", () => {
    const Wrapper = createWrapper();

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={() => {}}
          initialTo={["sender@test.com"]}
          initialSubject="Re: Hello"
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    expect(screen.getByText("sender@test.com")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Re: Hello")).toBeInTheDocument();
  });

  it("shows discard confirmation when closing with content", async () => {
    const Wrapper = createWrapper();
    const user = userEvent.setup();

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={() => {}}
          initialTo={["someone@test.com"]}
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await user.click(screen.getByLabelText("Close compose"));

    expect(screen.getByText("Save as draft?")).toBeInTheDocument();
  });

  it("sends mail on Send click", async () => {
    const Wrapper = createWrapper();
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ messageId: "msg-1", status: "queued" }),
    });

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={onOpenChange}
          initialTo={["recipient@test.com"]}
          initialSubject="Test Subject"
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await user.click(screen.getByText("Send"));

    await waitFor(() => {
      expect(globalThis.fetch).toHaveBeenCalled();
    });
  });

  it("shows attachment filename after file selection", async () => {
    const Wrapper = createWrapper();

    render(
      <Wrapper>
        <ComposeDialog open={true} onOpenChange={() => {}} currentUserEmail="test@local.test" />
      </Wrapper>,
    );

    const fileInput = document.querySelector(
      'input[type="file"]',
    ) as HTMLInputElement;

    const file = new File(["content"], "document.pdf", {
      type: "application/pdf",
    });

    fireEvent.change(fileInput, { target: { files: [file] } });

    expect(screen.getByText("document.pdf")).toBeInTheDocument();
  });
});

describe("ComposeDialog — draft persistence", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("auto-saves draft to localStorage after 2s debounce", async () => {
    vi.useFakeTimers();
    const Wrapper = createWrapper();

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={() => {}}
          initialTo={["alice@test.com"]}
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await vi.advanceTimersByTimeAsync(1900);
    expect(localStorage.getItem("draft:test@local.test")).toBeNull();

    await vi.advanceTimersByTimeAsync(200);
    const stored = loadDraft("test@local.test");
    expect(stored).not.toBeNull();
    expect(stored?.to).toEqual(["alice@test.com"]);
    expect(stored?.savedAt).toBeTruthy();
    vi.useRealTimers();
  });

  it("restores draft on open and shows restored subject", async () => {
    const savedAt = new Date().toISOString();
    localStorage.setItem(
      "draft:test@local.test",
      JSON.stringify({ to: ["bob@test.com"], subject: "Saved subject", body: "", savedAt }),
    );

    const Wrapper = createWrapper();

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={() => {}}
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await waitFor(() => {
      expect(screen.getByDisplayValue("Saved subject")).toBeInTheDocument();
    });
    expect(screen.getByText("bob@test.com")).toBeInTheDocument();
  });

  it("clears draft on send success", async () => {
    const Wrapper = createWrapper();
    const user = userEvent.setup();
    const onOpenChange = vi.fn();

    localStorage.setItem(
      "draft:test@local.test",
      JSON.stringify({ to: ["bob@test.com"], subject: "Draft", body: "", savedAt: new Date().toISOString() }),
    );

    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ messageId: "msg-1", status: "queued" }),
    });

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={onOpenChange}
          initialTo={["bob@test.com"]}
          initialSubject="Draft"
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await user.click(screen.getByText("Send"));

    await waitFor(() => {
      expect(localStorage.getItem("draft:test@local.test")).toBeNull();
    });
  });

  it("shows save-draft dialog on close with content", async () => {
    const Wrapper = createWrapper();
    const user = userEvent.setup();

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={() => {}}
          initialTo={["alice@test.com"]}
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await user.click(screen.getByLabelText("Close compose"));

    expect(screen.getByText("Save as draft?")).toBeInTheDocument();
    expect(screen.getByTestId("save-draft-btn")).toBeInTheDocument();
    expect(screen.getByTestId("discard-draft-btn")).toBeInTheDocument();
  });

  it("discard button clears draft and closes", async () => {
    vi.useFakeTimers();
    const Wrapper = createWrapper();
    const onOpenChange = vi.fn();

    render(
      <Wrapper>
        <ComposeDialog
          open={true}
          onOpenChange={onOpenChange}
          initialTo={["alice@test.com"]}
          currentUserEmail="test@local.test"
        />
      </Wrapper>,
    );

    await vi.advanceTimersByTimeAsync(2100);
    expect(hasDraftForEmail("test@local.test")).toBe(true);

    vi.useRealTimers();

    fireEvent.click(screen.getByLabelText("Close compose"));
    fireEvent.click(screen.getByTestId("discard-draft-btn"));

    expect(hasDraftForEmail("test@local.test")).toBe(false);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
