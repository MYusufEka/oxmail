"use client";

import { usePathname } from "next/navigation";
import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { NAV_ITEMS } from "@/components/layout/sidebar";

interface TopbarProps {
  onOpenCommandPalette: () => void;
}

export function Topbar({ onOpenCommandPalette }: TopbarProps) {
  const pathname = usePathname();

  const currentPage =
    NAV_ITEMS.find((item) =>
      item.href === "/" ? pathname === "/" : pathname.startsWith(item.href)
    ) ?? NAV_ITEMS[0];

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
        <div className="flex items-center gap-2">
          <span className="size-2 rounded-full bg-emerald-500" aria-label="System healthy" />
          <span className="text-xs text-muted-foreground">Healthy</span>
        </div>

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
