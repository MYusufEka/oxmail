"use client";

import { Inbox } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { ThreadRow } from "./thread-row";
import type { MailThread } from "@/types/api";

interface ThreadListProps {
  threads: MailThread[];
  selectedId: number | null;
  isLoading: boolean;
  contactsMap?: Record<string, string>;
  onSelect: (id: number) => void;
}

export function ThreadList({
  threads,
  selectedId,
  isLoading,
  contactsMap,
  onSelect,
}: ThreadListProps) {
  if (isLoading) {
    return (
      <div className="flex flex-col" data-testid="thread-list-loading">
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

  if (threads.length === 0) {
    return (
      <div
        data-testid="thread-list-empty"
        className="flex h-full flex-col items-center justify-center gap-3 p-8"
      >
        <Inbox className="size-10 text-muted-foreground/50" />
        <p className="text-sm text-muted-foreground">No conversations</p>
      </div>
    );
  }

  return (
    <div
      role="grid"
      aria-label="Thread list"
      className="min-h-0 h-full overflow-y-auto"
      data-testid="thread-list"
    >
      {threads.map((thread) => (
        <ThreadRow
          key={thread.threadId}
          thread={thread}
          selectedId={selectedId}
          contactsMap={contactsMap}
          onSelect={onSelect}
        />
      ))}
    </div>
  );
}
