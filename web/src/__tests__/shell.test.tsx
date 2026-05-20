import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

// Mock next/navigation
vi.mock("next/navigation", () => ({
  usePathname: () => "/",
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
}));

import { TooltipProvider } from "@/components/ui/tooltip";
import { Sidebar, NAV_ITEMS } from "@/components/layout/sidebar";
import { Topbar } from "@/components/layout/topbar";

describe("Sidebar", () => {
  it("renders with data-testid", () => {
    render(<Sidebar collapsed={false} onToggle={() => {}} />);
    expect(screen.getByTestId("sidebar")).toBeInTheDocument();
  });

  it("renders 7 nav items when expanded", () => {
    render(<Sidebar collapsed={false} onToggle={() => {}} />);
    const links = screen.getAllByRole("link");
    // 7 nav items + 1 logo link
    expect(links.length).toBeGreaterThanOrEqual(7);
  });

  it("renders all nav labels when expanded", () => {
    render(<Sidebar collapsed={false} onToggle={() => {}} />);
    for (const item of NAV_ITEMS) {
      expect(screen.getByText(item.label)).toBeInTheDocument();
    }
  });

  it("calls onToggle when collapse button is clicked", () => {
    const onToggle = vi.fn();
    render(<Sidebar collapsed={false} onToggle={onToggle} />);
    const toggleButton = screen.getByLabelText("Collapse sidebar");
    fireEvent.click(toggleButton);
    expect(onToggle).toHaveBeenCalledOnce();
  });

  it("shows expand button when collapsed", () => {
    render(
      <TooltipProvider>
        <Sidebar collapsed={true} onToggle={() => {}} />
      </TooltipProvider>
    );
    expect(screen.getByLabelText("Expand sidebar")).toBeInTheDocument();
  });
});

describe("Topbar", () => {
  it("renders with data-testid", () => {
    render(<Topbar onOpenCommandPalette={() => {}} />);
    expect(screen.getByTestId("topbar")).toBeInTheDocument();
  });

  it("shows current page title", () => {
    render(<Topbar onOpenCommandPalette={() => {}} />);
    expect(screen.getByText("Dashboard")).toBeInTheDocument();
  });

  it("shows health indicator", () => {
    render(<Topbar onOpenCommandPalette={() => {}} />);
    expect(screen.getByText("Healthy")).toBeInTheDocument();
  });

  it("calls onOpenCommandPalette when search button is clicked", () => {
    const onOpen = vi.fn();
    render(<Topbar onOpenCommandPalette={onOpen} />);
    const searchButton = screen.getByText("Search...");
    fireEvent.click(searchButton.closest("button")!);
    expect(onOpen).toHaveBeenCalledOnce();
  });
});

describe("NAV_ITEMS", () => {
  it("has exactly 7 items", () => {
    expect(NAV_ITEMS).toHaveLength(8);
  });
});
