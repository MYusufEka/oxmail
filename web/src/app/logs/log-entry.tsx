"use client";

import { cn } from "@/lib/utils";
import type { LogEntry } from "@/types/api";

const SERVICE_COLORS: Record<LogEntry["service"], string> = {
  postfix: "text-blue-400",
  dovecot: "text-green-400",
  rspamd: "text-orange-400",
  api: "text-purple-400",
};

const LEVEL_COLORS: Record<LogEntry["level"], string> = {
  error: "text-red-400",
  warn: "text-yellow-400",
  info: "text-muted-foreground",
  debug: "text-muted-foreground/70",
};

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

interface LogEntryRowProps {
  entry: LogEntry;
}

export function LogEntryRow({ entry }: LogEntryRowProps) {
  return (
    <div
      data-testid="log-entry"
      className="flex items-start gap-2 px-3 py-1 font-mono text-xs leading-5 hover:bg-secondary/50"
    >
      <span className="shrink-0 text-muted-foreground/70">
        [{formatTimestamp(entry.timestamp)}]
      </span>
      <span className={cn("shrink-0 uppercase", SERVICE_COLORS[entry.service])}>
        [{entry.service}]
      </span>
      <span className={cn("shrink-0 uppercase", LEVEL_COLORS[entry.level])}>
        [{entry.level}]
      </span>
      <span className="text-foreground/90 break-all">{entry.message}</span>
    </div>
  );
}
