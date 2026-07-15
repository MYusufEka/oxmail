"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { Search, Sun, Moon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { NAV_ITEMS } from "@/components/layout/sidebar";
import { useHealth } from "@/hooks/use-health";
import { cn } from "@/lib/utils";

interface TopbarProps {
  onOpenCommandPalette: () => void;
}

export function Topbar({ onOpenCommandPalette }: TopbarProps) {
  const pathname = usePathname();
  const { data: healthData } = useHealth();
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const currentPage =
    NAV_ITEMS.find((item) =>
      item.href === "/" ? pathname === "/" : pathname.startsWith(item.href)
    ) ?? NAV_ITEMS[0];

  const overallStatus = healthData?.status ?? "unhealthy";
  const anyUnhealthy = healthData?.services.some((s) => s.status === "unhealthy") ?? false;
  const derivedStatus = anyUnhealthy ? "degraded" : overallStatus;

  const badgeColor =
    derivedStatus === "healthy"
      ? "bg-emerald-500"
      : derivedStatus === "degraded"
        ? "bg-amber-500"
        : "bg-red-500";

  const badgeLabel =
    derivedStatus === "healthy"
      ? "Healthy"
      : derivedStatus === "degraded"
        ? "Degraded"
        : "Unhealthy";

  return (
    <header
      data-testid="topbar"
      className="flex h-14 shrink-0 items-center justify-between border-b border-border bg-background px-4"
    >
      <div className="flex items-center gap-2">
        <h1 className="text-sm font-semibold text-foreground">
          {currentPage.label}
        </h1>
      </div>

      <div className="flex items-center gap-3">
        <div className="flex items-center gap-2" aria-label={`System status: ${badgeLabel}`}>
          <span className={cn("size-2 rounded-full", badgeColor)} />
          <span className="text-xs text-muted-foreground">{badgeLabel}</span>
        </div>

        <Button
          variant="ghost"
          size="icon"
          className="size-8 text-muted-foreground"
          onClick={() => setTheme((theme ?? "dark") === "dark" ? "light" : "dark")}
          aria-label={`Switch to ${(theme ?? "dark") === "dark" ? "light" : "dark"} mode`}
          data-testid="theme-toggle"
        >
          {mounted ? (theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />) : <span className="size-4" />}
        </Button>

        <Button
          variant="outline"
          size="sm"
          onClick={onOpenCommandPalette}
          className="gap-2 text-muted-foreground"
        >
          <Search className="size-3.5" />
          <span className="text-xs">Search...</span>
          <kbd className="pointer-events-none ml-2 inline-flex h-5 items-center gap-0.5 rounded border border-border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">
            <span className="text-xs">⌘</span>K
          </kbd>
        </Button>
      </div>
    </header>
  );
}
