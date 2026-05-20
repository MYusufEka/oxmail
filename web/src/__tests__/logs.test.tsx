import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import type { LogEntry } from "@/types/api";

const mockEntries: LogEntry[] = [
  {
    id: 1,
    timestamp: "2024-01-15T10:30:45.000Z",
    service: "postfix",
    level: "info",
    message: "Mail delivered to user@example.com",
  },
  {
    id: 2,
    timestamp: "2024-01-15T10:30:46.000Z",
    service: "dovecot",
    level: "warn",
    message: "Connection timeout for client",
  },
  {
    id: 3,
    timestamp: "2024-01-15T10:30:47.000Z",
    service: "rspamd",
    level: "error",
    message: "Failed to scan message: timeout",
  },
];

const mockClearEntries = vi.fn();
let mockConnected = true;
let mockEntryState: LogEntry[] = [];

vi.mock("@/hooks/use-logs", () => ({
  useLogStream: () => ({
    entries: mockEntryState,
    connected: mockConnected,
    clearEntries: mockClearEntries,
  }),
}));

import LogsPage from "@/app/logs/page";
import { LogEntryRow } from "@/app/logs/log-entry";
import { LogFilters } from "@/app/logs/log-filters";

describe("LogEntryRow", () => {
  it("renders log entry with timestamp, service, level, and message", () => {
    render(<LogEntryRow entry={mockEntries[0]} />);
    const row = screen.getByTestId("log-entry");
    expect(row).toBeInTheDocument();
    expect(row).toHaveTextContent("[postfix]");
    expect(row).toHaveTextContent("[info]");
    expect(row).toHaveTextContent("Mail delivered to user@example.com");
  });

  it("applies blue color class for postfix service", () => {
    const { container } = render(<LogEntryRow entry={mockEntries[0]} />);
    const serviceSpan = container.querySelector(".text-blue-400");
    expect(serviceSpan).toBeInTheDocument();
  });

  it("applies green color class for dovecot service", () => {
    const { container } = render(<LogEntryRow entry={mockEntries[1]} />);
    const serviceSpan = container.querySelector(".text-green-400");
    expect(serviceSpan).toBeInTheDocument();
  });

  it("applies orange color class for rspamd service", () => {
    const { container } = render(<LogEntryRow entry={mockEntries[2]} />);
    const serviceSpan = container.querySelector(".text-orange-400");
    expect(serviceSpan).toBeInTheDocument();
  });

  it("applies red color class for error level", () => {
    const { container } = render(<LogEntryRow entry={mockEntries[2]} />);
    const levelSpan = container.querySelector(".text-red-400");
    expect(levelSpan).toBeInTheDocument();
  });

  it("applies yellow color class for warn level", () => {
    const { container } = render(<LogEntryRow entry={mockEntries[1]} />);
    const levelSpan = container.querySelector(".text-yellow-400");
    expect(levelSpan).toBeInTheDocument();
  });
});

