"use client";

import { useState } from "react";
import { ChevronDown, ChevronRight, Users } from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { MessageRow } from "./message-row";
import type { MailThread } from "@/types/api";

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();

  if (isToday) {
    return date.toLocaleTimeString("en-US", {
      hour: "numeric",
      minute: "2-digit",
      hour12: true,
    });
  }

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

interface ThreadRowProps {
  thread: MailThread;
  selectedId: number | null;
  contactsMap?: Record<string, string>;
  onSelect: (id: number) => void;
}

export function ThreadRow({
  thread,
  selectedId,
  contactsMap,
  onSelect,
}: ThreadRowProps) {
  const [expanded, setExpanded] = useState(false);
  const isMultiMessage = thread.messages.length > 1;
  const isSelected = thread.messages.some((m) => m.id === selectedId);

  return (
    <div data-testid="thread-row">
      <div
        role="row"
        tabIndex={0}
        aria-selected={isSelected}
        aria-expanded={isMultiMessage ? expanded : undefined}
        onClick={() => {
          if (isMultiMessage) {
            setExpanded((prev) => !prev);
          } else if (thread.messages[0]) {
            onSelect(thread.messages[0].id);
          }
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter") {
            if (isMultiMessage) {
              setExpanded((prev) => !prev);
            } else if (thread.messages[0]) {
              onSelect(thread.messages[0].id);
            }
          }
        }}
        className={cn(
          "flex cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors",
          "hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
          isSelected && !expanded && "bg-accent",
        )}
      >
        <span
          className={cn(
            "size-2 shrink-0 rounded-full",
            thread.unreadCount > 0 ? "bg-primary" : "bg-transparent",
          )}
          aria-label={thread.unreadCount > 0 ? "Has unread" : "All read"}
        />

        {isMultiMessage ? (
          <span className="shrink-0 text-muted-foreground">
            {expanded ? (
              <ChevronDown className="size-3.5" />
            ) : (
              <ChevronRight className="size-3.5" />
            )}
          </span>
        ) : (
          <span className="size-3.5 shrink-0" />
        )}

        <span className="w-28 shrink-0 truncate font-medium">
          {thread.subject}
        </span>

        <span className="flex min-w-0 flex-1 items-center gap-1.5 truncate text-muted-foreground">
          <span className="truncate">{thread.subject}</span>
          {isMultiMessage && (
            <Badge
              variant="secondary"
              className="shrink-0 px-1.5 py-0 text-xs"
              data-testid="thread-count-badge"
            >
              {thread.messages.length}
            </Badge>
          )}
        </span>

        <div className="flex shrink-0 items-center gap-1.5">
          {thread.participantCount > 1 && (
            <span
              className="flex items-center gap-0.5 text-xs text-muted-foreground"
              aria-label={`${thread.participantCount} participants`}
            >
              <Users className="size-3" />
              {thread.participantCount}
            </span>
          )}
          {thread.unreadCount > 0 && (
            <Badge
              variant="default"
              className="shrink-0 px-1.5 py-0 text-xs"
              data-testid="thread-unread-badge"
            >
              {thread.unreadCount}
            </Badge>
          )}
          <span className="text-xs text-muted-foreground">
            {thread.lastDate ? formatDate(thread.lastDate) : ""}
          </span>
        </div>
      </div>

      {expanded &&
        thread.messages.map((message) => (
          <MessageRow
            key={message.id}
            message={message}
            selected={message.id === selectedId}
            indented
            contactsMap={contactsMap}
            onSelect={onSelect}
          />
        ))}
    </div>
  );
}
