"use client";

import { Paperclip } from "lucide-react";
import { cn } from "@/lib/utils";
import type { MailMessage } from "@/types/api";

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

function extractSenderName(from: string): string {
  const match = from.match(/^"?([^"<]+)"?\s*</);
  if (match) return match[1].trim();
  return from.split("@")[0];
}

interface MessageRowProps {
  message: MailMessage;
  selected: boolean;
  indented?: boolean;
  onSelect: (id: number) => void;
}

export function MessageRow({
  message,
  selected,
  indented = false,
  onSelect,
}: MessageRowProps) {
  const hasAttachment = message.bodyHtml?.includes("attachment") ?? false;

  return (
    <div
      data-testid="message-row"
      role="row"
      tabIndex={0}
      aria-selected={selected}
      onClick={() => onSelect(message.id)}
      onKeyDown={(e) => {
        if (e.key === "Enter") onSelect(message.id);
      }}
      className={cn(
        "flex cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors",
        "hover:bg-accent/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50",
        selected && "bg-accent",
        indented && "pl-8",
      )}
    >
      {/* Read/unread indicator */}
      <span
        className={cn(
          "size-2 shrink-0 rounded-full",
          message.read ? "bg-transparent" : "bg-primary",
        )}
        aria-label={message.read ? "Read" : "Unread"}
      />

      {/* Sender */}
      <span
        className={cn(
          "w-36 shrink-0 truncate",
          message.read
            ? "text-muted-foreground"
            : "font-medium text-foreground",
        )}
      >
        {extractSenderName(message.from)}
      </span>

      {/* Subject */}
      <span
        className={cn(
          "flex-1 truncate",
          message.read
            ? "text-muted-foreground"
            : "font-medium text-foreground",
        )}
      >
        {message.subject}
      </span>

      {/* Attachment icon */}
      {hasAttachment && (
        <Paperclip className="size-3.5 shrink-0 text-muted-foreground" />
      )}

      {/* Date */}
      <span className="w-16 shrink-0 text-right text-xs text-muted-foreground">
        {formatDate(message.receivedAt)}
      </span>
    </div>
  );
}
