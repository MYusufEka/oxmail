"use client";

import { Search } from "lucide-react";
import { MessageRow } from "./message-row";
import { Skeleton } from "@/components/ui/skeleton";
import type { MailMessage } from "@/types/api";

function highlightText(text: string, query: string): React.ReactNode {
  if (!query.trim()) return text;

  // Strip filter prefixes from query for highlighting
  const cleanQuery = query
    .replace(/\b(from|to|subject|has):\S+/gi, "")
    .trim();

  if (!cleanQuery) return text;

  const escapedQuery = cleanQuery.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const regex = new RegExp(`(${escapedQuery})`, "gi");
  const parts = text.split(regex);

  if (parts.length === 1) return text;

  return parts.map((part, idx) =>
    regex.test(part) ? (
      <mark
        key={idx}
        className="rounded-sm bg-primary/20 px-0.5 text-foreground"
      >
        {part}
      </mark>
    ) : (
      part
    ),
  );
}

interface SearchResultsProps {
  results: MailMessage[];
  query: string;
  selectedId: number | null;
  isLoading: boolean;
  onSelect: (id: number) => void;
}

export function SearchResults({
  results,
  query,
  selectedId,
  isLoading,
  onSelect,
}: SearchResultsProps) {
  if (isLoading) {
    return (
      <div className="flex flex-col" data-testid="search-results-loading">
        {Array.from({ length: 5 }).map((_, idx) => (
          <div
            key={idx}
            className="flex items-center gap-3 border-b border-border px-4 py-2.5"
          >
            <Skeleton className="size-2 rounded-full" />
            <Skeleton className="h-4 w-28" />
            <Skeleton className="h-4 flex-1" />
            <Skeleton className="h-3 w-14" />
          </div>
        ))}
      </div>
    );
  }

  if (results.length === 0) {
    return (
      <div
        data-testid="search-empty"
        className="flex h-full flex-col items-center justify-center gap-3 p-8"
      >
        <Search className="size-10 text-muted-foreground/50" />
        <p className="text-sm text-muted-foreground">
          No messages match your search
        </p>
      </div>
    );
  }

  return (
    <div
      role="grid"
      aria-label="Search results"
      className="h-full overflow-y-auto"
      data-testid="search-results"
    >
      <div className="border-b border-border bg-accent/30 px-4 py-1.5">
        <p className="text-xs text-muted-foreground">
          {results.length} result{results.length !== 1 ? "s" : ""} for &ldquo;
          {query}&rdquo;
        </p>
      </div>
      {results.map((message) => (
        <SearchResultRow
          key={message.id}
          message={message}
          query={query}
          selected={message.id === selectedId}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
}

interface SearchResultRowProps {
  message: MailMessage;
  query: string;
  selected: boolean;
  onSelect: (id: number) => void;
}

function SearchResultRow({
  message,
  query,
  selected,
  onSelect,
}: SearchResultRowProps) {
  // Use MessageRow for consistent styling, but wrap subject with highlighting
  return (
    <div data-testid="search-result-row">
      <MessageRow
        message={{
          ...message,
          subject: message.subject,
        }}
        selected={selected}
        onSelect={onSelect}
      />
      {/* Snippet with highlighted matching text */}
      {message.bodyText && (
        <div className="border-b border-border/50 px-4 pb-2 pl-12">
          <p className="truncate text-xs text-muted-foreground">
            {highlightText(
              message.bodyText.slice(0, 120),
              query,
            )}
          </p>
        </div>
      )}
    </div>
  );
}
