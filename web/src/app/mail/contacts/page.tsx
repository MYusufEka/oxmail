"use client";

import { useState, useMemo, useCallback, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  BookUser,
  Check,
  ChevronsUpDown,
  Globe,
  Plus,
  Pencil,
  Trash2,
  Phone,
  Search,
  X,
  Mail,
  User,
  AlertTriangle,
} from "lucide-react";
import { toast } from "sonner";
import {
  useContacts,
  useCreateContact,
  useUpdateContact,
  useDeleteContact,
} from "@/hooks/use-contacts";
import { useDomains } from "@/hooks/use-domains";
import { useUsers } from "@/hooks/use-users";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cn } from "@/lib/utils";
import type { Contact } from "@/types/api";

export default function ContactsPage() {
  // Domain selector state
  const [selectedDomainId, setSelectedDomainId] = useState<number>(0);
  const [selectedDomainName, setSelectedDomainName] = useState<string>("");
  const [domainPopoverOpen, setDomainPopoverOpen] = useState(false);

  // User selector state
  const [userEmail, setUserEmail] = useState<string>("");
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

  // Default to first user of selected domain; reset when domain changes
  useEffect(() => {
    if (usersData?.data && usersData.data.length > 0) {
      setUserEmail(usersData.data[0].email);
    } else if (usersData?.data && usersData.data.length === 0) {
      setUserEmail("");
    }
  }, [usersData]);

  const { data: contacts, isLoading, error } = useContacts(userEmail);
  const createContact = useCreateContact(userEmail);
  const updateContact = useUpdateContact(userEmail);
  const deleteContact = useDeleteContact(userEmail);

  const [searchQuery, setSearchQuery] = useState("");

  // Add contact dialog
  const [addOpen, setAddOpen] = useState(false);
  const [addName, setAddName] = useState("");
  const [addEmail, setAddEmail] = useState("");
  const [addPhone, setAddPhone] = useState("");

  // Edit contact dialog
  const [editOpen, setEditOpen] = useState(false);
  const [editContact, setEditContact] = useState<Contact | null>(null);
  const [editName, setEditName] = useState("");
  const [editEmail, setEditEmail] = useState("");
  const [editPhone, setEditPhone] = useState("");

  // Delete contact dialog
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Contact | null>(null);

  const router = useRouter();

  const filtered = useMemo(() => {
    if (!contacts) return [];
    if (!searchQuery) return contacts;
    const q = searchQuery.toLowerCase();
    return contacts.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.email.toLowerCase().includes(q) ||
        (c.phone && c.phone.toLowerCase().includes(q)),
    );
  }, [contacts, searchQuery]);

  const handleAdd = useCallback(() => {
    if (!addName.trim() || !addEmail.trim()) {
      toast.error("Name and email are required");
      return;
    }
    createContact.mutate(
      {
        name: addName.trim(),
        email: addEmail.trim(),
        phone: addPhone.trim() || undefined,
      },
      {
        onSuccess: () => {
          toast.success("Contact added");
          setAddOpen(false);
          setAddName("");
          setAddEmail("");
          setAddPhone("");
        },
        onError: () => {
          toast.error("Failed to add contact");
        },
      },
    );
  }, [addName, addEmail, addPhone, createContact]);

  const handleEdit = useCallback(() => {
    if (!editContact || !editName.trim() || !editEmail.trim()) {
      toast.error("Name and email are required");
      return;
    }
    updateContact.mutate(
      {
        contactId: editContact.id,
        payload: {
          name: editName.trim(),
          email: editEmail.trim(),
          phone: editPhone.trim() || undefined,
        },
      },
      {
        onSuccess: () => {
          toast.success("Contact updated");
          setEditOpen(false);
          setEditContact(null);
        },
        onError: () => {
          toast.error("Failed to update contact");
        },
      },
    );
  }, [editContact, editName, editEmail, editPhone, updateContact]);

  const handleDeleteConfirm = useCallback(() => {
    if (!deleteTarget) return;
    deleteContact.mutate(deleteTarget.id, {
      onSuccess: () => {
        toast.success("Contact deleted");
        setDeleteOpen(false);
        setDeleteTarget(null);
      },
      onError: () => {
        toast.error("Failed to delete contact");
      },
    });
  }, [deleteTarget, deleteContact]);

  const openEdit = useCallback((contact: Contact) => {
    setEditContact(contact);
    setEditName(contact.name);
    setEditEmail(contact.email);
    setEditPhone(contact.phone ?? "");
    setEditOpen(true);
  }, []);

  const openDelete = useCallback((contact: Contact) => {
    setDeleteTarget(contact);
    setDeleteOpen(true);
  }, []);

  const handleComposeToContact = useCallback(
    (email: string) => {
      router.push(`/mail?composeTo=${encodeURIComponent(email)}`);
    },
    [router],
  );

  return (
    <div className="flex h-full flex-col gap-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <BookUser className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Contacts</h2>

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
                  data-testid="domain-selector"
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
                  title={userEmail}
                  data-testid="user-selector"
                  className="h-8 w-auto max-w-none shrink-0 justify-between gap-2"
                >
                  <User className="size-3.5 shrink-0 opacity-60" />
                  <span className="flex-1 truncate text-left text-xs font-medium">
                    {userEmail || "Select account..."}
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
                            setUserEmail(user.email);
                            setUserPopoverOpen(false);
                          }}
                        >
                          <Check
                            className={cn(
                              "mr-1.5 size-3.5 text-primary",
                              userEmail === user.email
                                ? "opacity-100"
                                : "opacity-0",
                            )}
                          />
                          {user.email}
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  </CommandList>
                </Command>
              </PopoverContent>
            </Popover>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {/* Search */}
          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search contacts..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-8 w-52 pl-8 pr-8 text-xs"
              data-testid="contacts-search"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery("")}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                aria-label="Clear search"
                data-testid="contacts-search-clear"
              >
                <X className="size-3.5" />
              </button>
            )}
          </div>

          <Button
            size="sm"
            className="h-8 gap-1.5 text-xs"
            onClick={() => setAddOpen(true)}
            disabled={!userEmail}
            data-testid="add-contact-button"
          >
            <Plus className="size-3.5" />
            Add Contact
          </Button>
        </div>
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="flex flex-col gap-2" data-testid="contacts-loading">
          {Array.from({ length: 6 }).map((_, idx) => (
            <div
              key={idx}
              className="flex items-center gap-4 rounded-md border border-border px-4 py-3"
            >
              <Skeleton className="size-8 rounded-full" />
              <div className="flex flex-1 flex-col gap-1.5">
                <Skeleton className="h-3.5 w-32" />
                <Skeleton className="h-3 w-48" />
              </div>
              <Skeleton className="h-7 w-20" />
              <Skeleton className="h-7 w-7" />
              <Skeleton className="h-7 w-7" />
            </div>
          ))}
        </div>
      ) : error ? (
        <div
          data-testid="contacts-error"
          className="flex h-full flex-col items-center justify-center gap-3 p-8"
        >
          <AlertTriangle className="size-10 text-destructive/60" />
          <p className="text-sm text-muted-foreground">Failed to load contacts</p>
          <p className="text-xs text-muted-foreground/70">
            {error instanceof Error ? error.message : "Unknown error"}
          </p>
        </div>
      ) : !userEmail ? (
        <div
          data-testid="contacts-no-user"
          className="flex h-full flex-col items-center justify-center gap-3 p-8"
        >
          <User className="size-10 text-muted-foreground/50" />
          <p className="text-sm text-muted-foreground">Select an account to view contacts</p>
        </div>
      ) : filtered.length === 0 ? (
        <div
          data-testid="contacts-empty"
          className="flex h-full flex-col items-center justify-center gap-3 p-8"
        >
          <BookUser className="size-10 text-muted-foreground/50" />
          <p className="text-sm text-muted-foreground">
            {searchQuery ? "No contacts match your search" : "No contacts yet"}
          </p>
          {!searchQuery && (
            <Button
              size="sm"
              variant="outline"
              className="mt-1 gap-1.5 text-xs"
              onClick={() => setAddOpen(true)}
              data-testid="add-contact-empty-button"
            >
              <Plus className="size-3.5" />
              Add your first contact
            </Button>
          )}
        </div>
      ) : (
        <div
          className="flex flex-col gap-1 overflow-y-auto"
          data-testid="contacts-list"
          role="list"
        >
          {filtered.map((contact) => (
            <div
              key={contact.id}
              role="listitem"
              data-testid="contact-row"
              className="flex items-center gap-4 rounded-md border border-border bg-card px-4 py-3 transition-colors hover:bg-accent/30"
            >
              {/* Avatar initial */}
              <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary">
                {contact.name.charAt(0).toUpperCase()}
              </div>

              {/* Name + email + phone */}
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-foreground">
                  {contact.name}
                </p>
                <p className="truncate text-xs text-muted-foreground">
                  {contact.email}
                </p>
                {contact.phone && (
                  <p className="flex items-center gap-1 truncate text-xs text-muted-foreground/70">
                    <Phone className="size-3" />
                    {contact.phone}
                  </p>
                )}
              </div>

              {/* Actions */}
              <div className="flex shrink-0 items-center gap-1">
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-7 gap-1.5 px-2 text-xs text-muted-foreground hover:text-foreground"
                  onClick={() => handleComposeToContact(contact.email)}
                  title={`Compose email to ${contact.email}`}
                  data-testid="contact-compose-button"
                >
                  <Mail className="size-3.5" />
                  Compose
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-7 text-muted-foreground hover:text-foreground"
                  onClick={() => openEdit(contact)}
                  title="Edit contact"
                  data-testid="contact-edit-button"
                >
                  <Pencil className="size-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-7 text-muted-foreground hover:text-destructive"
                  onClick={() => openDelete(contact)}
                  title="Delete contact"
                  data-testid="contact-delete-button"
                >
                  <Trash2 className="size-3.5" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Add Contact Dialog */}
      <Dialog open={addOpen} onOpenChange={setAddOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add Contact</DialogTitle>
            <DialogDescription>
              Add a new contact for {userEmail || "this account"}.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3 py-2">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="add-name" className="text-sm font-medium">
                Name <span className="text-destructive">*</span>
              </label>
              <Input
                id="add-name"
                placeholder="Full name"
                value={addName}
                onChange={(e) => setAddName(e.target.value)}
                data-testid="add-contact-name"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label htmlFor="add-email" className="text-sm font-medium">
                Email <span className="text-destructive">*</span>
              </label>
              <Input
                id="add-email"
                type="email"
                placeholder="email@example.com"
                value={addEmail}
                onChange={(e) => setAddEmail(e.target.value)}
                data-testid="add-contact-email"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label htmlFor="add-phone" className="text-sm font-medium">
                Phone
              </label>
              <Input
                id="add-phone"
                type="tel"
                placeholder="+1 555 000 0000"
                value={addPhone}
                onChange={(e) => setAddPhone(e.target.value)}
                data-testid="add-contact-phone"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setAddOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleAdd}
              disabled={createContact.isPending}
              data-testid="add-contact-submit"
            >
              {createContact.isPending ? "Adding..." : "Add Contact"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Contact Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Contact</DialogTitle>
            <DialogDescription>
              Update contact information for {editContact?.name}.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-3 py-2">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="edit-name" className="text-sm font-medium">
                Name <span className="text-destructive">*</span>
              </label>
              <Input
                id="edit-name"
                placeholder="Full name"
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                data-testid="edit-contact-name"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label htmlFor="edit-email" className="text-sm font-medium">
                Email <span className="text-destructive">*</span>
              </label>
              <Input
                id="edit-email"
                type="email"
                placeholder="email@example.com"
                value={editEmail}
                onChange={(e) => setEditEmail(e.target.value)}
                data-testid="edit-contact-email"
              />
            </div>
            <div className="flex flex-col gap-1.5">
              <label htmlFor="edit-phone" className="text-sm font-medium">
                Phone
              </label>
              <Input
                id="edit-phone"
                type="tel"
                placeholder="+1 555 000 0000"
                value={editPhone}
                onChange={(e) => setEditPhone(e.target.value)}
                data-testid="edit-contact-phone"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleEdit}
              disabled={updateContact.isPending}
              data-testid="edit-contact-submit"
            >
              {updateContact.isPending ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Contact Dialog */}
      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Contact</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-medium text-foreground">{deleteTarget?.name}</span>?
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDeleteConfirm}
              disabled={deleteContact.isPending}
              data-testid="delete-contact-confirm"
            >
              {deleteContact.isPending ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
