"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Inbox, Mail, PenSquare, Reply, ReplyAll } from "lucide-react";
import { useInbox, useMessage } from "@/hooks/use-mail";
import { apiClient } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { MessageList } from "./message-list";
import { MessagePreview } from "./message-preview";
import { ComposeDialog } from "./compose-dialog";
import { SearchBar } from "./search-bar";
import { SearchResults } from "./search-results";
import type { MailMessage } from "@/types/api";

// Hardcoded user until auth (T30)
const CURRENT_USER_ID = 1;
const CURRENT_USER_EMAIL = "alice@local.test";

export default function WebmailPage() {
  const [selectedMessageId, setSelectedMessageId] = useState<number | null>(
    null,
  );
  const [composeOpen, setComposeOpen] = useState(false);
  const [composeInitialTo, setComposeInitialTo] = useState<string[]>([]);
  const [composeInitialSubject, setComposeInitialSubject] = useState("");

  // Search state
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<MailMessage[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const isSearching = searchQuery.length > 0;

  const { data: inboxData, isLoading: inboxLoading } = useInbox(
    CURRENT_USER_ID,
    { page: 1, limit: 50 },
  );

  const { data: messageDetail, isLoading: messageLoading } = useMessage(
    CURRENT_USER_ID,
    selectedMessageId ?? 0,
  );

  const messages: MailMessage[] = useMemo(
    () => inboxData?.data ?? [],
    [inboxData],
  );

  const handleSelectMessage = useCallback(
    (id: number) => {
      setSelectedMessageId(id);
      // Mark as read
      const allMessages = isSearching ? searchResults : messages;
      const msg = allMessages.find((m) => m.id === id);
      if (msg && !msg.read) {
        apiClient
          .markAsRead(CURRENT_USER_ID, id)
          .catch(() => {/* silent fail for now */});
      }
    },
    [messages, searchResults, isSearching],
  );

  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
    setSearchLoading(true);
    setSelectedMessageId(null);
    apiClient
      .searchMail(query, CURRENT_USER_EMAIL)
      .then((results) => {
        setSearchResults(results);
      })
      .catch(() => {
        setSearchResults([]);
      })
      .finally(() => {
        setSearchLoading(false);
      });
  }, []);

  const handleClearSearch = useCallback(() => {
    setSearchQuery("");
    setSearchResults([]);
    setSelectedMessageId(null);
  }, []);

  // Keyboard navigation
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // Don't capture if user is typing in an input
      const target = event.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.isContentEditable
      ) {
        return;
      }

      if (event.key === "j") {
        event.preventDefault();
        setSelectedMessageId((current) => {
          const currentIdx = messages.findIndex((m) => m.id === current);
          const nextIdx = Math.min(currentIdx + 1, messages.length - 1);
          return messages[nextIdx]?.id ?? current;
        });
      }

      if (event.key === "k") {
        event.preventDefault();
        setSelectedMessageId((current) => {
          const currentIdx = messages.findIndex((m) => m.id === current);
          const prevIdx = Math.max(currentIdx - 1, 0);
          return messages[prevIdx]?.id ?? current;
        });
      }

      if (event.key === "u" && selectedMessageId !== null) {
        event.preventDefault();
        apiClient
          .toggleRead(CURRENT_USER_ID, selectedMessageId)
          .catch(() => {/* silent fail */});
      }

      if (event.key === "Delete" && selectedMessageId !== null) {
        event.preventDefault();
        apiClient
          .trashMessage(CURRENT_USER_ID, selectedMessageId)
          .catch(() => {/* silent fail */});
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [messages, selectedMessageId]);

  const handleNewMessage = useCallback(() => {
    setComposeInitialTo([]);
    setComposeInitialSubject("");
    setComposeOpen(true);
  }, []);

  const handleReply = useCallback(() => {
    if (!messageDetail) return;
    setComposeInitialTo([messageDetail.from]);
    const subj = messageDetail.subject;
    setComposeInitialSubject(subj.startsWith("Re: ") ? subj : `Re: ${subj}`);
    setComposeOpen(true);
  }, [messageDetail]);

  const handleReplyAll = useCallback(() => {
    if (!messageDetail) return;
    const allRecipients = [
      messageDetail.from,
      ...messageDetail.to.filter((addr) => addr !== "alice@local.test"),
    ];
    const subj = messageDetail.subject;
    setComposeInitialTo(allRecipients);
    setComposeInitialSubject(subj.startsWith("Re: ") ? subj : `Re: ${subj}`);
    setComposeOpen(true);
  }, [messageDetail]);

  return (
    <div className="flex h-full flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Mail className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Webmail</h2>
        </div>
        <div className="flex items-center gap-3">
          <SearchBar
            onSearch={handleSearch}
            onClear={handleClearSearch}
            isSearching={isSearching}
          />
          <Button onClick={handleNewMessage} size="sm">
            <PenSquare className="size-4" />
            New Message
          </Button>
        </div>
      </div>

      {/* Three-pane layout */}
      <div className="flex flex-1 overflow-hidden rounded-lg border border-border">
        {/* Folder sidebar */}
        <div className="w-44 shrink-0 border-r border-border bg-card/50 p-3">
          <div className="flex items-center gap-2 rounded-md bg-accent px-3 py-1.5 text-sm font-medium text-foreground">
            <Inbox className="size-4" />
            <span>Inbox</span>
            {messages.filter((m) => !m.read).length > 0 && (
              <span className="ml-auto text-xs text-muted-foreground">
                {messages.filter((m) => !m.read).length}
              </span>
            )}
          </div>
        </div>

        {/* Message list / Search results */}
        <div className="w-96 shrink-0 border-r border-border bg-card">
          {isSearching ? (
            <SearchResults
              results={searchResults}
              query={searchQuery}
              selectedId={selectedMessageId}
              isLoading={searchLoading}
              onSelect={handleSelectMessage}
            />
          ) : (
            <MessageList
              messages={messages}
              selectedId={selectedMessageId}
              isLoading={inboxLoading}
              onSelect={handleSelectMessage}
            />
          )}
        </div>

        {/* Preview pane */}
        <div className="flex-1 bg-card">
          <div className="flex h-full flex-col">
            <MessagePreview
              message={messageDetail ?? null}
              isLoading={messageLoading && selectedMessageId !== null}
            />
            {messageDetail && (
              <div className="shrink-0 border-t border-border px-5 py-3">
                <div className="flex items-center gap-2">
                  <Button variant="outline" size="sm" onClick={handleReply}>
                    <Reply className="size-4" />
                    Reply
                  </Button>
                  <Button variant="outline" size="sm" onClick={handleReplyAll}>
                    <ReplyAll className="size-4" />
                    Reply All
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      <ComposeDialog
        open={composeOpen}
        onOpenChange={setComposeOpen}
        initialTo={composeInitialTo}
        initialSubject={composeInitialSubject}
      />
    </div>
  );
}
