"use client";

import { useState, useMemo, useCallback } from "react";
import { Search, Users, AlertCircle, AlertTriangle, RefreshCw, X } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { useUsers } from "@/hooks/use-users";
import { useDomains } from "@/hooks/use-domains";
import { UserTable } from "@/app/users/user-table";
import { AddUserDialog } from "@/app/users/add-user-dialog";
import { EditUserDialog } from "@/app/users/edit-user-dialog";
import { DeleteUserDialog } from "@/app/users/delete-user-dialog";
import { MailSetupDialog } from "@/app/users/mail-setup-dialog";
import { ResetPasswordDialog } from "@/app/users/reset-password-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import type { User } from "@/types/api";

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const kb = bytes / 1024;
  if (kb < 1) return `${bytes} B`;
  const mb = kb / 1024;
  if (mb < 1) return `${Math.round(kb)} KB`;
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`;
  return `${Math.round(mb)} MB`;
}

export default function UsersPage() {
  const queryClient = useQueryClient();
  const [selectedDomainId, setSelectedDomainId] = useState<number>(0);
  const [userToDelete, setUserToDelete] = useState<User | null>(null);
  const [userToEdit, setUserToEdit] = useState<User | null>(null);
  const [userMailSetup, setUserMailSetup] = useState<User | null>(null);
  const [userToReset, setUserToReset] = useState<User | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [quotaDismissed, setQuotaDismissed] = useState(false);

  const { data: domainsResponse, isLoading: domainsLoading } = useDomains();
  const domains = domainsResponse?.data ?? [];

  const activeDomainId = selectedDomainId > 0 ? selectedDomainId : (domains[0]?.id ?? 0);
  const { data: usersResponse, isLoading: usersLoading, isError, error, refetch } = useUsers(activeDomainId);

  const handleBulkDelete = useCallback(
    async (selectedUsers: User[]) => {
      const results = await Promise.allSettled(
        selectedUsers.map((user) => apiClient.deleteUser(activeDomainId, user.id)),
      );
      await queryClient.invalidateQueries({ queryKey: ["users", activeDomainId] });

      const succeeded = results.filter((r) => r.status === "fulfilled").length;
      const failed = results.filter((r) => r.status === "rejected").length;

      if (failed === 0) {
        toast.success(`${succeeded} ${succeeded === 1 ? "user" : "users"} deleted`);
      } else {
        toast.error(`Deleted ${succeeded}/${results.length} users`);
      }
    },
    [activeDomainId, queryClient],
  );
  const users = usersResponse?.data ?? [];

  const filteredUsers = useMemo(() => {
    if (!searchQuery) return users;
    const q = searchQuery.toLowerCase();
    return users.filter(
      (u) =>
        u.email.toLowerCase().includes(q) ||
        (u.displayName && u.displayName.toLowerCase().includes(q)),
    );
  }, [users, searchQuery]);

  const usersNearQuota = useMemo(
    () =>
      users.filter(
        (u) =>
          u.quota > 0 &&
          u.storageUsed !== undefined &&
          u.storageUsed > 0 &&
          u.storageUsed / u.quota > 0.8,
      ),
    [users],
  );

  if (domainsLoading) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-3">
          <Users className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Users</h2>
        </div>
        <div className="flex items-center gap-3">
          <Skeleton className="h-9 w-48" />
          <Skeleton className="h-9 w-24" />
        </div>
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
    );
  }

  if (domains.length === 0) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-3">
          <Users className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Users</h2>
        </div>
        <div className="rounded-lg border border-border bg-card p-12 text-center">
          <p className="text-sm text-muted-foreground">
            No domains configured. Add a domain first before creating users.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <Users className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Users</h2>
      </div>

      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Select
            value={String(activeDomainId)}
            onValueChange={(value) => setSelectedDomainId(Number(value))}
          >
            <SelectTrigger className="w-56">
              <SelectValue placeholder="Select domain" />
            </SelectTrigger>
            <SelectContent>
              {domains.map((domain) => (
                <SelectItem key={domain.id} value={String(domain.id)}>
                  {domain.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="relative">
            <Search className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search users..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-9 w-56 pl-8 text-sm"
            />
          </div>
        </div>

        <AddUserDialog selectedDomainId={activeDomainId} domains={domains} />
      </div>

      {!quotaDismissed && usersNearQuota.length > 0 && (
        <div
          className="flex items-start gap-3 rounded-lg border border-amber-500/30 bg-amber-500/10 p-4"
          data-testid="quota-warning-banner"
        >
          <AlertTriangle className="mt-0.5 size-5 shrink-0 text-amber-500" />
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-amber-400">
              Storage Warning: {usersNearQuota.length}{" "}
              {usersNearQuota.length === 1 ? "user is" : "users are"} near
              their quota limit
            </p>
            <ul className="mt-1.5 space-y-0.5">
              {usersNearQuota.map((u) => {
                const used = u.storageUsed!;
                const pct = Math.round((used / u.quota) * 100);
                return (
                  <li
                    key={u.id}
                    className="text-xs text-amber-300/80"
                  >
                    {u.email} — {formatBytes(used)} / {formatBytes(u.quota)} ({pct}%)
                  </li>
                );
              })}
            </ul>
          </div>
          <Button
            variant="ghost"
            size="icon"
            className="size-6 shrink-0 text-amber-400 hover:text-amber-300"
            onClick={() => setQuotaDismissed(true)}
            data-testid="dismiss-quota-warning"
          >
            <X className="size-4" />
          </Button>
        </div>
      )}

      {isError ? (
        <div
          className="flex flex-col items-center justify-center rounded-lg border border-destructive/30 bg-card p-12"
          data-testid="users-error-state"
        >
          <div className="flex size-12 items-center justify-center rounded-full bg-destructive/10">
            <AlertCircle className="size-6 text-destructive" />
          </div>
          <h3 className="mt-4 text-base font-medium text-foreground">
            Failed to load users
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {error?.message ?? "Something went wrong. Please try again."}
          </p>
          <Button variant="outline" className="mt-6" onClick={() => refetch()} data-testid="retry-users">
            <RefreshCw className="size-4" />
            Retry
          </Button>
        </div>
      ) : (
        <UserTable
          users={filteredUsers}
          isLoading={usersLoading}
          onDelete={setUserToDelete}
          onEdit={setUserToEdit}
          onMailSetup={setUserMailSetup}
          onResetPassword={setUserToReset}
          onBulkDelete={handleBulkDelete}
        />
      )}

      <DeleteUserDialog
        user={userToDelete}
        open={userToDelete !== null}
        onOpenChange={(open) => {
          if (!open) setUserToDelete(null);
        }}
      />

      <EditUserDialog
        user={userToEdit}
        open={userToEdit !== null}
        onOpenChange={(open) => {
          if (!open) setUserToEdit(null);
        }}
      />

      <MailSetupDialog
        user={userMailSetup}
        open={userMailSetup !== null}
        onOpenChange={(open) => {
          if (!open) setUserMailSetup(null);
        }}
      />

      <ResetPasswordDialog
        user={userToReset}
        open={userToReset !== null}
        onOpenChange={(open) => {
          if (!open) setUserToReset(null);
        }}
        onSuccess={() => refetch()}
      />
    </div>
  );
}
