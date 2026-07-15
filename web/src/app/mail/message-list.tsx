"use client";

import { useMemo } from "react";
import { Inbox } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { MessageRow } from "./message-row";
import type { MailMessage } from "@/types/api";

function groupByThread(messages: MailMessage[]): MailMessage[][] {
  const threads: Map<string, MailMessage[]> = new Map();

  for (const msg of messages) {
    const normalizedSubject = msg.subject
      .replace(/^(Re:\s*|Fwd:\s*)+/i, "")
      .trim();
    const existing = threads.get(normalizedSubject);
    if (existing) {
      existing.push(msg);
    } else {
      threads.set(normalizedSubject, [msg]);
    }
  }

  return Array.from(threads.values());
}

interface MessageListProps {
  messages: MailMessage[];
  selectedId: number | null;
  isLoading: boolean;
  contactsMap?: Record<string, string>;
  onSelect: (id: number) => void;
}

export function MessageList({
  messages,
  selectedId,
  isLoading,
  contactsMap,
  onSelect,
}: MessageListProps) {
  const threads = useMemo(() => groupByThread(messages), [messages]);

  if (isLoading) {
    return (
      <div className="flex flex-col" data-testid="message-list-loading">
        {Array.from({ length: 8 }).map((_, idx) => (
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

  if (messages.length === 0) {
    return (
      <div
        data-testid="inbox-empty"
        className="flex h-full flex-col items-center justify-center gap-3 p-8"
      >
        <Inbox className="size-10 text-muted-foreground/50" />
        <p className="text-sm text-muted-foreground">Your inbox is empty</p>
      </div>
    );
  }

  return (
    <div
      role="grid"
      aria-label="Message list"
      className="min-h-0 h-full overflow-y-auto"
      data-testid="message-list"
    >
      {threads.map((thread) =>
        thread.map((message, idx) => (
          <MessageRow
            key={message.id}
            message={message}
            selected={message.id === selectedId}
            indented={idx > 0}
            contactsMap={contactsMap}
            onSelect={onSelect}
          />
        )),
      )}
    </div>
  );
}
