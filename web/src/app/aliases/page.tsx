"use client";

import { useState, useMemo } from "react";
import { ArrowLeftRight, Globe, Plus, AlertCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import type { Alias } from "@/types/api";
import { useAliases } from "@/hooks/use-aliases";
import { useDomains } from "@/hooks/use-domains";
import { useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { AliasTable, type GroupedAlias } from "./alias-table";
import { AddAliasDialog } from "./add-alias-dialog";
import { DeleteAliasDialog } from "./delete-alias-dialog";
import { EditAliasDialog } from "./edit-alias-dialog";

function groupAliases(aliases: Alias[]): GroupedAlias[] {
  const groups = new Map<
    string,
    { destinations: string[]; ids: number[]; active: boolean; createdAt: string }
  >();

  for (const alias of aliases) {
    const existing = groups.get(alias.sourceAddress);
    if (existing) {
      existing.destinations.push(alias.destinationAddress);
      existing.ids.push(alias.id);
      if (alias.active) existing.active = true;
      if (alias.createdAt < existing.createdAt) existing.createdAt = alias.createdAt;
    } else {
      groups.set(alias.sourceAddress, {
        destinations: [alias.destinationAddress],
        ids: [alias.id],
        active: alias.active,
        createdAt: alias.createdAt,
      });
    }
  }

  return Array.from(groups.entries())
    .map(([sourceAddress, data]) => ({
      sourceAddress,
      ...data,
      isCatchall: sourceAddress.startsWith("@"),
    }))
    .sort((a, b) => a.sourceAddress.localeCompare(b.sourceAddress));
}

function AliasPageSkeleton() {
  return (
    <div className="rounded-lg border border-border" data-testid="alias-page-skeleton">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Source</TableHead>
            <TableHead>Destinations</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 4 }).map((_, index) => (
            <TableRow key={`skeleton-${index}`}>
              <TableCell><Skeleton className="h-4 w-40" /></TableCell>
              <TableCell><Skeleton className="h-4 w-48" /></TableCell>
              <TableCell><Skeleton className="h-5 w-16 rounded-full" /></TableCell>
              <TableCell><Skeleton className="h-4 w-24" /></TableCell>
              <TableCell><Skeleton className="h-4 w-6" /></TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

export default function AliasesPage() {
  const queryClient = useQueryClient();
  const [selectedDomainId, setSelectedDomainId] = useState<number>(0);
  const [aliasToDelete, setAliasToDelete] = useState<GroupedAlias | null>(null);
  const [aliasToEdit, setAliasToEdit] = useState<GroupedAlias | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { data: domainsResponse, isLoading: domainsLoading } = useDomains();
  const domains = domainsResponse?.data ?? [];

  const activeDomainId =
    selectedDomainId > 0 ? selectedDomainId : (domains[0]?.id ?? 0);
  const {
    data: aliasesResponse,
    isLoading: aliasesLoading,
    isError,
    error,
    refetch,
  } = useAliases(activeDomainId);
  const aliases = aliasesResponse?.data ?? [];

  const groupedAliases = useMemo(() => groupAliases(aliases), [aliases]);

  async function handleDeleteConfirm() {
    if (!aliasToDelete) return;

    const { ids, sourceAddress } = aliasToDelete;
    setIsDeleting(true);

    const results = await Promise.allSettled(
      ids.map((id) => apiClient.deleteAlias(activeDomainId, id)),
    );

    queryClient.invalidateQueries({ queryKey: ["aliases", activeDomainId] });

    const succeeded = results.filter(
      (r) => r.status === "fulfilled",
    ).length;
    const failed = results.filter((r) => r.status === "rejected").length;

    if (failed === 0) {
      toast.success("Alias deleted", {
        description: `${sourceAddress} → all destinations removed.`,
      });
    } else {
      toast.error(
        `Deleted ${succeeded}/${results.length} destinations for ${sourceAddress}`,
      );
    }

    setIsDeleting(false);
    setAliasToDelete(null);
  }

  async function handleBulkDelete(selectedAliases: GroupedAlias[]) {
    const allIds = selectedAliases.flatMap((alias) => alias.ids);
    const results = await Promise.allSettled(
      allIds.map((id) => apiClient.deleteAlias(activeDomainId, id)),
    );
    await queryClient.invalidateQueries({ queryKey: ["aliases", activeDomainId] });

    const succeeded = results.filter((r) => r.status === "fulfilled").length;
    const failed = results.filter((r) => r.status === "rejected").length;

    if (failed === 0) {
      toast.success(`${selectedAliases.length} ${selectedAliases.length === 1 ? "alias" : "aliases"} deleted`);
    } else {
      toast.error(`Deleted ${succeeded}/${allIds.length} alias destinations`);
    }
  }

  if (domainsLoading) {
    return (
      <div className="flex flex-col gap-4" data-testid="aliases-loading">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <ArrowLeftRight className="size-5 text-primary" />
            <h2 className="text-lg font-semibold text-foreground">Aliases</h2>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Skeleton className="h-9 w-48" />
          <Skeleton className="h-9 w-24" />
        </div>
        <AliasPageSkeleton />
      </div>
    );
  }

  if (domains.length === 0) {
    return (
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-3">
          <ArrowLeftRight className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Aliases</h2>
        </div>
        <div
          className="flex flex-col items-center justify-center rounded-lg border border-border bg-card p-12"
          data-testid="aliases-no-domains"
        >
          <div className="flex size-12 items-center justify-center rounded-full bg-muted">
            <Globe className="size-6 text-muted-foreground" />
          </div>
          <h3 className="mt-4 text-base font-medium text-foreground">
            No domains configured
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Add a domain first before creating aliases.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <ArrowLeftRight className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Aliases</h2>
      </div>

      <div className="flex items-center gap-3">
        <Select
          value={String(activeDomainId)}
          onValueChange={(value) => setSelectedDomainId(Number(value))}
        >
          <SelectTrigger className="w-48" data-testid="domain-selector">
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

        <AddAliasDialog domainId={activeDomainId} />
      </div>

      {isError ? (
        <div
          className="flex flex-col items-center justify-center rounded-lg border border-destructive/30 bg-card p-12"
          data-testid="aliases-error-state"
        >
          <div className="flex size-12 items-center justify-center rounded-full bg-destructive/10">
            <AlertCircle className="size-6 text-destructive" />
          </div>
          <h3 className="mt-4 text-base font-medium text-foreground">
            Failed to load aliases
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            {error?.message ?? "Something went wrong. Please try again."}
          </p>
          <Button
            variant="outline"
            className="mt-6"
            onClick={() => refetch()}
            data-testid="retry-aliases"
          >
            <RefreshCw className="size-4" />
            Retry
          </Button>
        </div>
      ) : (
        <AliasTable
          aliases={groupedAliases}
          isLoading={aliasesLoading}
          onDelete={setAliasToDelete}
          onEdit={setAliasToEdit}
          onBulkDelete={handleBulkDelete}
        />
      )}

      <DeleteAliasDialog
        sourceAddress={aliasToDelete?.sourceAddress ?? ""}
        destinations={aliasToDelete?.destinations ?? []}
        open={aliasToDelete !== null}
        onOpenChange={(open) => {
          if (!open) setAliasToDelete(null);
        }}
        onConfirm={handleDeleteConfirm}
        isPending={isDeleting}
      />

      <EditAliasDialog
        alias={aliasToEdit}
        domainId={activeDomainId}
        open={aliasToEdit !== null}
        onOpenChange={(open) => {
          if (!open) setAliasToEdit(null);
        }}
      />
    </div>
  );
}
