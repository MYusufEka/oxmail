"use client";

import { Button } from "@/components/ui/button";
import type { LogEntry } from "@/types/api";

export type ServiceFilter = LogEntry["service"] | "all";
export type LevelFilter = LogEntry["level"] | "all";

interface LogFiltersProps {
  service: ServiceFilter;
  level: LevelFilter;
  search: string;
  onServiceChange: (service: ServiceFilter) => void;
  onLevelChange: (level: LevelFilter) => void;
  onSearchChange: (search: string) => void;
  onClear: () => void;
}

export function LogFilters({
  service,
  level,
  search,
  onServiceChange,
  onLevelChange,
  onSearchChange,
  onClear,
}: LogFiltersProps) {
  return (
    <div data-testid="log-filters" className="flex flex-wrap items-center gap-2">
      <select
        data-testid="filter-service"
        value={service}
        onChange={(e) => onServiceChange(e.target.value as ServiceFilter)}
        className="h-8 rounded-md border border-border bg-secondary px-2 text-xs text-foreground outline-none focus:ring-2 focus:ring-ring/50"
      >
        <option value="all">All Services</option>
        <option value="postfix">Postfix</option>
        <option value="dovecot">Dovecot</option>
        <option value="rspamd">Rspamd</option>
        <option value="api">API</option>
      </select>

      <select
        data-testid="filter-level"
        value={level}
        onChange={(e) => onLevelChange(e.target.value as LevelFilter)}
        className="h-8 rounded-md border border-border bg-secondary px-2 text-xs text-foreground outline-none focus:ring-2 focus:ring-ring/50"
      >
        <option value="all">All Levels</option>
        <option value="error">Error</option>
        <option value="warn">Warn</option>
        <option value="info">Info</option>
        <option value="debug">Debug</option>
      </select>

      <input
        data-testid="filter-search"
        type="text"
        placeholder="Search logs..."
        value={search}
        onChange={(e) => onSearchChange(e.target.value)}
        className="h-8 w-48 rounded-md border border-border bg-secondary px-2 text-xs text-foreground placeholder:text-muted-foreground outline-none focus:ring-2 focus:ring-ring/50"
      />

      <Button
        data-testid="clear-logs"
        variant="ghost"
        size="sm"
        onClick={onClear}
      >
        Clear
      </Button>
    </div>
  );
}
