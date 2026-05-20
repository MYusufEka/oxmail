"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { ScrollText, ArrowDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useLogStream } from "@/hooks/use-logs";
import { LogEntryRow } from "./log-entry";
import { LogFilters, type ServiceFilter, type LevelFilter } from "./log-filters";

export default function LogsPage() {
  const { entries, connected, clearEntries } = useLogStream();

  const [service, setService] = useState<ServiceFilter>("all");
  const [level, setLevel] = useState<LevelFilter>("all");
  const [search, setSearch] = useState("");
  const [autoScroll, setAutoScroll] = useState(true);

  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  const filteredEntries = useMemo(() => {
    return entries.filter((entry) => {
      if (service !== "all" && entry.service !== service) return false;
      if (level !== "all" && entry.level !== level) return false;
      if (search && !entry.message.toLowerCase().includes(search.toLowerCase())) {
        return false;
      }
      return true;
    });
  }, [entries, service, level, search]);

  useEffect(() => {
    if (autoScroll && bottomRef.current?.scrollIntoView) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [filteredEntries, autoScroll]);

  const handleScroll = useCallback(() => {
    const container = scrollContainerRef.current;
    if (!container) return;

    const { scrollTop, scrollHeight, clientHeight } = container;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 40;
    setAutoScroll(isAtBottom);
  }, []);

  const handleResumeAutoScroll = useCallback(() => {
    setAutoScroll(true);
    if (bottomRef.current?.scrollIntoView) {
      bottomRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, []);

  const handleClear = useCallback(() => {
    clearEntries();
  }, [clearEntries]);

  const connectionStatus = connected ? "connected" : "disconnected";

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <ScrollText className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Logs</h2>
          <div
            data-testid="connection-status"
            className="flex items-center gap-1.5"
          >
            <span className="relative flex size-2">
              {connected && (
                <span
                  className="absolute inline-flex size-full rounded-full bg-green-500 opacity-75"
                  style={{ animation: "pulse-dot 2s ease-in-out infinite" }}
                />
              )}
              <span
                className={`relative inline-flex size-2 rounded-full ${
                  connected ? "bg-green-500" : "bg-red-500"
                }`}
              />
            </span>
            <span className="text-xs text-muted-foreground capitalize">
              {connectionStatus}
            </span>
          </div>
        </div>

        <LogFilters
          service={service}
          level={level}
          search={search}
          onServiceChange={setService}
          onLevelChange={setLevel}
          onSearchChange={setSearch}
          onClear={handleClear}
        />
      </div>

      <div
        data-testid="log-viewer"
        ref={scrollContainerRef}
        onScroll={handleScroll}
        className="relative flex-1 overflow-y-auto rounded-lg border border-border bg-card"
        style={{ maxHeight: "calc(100vh - 180px)" }}
      >
        {filteredEntries.length === 0 ? (
          <div
            data-testid="log-empty-state"
            className="flex h-64 items-center justify-center"
          >
            <p className="text-sm text-muted-foreground">
              No log entries yet. Waiting for events...
            </p>
          </div>
        ) : (
          <div className="py-2">
            {filteredEntries.map((entry) => (
              <LogEntryRow key={entry.id} entry={entry} />
            ))}
            <div ref={bottomRef} />
          </div>
        )}

        {!autoScroll && filteredEntries.length > 0 && (
          <Button
            data-testid="resume-scroll"
            variant="secondary"
            size="sm"
            className="absolute bottom-3 right-3 shadow-lg"
            onClick={handleResumeAutoScroll}
          >
            <ArrowDown className="size-3" />
            Resume
          </Button>
        )}
      </div>
    </div>
  );
}
