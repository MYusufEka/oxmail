"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Clock, Search, X } from "lucide-react";
import { cn } from "@/lib/utils";

const RECENT_SEARCHES_KEY = "oxmail-recent-searches";
const MAX_RECENT_SEARCHES = 5;

interface FilterChip {
  type: "from" | "to" | "subject" | "has";
  value: string;
}

function parseFilterChips(query: string): {
  chips: FilterChip[];
  remainder: string;
} {
  const chips: FilterChip[] = [];
  let remainder = query;

  const filterPattern = /\b(from|to|subject|has):(\S+)/gi;
  let match: RegExpExecArray | null = filterPattern.exec(query);

  while (match !== null) {
    const type = match[1].toLowerCase() as FilterChip["type"];
    const value = match[2];
    chips.push({ type, value });
    match = filterPattern.exec(query);
  }

  remainder = query.replace(filterPattern, "").trim();
  return { chips, remainder };
}

function getRecentSearches(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const stored = localStorage.getItem(RECENT_SEARCHES_KEY);
    if (!stored) return [];
    const parsed: unknown = JSON.parse(stored);
    if (Array.isArray(parsed)) {
      return parsed.filter((s): s is string => typeof s === "string");
    }
    return [];
  } catch {
    return [];
  }
}

function saveRecentSearch(query: string): void {
  if (typeof window === "undefined" || !query.trim()) return;
  const recent = getRecentSearches().filter((s) => s !== query);
  recent.unshift(query);
  const trimmed = recent.slice(0, MAX_RECENT_SEARCHES);
  localStorage.setItem(RECENT_SEARCHES_KEY, JSON.stringify(trimmed));
}

interface SearchBarProps {
  onSearch: (query: string) => void;
  onClear: () => void;
  isSearching: boolean;
}

export function SearchBar({ onSearch, onClear, isSearching }: SearchBarProps) {
  const [inputValue, setInputValue] = useState("");
  const [showDropdown, setShowDropdown] = useState(false);
  const [recentSearches, setRecentSearches] = useState<string[]>([]);
  const inputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const { chips } = parseFilterChips(inputValue);

  useEffect(() => {
    setRecentSearches(getRecentSearches());
  }, []);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setShowDropdown(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSubmit = useCallback(
    (query: string) => {
      const trimmed = query.trim();
      if (!trimmed) return;
      saveRecentSearch(trimmed);
      setRecentSearches(getRecentSearches());
      setShowDropdown(false);
      onSearch(trimmed);
    },
    [onSearch],
  );

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent<HTMLInputElement>) => {
      if (event.key === "Enter") {
        event.preventDefault();
        handleSubmit(inputValue);
      }
      if (event.key === "Escape") {
        setShowDropdown(false);
        inputRef.current?.blur();
      }
    },
    [handleSubmit, inputValue],
  );

  const handleClear = useCallback(() => {
    setInputValue("");
    setShowDropdown(false);
    onClear();
    inputRef.current?.focus();
  }, [onClear]);

  const handleFocus = useCallback(() => {
    setRecentSearches(getRecentSearches());
    setShowDropdown(true);
  }, []);

  const handleRecentClick = useCallback(
    (query: string) => {
      setInputValue(query);
      handleSubmit(query);
    },
    [handleSubmit],
  );

  const handleRemoveChip = useCallback(
    (chipToRemove: FilterChip) => {
      const pattern = new RegExp(
        `\\b${chipToRemove.type}:${chipToRemove.value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\s*`,
        "i",
      );
      const newValue = inputValue.replace(pattern, "").trim();
      setInputValue(newValue);
      if (isSearching && newValue) {
        handleSubmit(newValue);
      } else if (!newValue) {
        onClear();
      }
    },
    [inputValue, isSearching, handleSubmit, onClear],
  );

  return (
    <div ref={containerRef} className="relative w-full max-w-md">
      {/* Search input */}
      <div
        className={cn(
          "flex items-center gap-2 overflow-hidden rounded-md border border-border bg-card px-3 py-1.5 text-sm transition-colors",
          "focus-within:border-primary/50 focus-within:ring-2 focus-within:ring-inset focus-within:ring-ring/20",
        )}
      >
        <Search className="size-4 shrink-0 text-muted-foreground" />
        <input
          ref={inputRef}
          data-testid="search-input"
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          onFocus={handleFocus}
          placeholder="Search mail... (from:, to:, subject:, has:attachment)"
          aria-label="Search mail"
          className="flex-1 bg-transparent text-foreground placeholder:text-muted-foreground/60 focus:outline-none"
        />
        {(inputValue || isSearching) && (
          <button
            data-testid="search-clear"
            type="button"
            onClick={handleClear}
            aria-label="Clear search"
            className="shrink-0 rounded p-0.5 text-muted-foreground transition-colors hover:text-foreground"
          >
            <X className="size-4" />
          </button>
        )}
      </div>

      {/* Filter chips */}
      {chips.length > 0 && (
        <div
          data-testid="filter-chips"
          className="mt-1.5 flex flex-wrap gap-1.5"
        >
          {chips.map((chip, idx) => (
            <span
              key={`${chip.type}-${chip.value}-${idx}`}
              className="inline-flex items-center gap-1 rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary"
            >
              {chip.type}:{chip.value}
              <button
                type="button"
                onClick={() => handleRemoveChip(chip)}
                aria-label={`Remove filter ${chip.type}:${chip.value}`}
                className="ml-0.5 rounded-full p-0.5 transition-colors hover:bg-primary/20"
              >
                <X className="size-3" />
              </button>
            </span>
          ))}
        </div>
      )}

      {/* Recent searches dropdown */}
      {showDropdown && recentSearches.length > 0 && !isSearching && (
        <div
          data-testid="recent-searches"
          className="absolute top-full z-50 mt-1 w-full rounded-md border border-border bg-card p-1 shadow-md"
        >
          <p className="px-2 py-1 text-xs font-medium text-muted-foreground">
            Recent searches
          </p>
          {recentSearches.map((query) => (
            <button
              key={query}
              type="button"
              onClick={() => handleRecentClick(query)}
              className="flex w-full items-center gap-2 rounded-sm px-2 py-1.5 text-left text-sm text-foreground transition-colors hover:bg-accent"
            >
              <Clock className="size-3.5 text-muted-foreground" />
              <span className="truncate">{query}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
