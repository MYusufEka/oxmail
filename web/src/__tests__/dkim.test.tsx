import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import DkimPage from "@/app/dkim/page";
import { DkimDomainCard } from "@/app/dkim/dkim-domain-card";
import { DnsRecordDisplay } from "@/app/dkim/dns-record-display";
import type { PaginatedResponse, Domain, DKIMKey } from "@/types/api";

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

const mockDomains: PaginatedResponse<Domain> = {
  data: [
    { id: 1, name: "example.com", active: true, createdAt: "2024-01-01T00:00:00Z", updatedAt: "2024-01-01T00:00:00Z" },
    { id: 2, name: "test.org", active: true, createdAt: "2024-01-02T00:00:00Z", updatedAt: "2024-01-02T00:00:00Z" },
  ],
  pagination: { page: 1, limit: 20, total: 2 },
};

const mockDkimKey: DKIMKey = {
  domain: "example.com",
  selector: "default",
  publicKey: "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A",
  dnsRecord: "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A",
  createdAt: "2024-01-01T00:00:00Z",
};

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
  Object.assign(navigator, {
    clipboard: {
      writeText: vi.fn().mockResolvedValue(undefined),
    },
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

function mockFetchResponses(...responses: Array<{ status: number; body: unknown }>) {
  const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
  for (const response of responses) {
    fetchMock.mockResolvedValueOnce({
      ok: response.status >= 200 && response.status < 300,
      status: response.status,
      json: () => Promise.resolve(response.body),
    });
  }
}

describe("DkimPage", () => {
  it("shows loading state", () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockReturnValue(new Promise(() => {}));

    render(<DkimPage />, { wrapper: createWrapper() });

    expect(screen.getByText("DKIM")).toBeInTheDocument();
  });

  it("shows empty state when no domains exist", async () => {
    mockFetchResponses({
      status: 200,
      body: { data: [], pagination: { page: 1, limit: 20, total: 0 } },
    });

    render(<DkimPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("No domains configured")).toBeInTheDocument();
    });
    expect(screen.getByText("Add a domain first to manage DKIM keys.")).toBeInTheDocument();
  });

  it("renders domain cards when domains exist", async () => {
    mockFetchResponses(
      { status: 200, body: mockDomains },
      { status: 200, body: mockDkimKey },
      { status: 404, body: { error: { code: "NOT_FOUND", message: "No DKIM key" } } },
    );

    render(<DkimPage />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("example.com")).toBeInTheDocument();
    });
  });
});

describe("DkimDomainCard", () => {
  it("shows Generate Key button when no DKIM key exists", async () => {
    (globalThis.fetch as ReturnType<typeof vi.fn>).mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: { code: "NOT_FOUND", message: "No DKIM key" } }),
    });

    render(<DkimDomainCard domain="test.org" />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("Not Generated")).toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: /Generate Key/i })).toBeInTheDocument();
  });

  it("shows key details when DKIM key exists", async () => {
    mockFetchResponses({ status: 200, body: mockDkimKey });

    render(<DkimDomainCard domain="example.com" />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByText("Active")).toBeInTheDocument();
    });
    expect(screen.getByText("RSA")).toBeInTheDocument();
    expect(screen.getByText("2048")).toBeInTheDocument();
    expect(screen.getByText("default")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Rotate Key/i })).toBeInTheDocument();
  });

  it("shows rotate confirmation dialog", async () => {
    mockFetchResponses({ status: 200, body: mockDkimKey });

    render(<DkimDomainCard domain="example.com" />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Rotate Key/i })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: /Rotate Key/i }));

    await waitFor(() => {
      expect(screen.getByText("Rotate DKIM Key")).toBeInTheDocument();
    });
    expect(
      screen.getByText(/Rotating the key will invalidate the current DNS record/),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Cancel/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Confirm Rotate/i })).toBeInTheDocument();
  });

  it("calls generate mutation on Generate Key click", async () => {
    const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValueOnce({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: { code: "NOT_FOUND", message: "No DKIM key" } }),
    });

    render(<DkimDomainCard domain="test.org" />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: /Generate Key/i })).toBeInTheDocument();
    });

    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ ...mockDkimKey, domain: "test.org" }),
    });

    fireEvent.click(screen.getByRole("button", { name: /Generate Key/i }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
  });
});

describe("DnsRecordDisplay", () => {
  it("renders the full DNS record", () => {
    render(
      <DnsRecordDisplay
        selector="default"
        domain="example.com"
        dnsRecord="v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A"
      />,
    );

    expect(
      screen.getByText(
        'default._domainkey.example.com IN TXT "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A"',
      ),
    ).toBeInTheDocument();
  });

  it("copies record to clipboard on button click", async () => {
    render(
      <DnsRecordDisplay
        selector="default"
        domain="example.com"
        dnsRecord="v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A"
      />,
    );

    const copyButton = screen.getByRole("button", { name: /Copy DNS record/i });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        'default._domainkey.example.com IN TXT "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A"',
      );
    });

    await waitFor(() => {
      expect(screen.getByText("Copied!")).toBeInTheDocument();
    });
  });
});
