"use client";

import { useState } from "react";
import { Globe, Plus, AlertCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import type { Domain } from "@/types/api";
import { useDomains, useCreateDomain, useDeleteDomain } from "@/hooks/use-domains";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { DomainTable } from "./domain-table";
import { AddDomainDialog } from "./add-domain-dialog";
import { DeleteDomainDialog } from "./delete-domain-dialog";

function DomainTableSkeleton() {
  return (
    <div className="rounded-lg border border-border" data-testid="domain-table-skeleton">
      <div className="border-b border-border p-2">
        <Skeleton className="h-8 w-full" />
      </div>
      {Array.from({ length: 4 }).map((_, index) => (
        <div key={index} className="border-b border-border p-2 last:border-0">
          <Skeleton className="h-10 w-full" />
        </div>
      ))}
    </div>
  );
}

function DomainEmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <div
      className="flex flex-col items-center justify-center rounded-lg border border-border bg-card p-12"
      data-testid="domain-empty-state"
    >
      <div className="flex size-12 items-center justify-center rounded-full bg-muted">
        <Globe className="size-6 text-muted-foreground" />
      </div>
      <h3 className="mt-4 text-base font-medium text-foreground">
        No domains yet
      </h3>
      <p className="mt-1 text-sm text-muted-foreground">
        Add your first domain to start receiving mail.
      </p>
      <Button className="mt-6" onClick={onAdd} data-testid="empty-add-domain">
        <Plus className="size-4" />
        Add Domain
      </Button>
    </div>
  );
}

function DomainErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div
      className="flex flex-col items-center justify-center rounded-lg border border-destructive/30 bg-card p-12"
      data-testid="domain-error-state"
    >
      <div className="flex size-12 items-center justify-center rounded-full bg-destructive/10">
        <AlertCircle className="size-6 text-destructive" />
      </div>
      <h3 className="mt-4 text-base font-medium text-foreground">
        Failed to load domains
      </h3>
      <p className="mt-1 text-sm text-muted-foreground">
        Something went wrong. Please try again.
      </p>
      <Button variant="outline" className="mt-6" onClick={onRetry} data-testid="retry-domains">
        <RefreshCw className="size-4" />
        Retry
      </Button>
    </div>
  );
}

export default function DomainsPage() {
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Domain | null>(null);

  const { data, isLoading, isError, refetch } = useDomains();
  const createDomain = useCreateDomain();
  const deleteDomain = useDeleteDomain();

  const domains = data?.data ?? [];

  function handleCreateDomain(name: string) {
    createDomain.mutate(
      { name },
      {
        onSuccess: () => {
          setAddDialogOpen(false);
          toast.success(`Domain "${name}" added successfully`);
        },
        onError: (error) => {
          toast.error(error.message || "Failed to add domain");
        },
      }
    );
  }

  function handleDeleteDomain() {
    if (!deleteTarget) return;
    const domainName = deleteTarget.name;

    deleteDomain.mutate(deleteTarget.id, {
      onSuccess: () => {
        setDeleteTarget(null);
        toast.success(`Domain "${domainName}" deleted`);
      },
      onError: (error) => {
        toast.error(error.message || "Failed to delete domain");
      },
    });
  }

  return (
    <div className="flex flex-col gap-6" data-testid="domains-page">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Globe className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Domains</h2>
        </div>
        {!isLoading && !isError && domains.length > 0 && (
          <Button onClick={() => setAddDialogOpen(true)} data-testid="add-domain-button">
            <Plus className="size-4" />
            Add Domain
          </Button>
        )}
      </div>

      {isLoading && <DomainTableSkeleton />}

      {isError && <DomainErrorState onRetry={() => refetch()} />}

      {!isLoading && !isError && domains.length === 0 && (
        <DomainEmptyState onAdd={() => setAddDialogOpen(true)} />
      )}

      {!isLoading && !isError && domains.length > 0 && (
        <DomainTable domains={domains} onDelete={setDeleteTarget} />
      )}

      <AddDomainDialog
        open={addDialogOpen}
        onOpenChange={setAddDialogOpen}
        onSubmit={handleCreateDomain}
        isPending={createDomain.isPending}
      />

      <DeleteDomainDialog
        domain={deleteTarget}
        open={deleteTarget !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        onConfirm={handleDeleteDomain}
        isPending={deleteDomain.isPending}
      />
    </div>
  );
}
