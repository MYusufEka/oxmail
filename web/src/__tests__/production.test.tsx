import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { ReactNode } from "react";
import ProductionPage from "@/app/production/page";
import { DnsRecordStep } from "@/app/production/dns-record-step";
import { DnsWizard } from "@/app/production/dns-wizard";
import { ProductionSettings } from "@/app/production/production-settings";

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
  Object.assign(navigator, {
    clipboard: {
      writeText: vi.fn().mockResolvedValue(undefined),
    },
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("ProductionPage", () => {
  it("renders the page with title and shield icon", () => {
    render(<ProductionPage />, { wrapper: createWrapper() });

    expect(screen.getByText("Production")).toBeInTheDocument();
    expect(screen.getByTestId("production-page")).toBeInTheDocument();
  });

  it("shows dev mode warning banner", () => {
    render(<ProductionPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("dev-mode-warning")).toBeInTheDocument();
    expect(
      screen.getByText("You're viewing production settings in dev mode")
    ).toBeInTheDocument();
  });

  it("renders DNS wizard section", () => {
    render(<ProductionPage />, { wrapper: createWrapper() });

    expect(screen.getByText("DNS Setup Wizard")).toBeInTheDocument();
    expect(screen.getByTestId("dns-wizard")).toBeInTheDocument();
  });

  it("renders production settings panel", () => {
    render(<ProductionPage />, { wrapper: createWrapper() });

    expect(screen.getByTestId("production-settings")).toBeInTheDocument();
  });
});

describe("DnsRecordStep", () => {
  const defaultProps = {
    stepNumber: 1,
    title: "MX Record",
    recordType: "MX",
    recordName: "example.com",
    recordValue: "10 mail.example.com",
    verified: null,
    verifying: false,
    onVerify: vi.fn(),
  };

  it("renders step with record details", () => {
    render(<DnsRecordStep {...defaultProps} />);

    expect(screen.getByTestId("dns-step-1")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "MX Record" })).toBeInTheDocument();
    expect(
      screen.getByText("example.com IN MX 10 mail.example.com")
    ).toBeInTheDocument();
  });

  it("copies record to clipboard on button click", async () => {
    render(<DnsRecordStep {...defaultProps} />);

    const copyButton = screen.getByRole("button", {
      name: /Copy MX Record record/i,
    });
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        "example.com IN MX 10 mail.example.com"
      );
    });

    await waitFor(() => {
      expect(screen.getByText("Copied!")).toBeInTheDocument();
    });
  });

  it("shows Check DNS button and calls onVerify", () => {
    const onVerify = vi.fn();
    render(<DnsRecordStep {...defaultProps} onVerify={onVerify} />);

    const verifyButton = screen.getByTestId("verify-step-1");
    fireEvent.click(verifyButton);

    expect(onVerify).toHaveBeenCalledTimes(1);
  });

  it("shows verified state with green checkmark", () => {
    render(<DnsRecordStep {...defaultProps} verified={true} />);

    expect(screen.getByText("Verified")).toBeInTheDocument();
  });

  it("shows failed state with red indicator", () => {
    render(<DnsRecordStep {...defaultProps} verified={false} />);

    expect(screen.getByText("Not found")).toBeInTheDocument();
  });

  it("shows loading state when verifying", () => {
    render(<DnsRecordStep {...defaultProps} verifying={true} />);

    const verifyButton = screen.getByTestId("verify-step-1");
    expect(verifyButton).toBeDisabled();
  });

  it("shows manual-only message for rDNS step", () => {
    render(
      <DnsRecordStep
        {...defaultProps}
        stepNumber={5}
        title="rDNS / PTR Record"
        manualOnly={true}
      />
    );

    expect(
      screen.getByText(
        "This record must be configured by your hosting provider. Verification is manual."
      )
    ).toBeInTheDocument();
    expect(screen.queryByTestId("verify-step-5")).not.toBeInTheDocument();
  });

  it("renders description when provided", () => {
    render(
      <DnsRecordStep
        {...defaultProps}
        description="Routes incoming email to your mail server."
      />
    );

    expect(
      screen.getByText("Routes incoming email to your mail server.")
    ).toBeInTheDocument();
  });
});

