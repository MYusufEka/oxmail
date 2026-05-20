"use client";

import { Paperclip, Loader2 } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import type { MailMessage } from "@/types/api";

function formatFullDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    weekday: "short",
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

interface Attachment {
  filename: string;
  size: number;
}

function extractAttachments(bodyHtml?: string): Attachment[] {
  if (!bodyHtml) return [];
  // Placeholder: in a real implementation, attachments come from the API
  // For now, return empty array
  return [];
}

interface MessagePreviewProps {
  message: MailMessage | null;
  isLoading: boolean;
}

export function MessagePreview({ message, isLoading }: MessagePreviewProps) {
  if (isLoading) {
    return (
      <div
        data-testid="message-preview-loading"
        className="flex h-full items-center justify-center"
      >
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!message) {
    return (
      <div
        data-testid="message-preview-empty"
        className="flex h-full items-center justify-center"
      >
        <p className="text-sm text-muted-foreground">
          Select a message to read
        </p>
      </div>
    );
  }

  const attachments = extractAttachments(message.bodyHtml);

  return (
    <div
      data-testid="message-preview"
      className="flex h-full flex-col overflow-hidden"
    >
      {/* Header */}
      <div className="shrink-0 border-b border-border px-5 py-4">
        <h3 className="text-base font-semibold text-foreground">
          {message.subject}
        </h3>
        <div className="mt-2 flex items-center gap-2 text-sm">
          <span className="font-medium text-foreground">{message.from}</span>
          <span className="text-muted-foreground">→</span>
          <span className="text-muted-foreground">
            {message.to.join(", ")}
          </span>
        </div>
        {message.cc && message.cc.length > 0 && (
          <div className="mt-1 text-xs text-muted-foreground">
            CC: {message.cc.join(", ")}
          </div>
        )}
        <div className="mt-1 text-xs text-muted-foreground">
          {formatFullDate(message.receivedAt)}
        </div>
      </div>

      {/* Body */}
      <div className="flex-1 overflow-y-auto px-5 py-4">
        {message.bodyHtml ? (
          <div
            data-testid="message-body-html"
            className="prose prose-sm prose-invert max-w-none text-foreground"
            dangerouslySetInnerHTML={{ __html: message.bodyHtml }}
          />
        ) : (
          <pre
            data-testid="message-body-text"
            className="whitespace-pre-wrap font-sans text-sm text-foreground"
          >
            {message.bodyText ?? ""}
          </pre>
        )}
      </div>

      {/* Attachments */}
      {attachments.length > 0 && (
        <div className="shrink-0 border-t border-border px-5 py-3">
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Paperclip className="size-3.5" />
            <span>{attachments.length} attachment(s)</span>
          </div>
          <div className="mt-2 flex flex-wrap gap-2">
            {attachments.map((attachment) => (
              <div
                key={attachment.filename}
                className="flex items-center gap-2 rounded-md border border-border bg-secondary/50 px-3 py-1.5 text-xs"
              >
                <span className="text-foreground">{attachment.filename}</span>
                <span className="text-muted-foreground">
                  {formatFileSize(attachment.size)}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
