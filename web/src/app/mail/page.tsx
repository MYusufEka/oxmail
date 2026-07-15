"use client";

import { Suspense, useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  Archive,
  AlertTriangle,
  BookUser,
  Check,
  ChevronsUpDown,
  FileEdit,
  Filter,
  Globe,
  Inbox,
  Luggage,
  Mail,
  PenSquare,
  Reply,
  ReplyAll,
  Send,
  Trash2,
  User,
} from "lucide-react";
import { toast } from "sonner";
import {
  useMailFolders,
  useFolderMessages,
  useMessage,
  useCreateFolder,
  useDeleteFolder,
  useRenameFolder,
  useMoveMessage,
} from "@/hooks/use-mail";
import { useDomains } from "@/hooks/use-domains";
import { useUsers } from "@/hooks/use-users";
import { useContacts } from "@/hooks/use-contacts";
import { apiClient } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";
import { MessageList } from "./message-list";
import { MessagePreview } from "./message-preview";
import { ComposeDialog } from "./compose-dialog";
import { SearchBar } from "./search-bar";
import { SearchResults } from "./search-results";
import type { MailMessage } from "@/types/api";

const FOLDER_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  Inbox: Inbox,
  Sent: Send,
  Drafts: FileEdit,
  Trash: Trash2,
  Junk: AlertTriangle,
  Archive: Archive,
};

function getFolderIcon(name: string) {
  const baseName = name.split("/").pop() ?? name;
  return FOLDER_ICONS[baseName] ?? Mail;
}

const MAIN_FOLDERS = ["INBOX", "Sent", "Drafts"];
const MORE_FOLDERS = ["Trash", "Junk", "Archive"];
const SYSTEM_FOLDERS = new Set(["INBOX", "Sent", "Drafts", "Trash", "Junk", "Archive"]);