describe("DnsWizard", () => {
  it("renders all 5 DNS steps", () => {
    render(<DnsWizard domain="example.com" />, { wrapper: createWrapper() });

    expect(screen.getByTestId("dns-step-1")).toBeInTheDocument();
    expect(screen.getByTestId("dns-step-2")).toBeInTheDocument();
    expect(screen.getByTestId("dns-step-3")).toBeInTheDocument();
    expect(screen.getByTestId("dns-step-4")).toBeInTheDocument();
    expect(screen.getByTestId("dns-step-5")).toBeInTheDocument();
  });

  it("shows progress bar at 0/5 initially", () => {
    render(<DnsWizard domain="example.com" />, { wrapper: createWrapper() });

    expect(screen.getByTestId("dns-progress-text")).toHaveTextContent(
      "0/5 records verified"
    );
  });

  it("shows progress bar with correct aria attributes", () => {
    render(<DnsWizard domain="example.com" />, { wrapper: createWrapper() });

    const progressBar = screen.getByTestId("dns-progress-bar");
    expect(progressBar).toHaveAttribute("role", "progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "0");
    expect(progressBar).toHaveAttribute("aria-valuemax", "5");
  });

  it("calls DNS check API when verify is clicked", async () => {
    const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          results: [
            {
              domain: "example.com",
              record: "mx",
              expected: "10 mail.example.com",
              actual: "10 mail.example.com",
              valid: true,
            },
          ],
        }),
    });

    render(<DnsWizard domain="example.com" />, { wrapper: createWrapper() });

    const verifyButton = screen.getByTestId("verify-step-1");
    fireEvent.click(verifyButton);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    await waitFor(() => {
      expect(screen.getByTestId("dns-progress-text")).toHaveTextContent(
        "1/5 records verified"
      );
    });
  });

  it("shows failed state when DNS check returns invalid", async () => {
    const fetchMock = globalThis.fetch as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () =>
        Promise.resolve({
          results: [
            {
              domain: "example.com",
              record: "mx",
              expected: "10 mail.example.com",
              actual: "",
              valid: false,
            },
          ],
        }),
    });

    render(<DnsWizard domain="example.com" />, { wrapper: createWrapper() });

    const verifyButton = screen.getByTestId("verify-step-1");
    fireEvent.click(verifyButton);

    await waitFor(() => {
      expect(screen.getByText("Not found")).toBeInTheDocument();
    });
  });
});

describe("ProductionSettings", () => {
  const defaultProps = {
    hostname: "mail.example.com",
    publicIp: "203.0.113.1",
    tlsEnabled: true,
    outboundRateLimit: 500,
  };

  it("renders all settings fields", () => {
    render(<ProductionSettings {...defaultProps} />);

    expect(screen.getByText("Server Configuration")).toBeInTheDocument();
    expect(screen.getAllByText("mail.example.com").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("203.0.113.1")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("500 msg/hour")).toBeInTheDocument();
  });

  it("shows inactive TLS badge when disabled", () => {
    render(<ProductionSettings {...defaultProps} tlsEnabled={false} />);

    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });

  it("shows unlimited when rate limit is 0", () => {
    render(<ProductionSettings {...defaultProps} outboundRateLimit={0} />);

    expect(screen.getByText("Unlimited")).toBeInTheDocument();
  });

  it("shows placeholder when public IP is empty", () => {
    render(<ProductionSettings {...defaultProps} publicIp="" />);

    expect(screen.getByText("Not configured")).toBeInTheDocument();
  });

  it("shows dash when hostname is empty", () => {
    render(<ProductionSettings {...defaultProps} hostname="" />);

    expect(screen.getAllByText("—").length).toBeGreaterThanOrEqual(1);
  });
});
