import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import { RecipientInput } from "@/app/mail/recipient-input";
import { ComposeDialog } from "@/app/mail/compose-dialog";

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
        <ComposeDialog open={true} onOpenChange={() => {}} />
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
        <ComposeDialog open={true} onOpenChange={() => {}} />
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
        />
      </Wrapper>,
    );

    // Click close button (has content because initialTo is set)
    await user.click(screen.getByLabelText("Close compose"));

    expect(screen.getByText("Discard draft?")).toBeInTheDocument();
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
        <ComposeDialog open={true} onOpenChange={() => {}} />
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