describe("LogFilters", () => {
  const defaultProps = {
    service: "all" as const,
    level: "all" as const,
    search: "",
    onServiceChange: vi.fn(),
    onLevelChange: vi.fn(),
    onSearchChange: vi.fn(),
    onClear: vi.fn(),
  };

  it("renders filter bar with data-testid", () => {
    render(<LogFilters {...defaultProps} />);
    expect(screen.getByTestId("log-filters")).toBeInTheDocument();
  });

  it("renders service dropdown", () => {
    render(<LogFilters {...defaultProps} />);
    const select = screen.getByTestId("filter-service");
    expect(select).toBeInTheDocument();
    expect(select).toHaveValue("all");
  });

  it("renders level dropdown", () => {
    render(<LogFilters {...defaultProps} />);
    const select = screen.getByTestId("filter-level");
    expect(select).toBeInTheDocument();
    expect(select).toHaveValue("all");
  });

  it("renders search input", () => {
    render(<LogFilters {...defaultProps} />);
    const input = screen.getByTestId("filter-search");
    expect(input).toBeInTheDocument();
  });

  it("calls onServiceChange when service dropdown changes", () => {
    render(<LogFilters {...defaultProps} />);
    fireEvent.change(screen.getByTestId("filter-service"), {
      target: { value: "postfix" },
    });
    expect(defaultProps.onServiceChange).toHaveBeenCalledWith("postfix");
  });

  it("calls onLevelChange when level dropdown changes", () => {
    render(<LogFilters {...defaultProps} />);
    fireEvent.change(screen.getByTestId("filter-level"), {
      target: { value: "error" },
    });
    expect(defaultProps.onLevelChange).toHaveBeenCalledWith("error");
  });

  it("calls onSearchChange when search input changes", () => {
    render(<LogFilters {...defaultProps} />);
    fireEvent.change(screen.getByTestId("filter-search"), {
      target: { value: "timeout" },
    });
    expect(defaultProps.onSearchChange).toHaveBeenCalledWith("timeout");
  });

  it("calls onClear when clear button is clicked", () => {
    render(<LogFilters {...defaultProps} />);
    fireEvent.click(screen.getByTestId("clear-logs"));
    expect(defaultProps.onClear).toHaveBeenCalledOnce();
  });
});

describe("LogsPage", () => {
  beforeEach(() => {
    mockConnected = true;
    mockEntryState = [];
    mockClearEntries.mockClear();
  });

  it("renders log viewer container", () => {
    render(<LogsPage />);
    expect(screen.getByTestId("log-viewer")).toBeInTheDocument();
  });

  it("shows empty state when no entries", () => {
    render(<LogsPage />);
    expect(screen.getByTestId("log-empty-state")).toBeInTheDocument();
    expect(
      screen.getByText("No log entries yet. Waiting for events...")
    ).toBeInTheDocument();
  });

  it("shows connection status as connected (green)", () => {
    render(<LogsPage />);
    const status = screen.getByTestId("connection-status");
    expect(status).toBeInTheDocument();
    expect(status).toHaveTextContent("connected");
  });

  it("shows connection status as disconnected (red)", () => {
    mockConnected = false;
    render(<LogsPage />);
    const status = screen.getByTestId("connection-status");
    expect(status).toHaveTextContent("disconnected");
  });

  it("renders log entries when available", () => {
    mockEntryState = mockEntries;
    render(<LogsPage />);
    const logEntries = screen.getAllByTestId("log-entry");
    expect(logEntries).toHaveLength(3);
  });

  it("filters entries by service", () => {
    mockEntryState = mockEntries;
    render(<LogsPage />);
    fireEvent.change(screen.getByTestId("filter-service"), {
      target: { value: "postfix" },
    });
    const logEntries = screen.getAllByTestId("log-entry");
    expect(logEntries).toHaveLength(1);
  });

  it("filters entries by level", () => {
    mockEntryState = mockEntries;
    render(<LogsPage />);
    fireEvent.change(screen.getByTestId("filter-level"), {
      target: { value: "error" },
    });
    const logEntries = screen.getAllByTestId("log-entry");
    expect(logEntries).toHaveLength(1);
  });

  it("filters entries by search text", () => {
    mockEntryState = mockEntries;
    render(<LogsPage />);
    fireEvent.change(screen.getByTestId("filter-search"), {
      target: { value: "timeout" },
    });
    const logEntries = screen.getAllByTestId("log-entry");
    expect(logEntries).toHaveLength(2);
  });

  it("clears entries when clear button is clicked", () => {
    mockEntryState = mockEntries;
    render(<LogsPage />);
    fireEvent.click(screen.getByTestId("clear-logs"));
    expect(mockClearEntries).toHaveBeenCalledOnce();
  });

  it("renders filter controls", () => {
    render(<LogsPage />);
    expect(screen.getByTestId("log-filters")).toBeInTheDocument();
    expect(screen.getByTestId("filter-service")).toBeInTheDocument();
    expect(screen.getByTestId("filter-level")).toBeInTheDocument();
    expect(screen.getByTestId("filter-search")).toBeInTheDocument();
  });
});
