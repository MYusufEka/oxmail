"use client";

import { useState } from "react";
import { Users, AlertCircle, RefreshCw } from "lucide-react";
import { useUsers } from "@/hooks/use-users";
import { useDomains } from "@/hooks/use-domains";
import { UserTable } from "@/app/users/user-table";
import { AddUserDialog } from "@/app/users/add-user-dialog";
import { DeleteUserDialog } from "@/app/users/delete-user-dialog";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import type { User } from "@/types/api";

export default function UsersPage() {
  const [selectedDomainId, setSelectedDomainId] = useState<number>(0);
  const [userToDelete, setUserToDelete] = useState<User | null>(null);

  const { data: domainsResponse, isLoading: domainsLoading } = useDomains();
  const domains = domainsResponse?.data ?? [];

  const activeDomainId = selectedDomainId > 0 ? selectedDomainId : (domains[0]?.id ?? 0);
  const { data: usersResponse, isLoading: usersLoading, isError, error, refetch } = useUsers(activeDomainId);
  const users = usersResponse?.data ?? [];

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

        <AddUserDialog selectedDomainId={activeDomainId} domains={domains} />
      </div>

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
          users={users}
          isLoading={usersLoading}
          onDelete={setUserToDelete}
        />
      )}

      <DeleteUserDialog
        user={userToDelete}
        open={userToDelete !== null}
        onOpenChange={(open) => {
          if (!open) setUserToDelete(null);
        }}
      />
    </div>
  );
}