function WebmailPageInner() {
  const [selectedMessageId, setSelectedMessageId] = useState<number | null>(
    null,
  );
  const [selectedFolder, setSelectedFolder] = useState("INBOX");
  const [composeOpen, setComposeOpen] = useState(false);
  const [composeInitialTo, setComposeInitialTo] = useState<string[]>([]);
  const [composeInitialSubject, setComposeInitialSubject] = useState("");

  const searchParams = useSearchParams();
  const router = useRouter();

  useEffect(() => {
    const composeTo = searchParams.get("composeTo");
    if (composeTo) {
      setComposeInitialTo([composeTo]);
      setComposeInitialSubject("");
      setComposeOpen(true);
      router.replace("/mail");
    }
  }, [searchParams, router]);

  // Domain combobox state
  const [selectedDomainId, setSelectedDomainId] = useState<number>(0);
  const [selectedDomainName, setSelectedDomainName] = useState<string>("");
  const [domainPopoverOpen, setDomainPopoverOpen] = useState(false);

  // User combobox state
  const [currentUserId, setCurrentUserId] = useState<number>(0);
  const [currentUserEmail, setCurrentUserEmail] = useState<string>("");
  const [userPopoverOpen, setUserPopoverOpen] = useState(false);

  const { data: domainsData } = useDomains({ limit: 100 });
  const { data: usersData } = useUsers(selectedDomainId);

  // Default to first domain once loaded
  useEffect(() => {
    if (domainsData?.data && domainsData.data.length > 0 && selectedDomainId === 0) {
      const firstDomain = domainsData.data[0];
      setSelectedDomainId(firstDomain.id);
      setSelectedDomainName(firstDomain.name);
    }
  }, [domainsData, selectedDomainId]);

  // Default to first user of selected domain once loaded; reset when domain changes
  useEffect(() => {
    if (usersData?.data && usersData.data.length > 0) {
      const firstUser = usersData.data[0];
      setCurrentUserId(firstUser.id);
      setCurrentUserEmail(firstUser.email);
    } else if (usersData?.data && usersData.data.length === 0) {
      setCurrentUserId(0);
      setCurrentUserEmail("");
    }
  }, [usersData]);

  // Search state
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<MailMessage[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const isSearching = searchQuery.length > 0;

  const { data: foldersData } = useMailFolders(currentUserEmail);

  const { data: contactsData } = useContacts(currentUserEmail);

  const contactsMap: Record<string, string> = useMemo(() => {
    if (!contactsData) return {};
    const map: Record<string, string> = {};
    for (const c of contactsData) {
      map[c.email] = c.name;
    }
    return map;
  }, [contactsData]);

  const { data: folderData, isLoading: folderLoading } = useFolderMessages(
    selectedFolder,
    currentUserEmail,
    { page: 1, limit: 50 },
  );

  const { data: messageDetail, isLoading: messageLoading } = useMessage(
    currentUserId,
    selectedMessageId ?? 0,
    currentUserEmail,
  );

  const messages: MailMessage[] = useMemo(
    () => folderData?.messages ?? [],
    [folderData],
  );

  const allFolders = useMemo(() => {
    const apiFolders = foldersData?.folders ?? [];

    const mainFolderItems = MAIN_FOLDERS.map((name) => {
      const found = apiFolders.find((f) => f.name === name);
      return found ?? { name, delimiter: "/", unread: 0, total: 0 };
    });

    const moreFolderItems = MORE_FOLDERS.map((name) => {
      const found = apiFolders.find((f) => f.name === name);
      return found ?? { name, delimiter: "/", unread: 0, total: 0 };
    });

    const knownNames = new Set([...MAIN_FOLDERS, ...MORE_FOLDERS]);
    const otherFolders = apiFolders.filter((f) => !knownNames.has(f.name));

    return { main: mainFolderItems, more: moreFolderItems, other: otherFolders };
  }, [foldersData]);

  const handleSelectMessage = useCallback(
    (id: number) => {
      setSelectedMessageId(id);
      const allMessages = isSearching ? searchResults : messages;
      const msg = allMessages.find((m) => m.id === id);
      if (msg && !msg.read) {
        apiClient
          .markAsRead(currentUserId, id, currentUserEmail)
          .catch(() => {toast.error("Failed to mark as read");});
      }
    },
    [messages, searchResults, isSearching, currentUserId, currentUserEmail],
  );

  const handleSelectFolder = useCallback((folderName: string) => {
    setSelectedFolder(folderName);
    setSelectedMessageId(null);
  }, []);

  const createFolderMutation = useCreateFolder(currentUserEmail);
  const deleteFolderMutation = useDeleteFolder(currentUserEmail);
  const renameFolderMutation = useRenameFolder(currentUserEmail);
  const moveMsgMutation = useMoveMessage(currentUserEmail);

  const [folderDialogMode, setFolderDialogMode] = useState<"create" | "rename" | "delete" | null>(null);
  const [folderDialogTarget, setFolderDialogTarget] = useState<string>("");
  const [folderNameInput, setFolderNameInput] = useState<string>("");

  const openCreateDialog = useCallback(() => {
    setFolderNameInput("");
    setFolderDialogMode("create");
  }, []);

  const openRenameDialog = useCallback((folderName: string) => {
    setFolderDialogTarget(folderName);
    setFolderNameInput(folderName);
    setFolderDialogMode("rename");
  }, []);

  const openDeleteDialog = useCallback((folderName: string) => {
    setFolderDialogTarget(folderName);
    setFolderDialogMode("delete");
  }, []);

  const closeFolderDialog = useCallback(() => {
    setFolderDialogMode(null);
    setFolderNameInput("");
    setFolderDialogTarget("");
  }, []);

  const handleCreateFolderSubmit = useCallback(() => {
    if (!folderNameInput.trim()) return;
    createFolderMutation.mutate(folderNameInput.trim(), {
      onSuccess: () => {
        toast.success(`Folder "${folderNameInput.trim()}" created`);
        closeFolderDialog();
      },
      onError: () => toast.error("Failed to create folder"),
    });
  }, [createFolderMutation, folderNameInput, closeFolderDialog]);

  const handleRenameFolderSubmit = useCallback(() => {
    if (!folderNameInput.trim() || !folderDialogTarget) return;
    renameFolderMutation.mutate(
      { oldName: folderDialogTarget, newName: folderNameInput.trim() },
      {
        onSuccess: () => {
          if (selectedFolder === folderDialogTarget) {
            setSelectedFolder(folderNameInput.trim());
          }
          toast.success(`Folder renamed to "${folderNameInput.trim()}"`);
          closeFolderDialog();
        },
        onError: () => toast.error("Failed to rename folder"),
      },
    );
  }, [renameFolderMutation, folderNameInput, folderDialogTarget, selectedFolder, closeFolderDialog]);

  const handleDeleteFolderSubmit = useCallback(() => {
    if (!folderDialogTarget) return;
    deleteFolderMutation.mutate(folderDialogTarget, {
      onSuccess: () => {
        if (selectedFolder === folderDialogTarget) {
          setSelectedFolder("INBOX");
          setSelectedMessageId(null);
        }
        toast.success(`Folder "${folderDialogTarget}" deleted`);
        closeFolderDialog();
      },
      onError: () => toast.error("Failed to delete folder"),
    });
  }, [deleteFolderMutation, folderDialogTarget, selectedFolder, closeFolderDialog]);

  const handleSearch = useCallback(
    (query: string) => {
      setSearchQuery(query);
      setSearchLoading(true);
      setSelectedMessageId(null);
      apiClient
        .searchMail(query, currentUserEmail)
        .then((results) => {
          setSearchResults(results);
        })
        .catch(() => {
          setSearchResults([]);
        })
        .finally(() => {
          setSearchLoading(false);
        });
    },
    [currentUserEmail],
  );

  const handleClearSearch = useCallback(() => {
    setSearchQuery("");
    setSearchResults([]);
    setSelectedMessageId(null);
  }, []);

  // Keyboard navigation
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
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
          .toggleRead(currentUserId, selectedMessageId, currentUserEmail)
          .catch(() => {toast.error("Failed to toggle read");});
      }

      if (event.key === "Delete" && selectedMessageId !== null) {
        event.preventDefault();
        apiClient
          .trashMessage(currentUserId, selectedMessageId, currentUserEmail)
          .catch(() => {toast.error("Failed to delete message");});
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [messages, selectedMessageId, currentUserId, currentUserEmail]);

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
      ...messageDetail.to.filter((addr) => addr !== currentUserEmail),
    ];
    const subj = messageDetail.subject;
    setComposeInitialTo(allRecipients);
    setComposeInitialSubject(subj.startsWith("Re: ") ? subj : `Re: ${subj}`);
    setComposeOpen(true);
  }, [messageDetail, currentUserEmail]);

  return (
    <div className="flex h-full flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Mail className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Webmail</h2>

          <div className="h-4 w-px bg-border opacity-40" />

          {/* Domain + User comboboxes */}
          <div className="flex items-center gap-1.5">
            {/* Domain combobox */}
            <Popover open={domainPopoverOpen} onOpenChange={setDomainPopoverOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={domainPopoverOpen}
                  title={selectedDomainName}
                  className="h-8 w-auto max-w-none shrink-0 justify-between gap-2"
                >
                  <Globe className="size-3.5 shrink-0 opacity-60" />
                  <span className="flex-1 truncate text-left text-xs font-medium">
                    {selectedDomainName || "Select domain..."}
                  </span>
                  <ChevronsUpDown className="ml-auto size-3.5 shrink-0 opacity-40" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto min-w-[200px] border border-border/50 p-0 shadow-lg shadow-black/20 backdrop-blur-sm">
                <Command>
                  <CommandInput
                    placeholder="Search domain..."
                    className="text-xs placeholder:text-xs"
                  />
                  <CommandList>
                    <CommandEmpty>No domains.</CommandEmpty>
                    <CommandGroup>
                      {(domainsData?.data ?? []).map((domain) => (
                        <CommandItem
                          key={domain.id}
                          value={domain.name}
                          className="py-1.5 px-2 text-xs"
                          onSelect={() => {
                            setSelectedDomainId(domain.id);
                            setSelectedDomainName(domain.name);
                            setDomainPopoverOpen(false);
                            setSelectedMessageId(null);
                          }}
                        >
                          <Check
                            className={cn(
                              "mr-1.5 size-3.5 text-primary",
                              selectedDomainId === domain.id
                                ? "opacity-100"
                                : "opacity-0",
                            )}
                          />
                          {domain.name}
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>

            {/* User combobox */}
            <Popover open={userPopoverOpen} onOpenChange={setUserPopoverOpen}>
              <PopoverTrigger asChild>
                <Button
                  variant="outline"
                  role="combobox"
                  aria-expanded={userPopoverOpen}
                  title={currentUserEmail}
                  className="h-8 w-auto max-w-none shrink-0 justify-between gap-2"
                >
                  <User className="size-3.5 shrink-0 opacity-60" />
                  <span className="flex-1 truncate text-left text-xs font-medium">
                    {currentUserEmail || "Select account..."}
                  </span>
                  <ChevronsUpDown className="ml-auto size-3.5 shrink-0 opacity-40" />
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-auto min-w-[200px] border border-border/50 p-0 shadow-lg shadow-black/20 backdrop-blur-sm">
                <Command>
                  <CommandInput
                    placeholder="Search account..."
                    className="text-xs placeholder:text-xs"
                  />
                  <CommandList>
                    <CommandEmpty>No accounts.</CommandEmpty>
                    <CommandGroup>
                      {(usersData?.data ?? []).map((user) => (
                        <CommandItem
                          key={user.id}
                          value={user.email}
                          className="py-1.5 px-2 text-xs"
                          onSelect={() => {
                            setCurrentUserId(user.id);
                            setCurrentUserEmail(user.email);
                            setUserPopoverOpen(false);
                            setSelectedMessageId(null);
                          }}
                        >
                          <Check
                            className={cn(
                              "mr-1.5 size-3.5 text-primary",
                              currentUserId === user.id
                                ? "opacity-100"
                                : "opacity-0",
                            )}
                          />
                          <span className="truncate">{user.email}</span>
                          {user.displayName && (
                            <span className="ml-1 truncate text-xs text-muted-foreground">
                              {user.displayName}
                            </span>
                          )}
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>
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
        <div className="flex w-44 shrink-0 flex-col border-r border-border bg-card/50 p-3">
          {/* Filters + Vacation + Contacts nav */}
          <div className="flex flex-col gap-0.5">
            <Link
              href="/mail/filters"
              className="flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground"
            >
              <Filter className="size-4 shrink-0" />
              <span className="flex-1 text-left">Filters</span>
            </Link>
            <Link
              href="/mail/vacation"
              className="flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground"
            >
              <Luggage className="size-4 shrink-0" />
              <span className="flex-1 text-left">Vacation</span>
            </Link>
            <Link
              href="/mail/contacts"
              className="flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors hover:bg-accent/50 hover:text-foreground"
            >
              <BookUser className="size-4 shrink-0" />
              <span className="flex-1 text-left">Contacts</span>
            </Link>
          </div>

          <div className="my-2 h-px bg-border opacity-40" />

          <div className={cn("flex flex-col gap-0.5", !currentUserEmail && "pointer-events-none opacity-50")}>
            {allFolders.main.map((folder) => {
              const Icon = getFolderIcon(folder.name);
              const isActive = selectedFolder === folder.name;
              const isSystem = SYSTEM_FOLDERS.has(folder.name);
              return (
                <DropdownMenu key={folder.name}>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      data-testid={`folder-${folder.name}`}
                      onClick={() => handleSelectFolder(folder.name)}
                      onContextMenu={(e) => e.preventDefault()}
                      className={cn(
                        "flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                        isActive
                          ? "bg-accent text-foreground"
                          : "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
                      )}
                    >
                      <Icon className="size-4 shrink-0" />
                      <span className="flex-1 text-left">{folder.name}</span>
                      {folder.unread > 0 && (
                        <span className="ml-auto text-xs tabular-nums text-muted-foreground">
                          {folder.unread}
                        </span>
                      )}
                    </button>
                  </DropdownMenuTrigger>
                  {currentUserEmail && (
                    <DropdownMenuContent side="right" align="start">
                      <DropdownMenuItem onSelect={() => openCreateDialog()}>
                        New subfolder
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        disabled={isSystem}
                        onSelect={() => openRenameDialog(folder.name)}
                      >
                        Rename
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        disabled={isSystem}
                        className="text-destructive focus:text-destructive"
                        onSelect={() => openDeleteDialog(folder.name)}
                      >
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  )}
                </DropdownMenu>
              );
            })}
          </div>

          <div className="my-2 h-px bg-border opacity-40" />

          <div className={cn("flex flex-col gap-0.5", !currentUserEmail && "pointer-events-none opacity-50")}>
            {allFolders.more.map((folder) => {
              const Icon = getFolderIcon(folder.name);
              const isActive = selectedFolder === folder.name;
              const isSystem = SYSTEM_FOLDERS.has(folder.name);
              return (
                <DropdownMenu key={folder.name}>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      data-testid={`folder-${folder.name}`}
                      onClick={() => handleSelectFolder(folder.name)}
                      onContextMenu={(e) => e.preventDefault()}
                      className={cn(
                        "flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                        isActive
                          ? "bg-accent text-foreground"
                          : "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
                      )}
                    >
                      <Icon className="size-4 shrink-0" />
                      <span className="flex-1 text-left">{folder.name}</span>
                      {folder.unread > 0 && (
                        <span className="ml-auto text-xs tabular-nums text-muted-foreground">
                          {folder.unread}
                        </span>
                      )}
                    </button>
                  </DropdownMenuTrigger>
                  {currentUserEmail && (
                    <DropdownMenuContent side="right" align="start">
                      <DropdownMenuItem onSelect={() => openCreateDialog()}>
                        New subfolder
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        disabled={isSystem}
                        onSelect={() => openRenameDialog(folder.name)}
                      >
                        Rename
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        disabled={isSystem}
                        className="text-destructive focus:text-destructive"
                        onSelect={() => openDeleteDialog(folder.name)}
                      >
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  )}
                </DropdownMenu>
              );
            })}

            {allFolders.other.map((folder) => {
              const Icon = getFolderIcon(folder.name);
              const isActive = selectedFolder === folder.name;
              return (
                <DropdownMenu key={folder.name}>
                  <DropdownMenuTrigger asChild>
                    <button
                      type="button"
                      data-testid={`folder-${folder.name}`}
                      onClick={() => handleSelectFolder(folder.name)}
                      onContextMenu={(e) => e.preventDefault()}
                      className={cn(
                        "flex items-center gap-2 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                        isActive
                          ? "bg-accent text-foreground"
                          : "text-muted-foreground hover:bg-accent/50 hover:text-foreground",
                      )}
                    >
                      <Icon className="size-4 shrink-0" />
                      <span className="flex-1 text-left">{folder.name}</span>
                      {folder.unread > 0 && (
                        <span className="ml-auto text-xs tabular-nums text-muted-foreground">
                          {folder.unread}
                        </span>
                      )}
                    </button>
                  </DropdownMenuTrigger>
                  {currentUserEmail && (
                    <DropdownMenuContent side="right" align="start">
                      <DropdownMenuItem onSelect={() => openCreateDialog()}>
                        New subfolder
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onSelect={() => openRenameDialog(folder.name)}>
                        Rename
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        className="text-destructive focus:text-destructive"
                        onSelect={() => openDeleteDialog(folder.name)}
                      >
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  )}
                </DropdownMenu>
              );
            })}
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
              contactsMap={contactsMap}
              onSelect={handleSelectMessage}
            />
          ) : (
            <MessageList
              messages={messages}
              selectedId={selectedMessageId}
              isLoading={folderLoading}
              contactsMap={contactsMap}
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
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="outline"
                        size="sm"
                        data-testid="move-message-trigger"
                        disabled={moveMsgMutation.isPending}
                      >
                        Move to...
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="start">
                      {(foldersData?.folders ?? [])
                        .filter((f) => f.name !== selectedFolder)
                        .map((f) => (
                          <DropdownMenuItem
                            key={f.name}
                            onSelect={() => {
                              moveMsgMutation.mutate(
                                {
                                  uid: messageDetail.id,
                                  fromFolder: selectedFolder,
                                  toFolder: f.name,
                                },
                                {
                                  onSuccess: () => {
                                    setSelectedMessageId(null);
                                    toast.success(`Moved to ${f.name}`);
                                  },
                                  onError: () => toast.error("Failed to move message"),
                                },
                              );
                            }}
                          >
                            {f.name}
                          </DropdownMenuItem>
                        ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
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
        currentUserEmail={currentUserEmail}
      />

      <Dialog
        open={folderDialogMode === "create"}
        onOpenChange={(open) => { if (!open) closeFolderDialog(); }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Subfolder</DialogTitle>
            <DialogDescription>Enter a name for the new folder.</DialogDescription>
          </DialogHeader>
          <Input
            data-testid="folder-name-input"
            value={folderNameInput}
            onChange={(e) => setFolderNameInput(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleCreateFolderSubmit(); }}
            placeholder="Folder name"
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={closeFolderDialog}>Cancel</Button>
            <Button
              onClick={handleCreateFolderSubmit}
              disabled={!folderNameInput.trim() || createFolderMutation.isPending}
              data-testid="folder-create-confirm"
            >
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={folderDialogMode === "rename"}
        onOpenChange={(open) => { if (!open) closeFolderDialog(); }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename Folder</DialogTitle>
            <DialogDescription>Enter a new name for &quot;{folderDialogTarget}&quot;.</DialogDescription>
          </DialogHeader>
          <Input
            data-testid="folder-rename-input"
            value={folderNameInput}
            onChange={(e) => setFolderNameInput(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleRenameFolderSubmit(); }}
            placeholder="New folder name"
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={closeFolderDialog}>Cancel</Button>
            <Button
              onClick={handleRenameFolderSubmit}
              disabled={!folderNameInput.trim() || renameFolderMutation.isPending}
              data-testid="folder-rename-confirm"
            >
              Rename
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={folderDialogMode === "delete"}
        onOpenChange={(open) => { if (!open) closeFolderDialog(); }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Folder</DialogTitle>
            <DialogDescription>
              This will delete &quot;{folderDialogTarget}&quot; and all messages in it. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={closeFolderDialog}>Cancel</Button>
            <Button
              variant="destructive"
              onClick={handleDeleteFolderSubmit}
              disabled={deleteFolderMutation.isPending}
              data-testid="folder-delete-confirm"
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default function WebmailPage() {
  return (
    <Suspense fallback={<div className="h-full w-full" />}>
      <WebmailPageInner />
    </Suspense>
  );
}
