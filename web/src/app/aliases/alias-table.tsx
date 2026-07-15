"use client";

import { useState } from "react";
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from "@tanstack/react-table";
import { ArrowLeftRight, Pencil, Trash2, X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

export interface GroupedAlias {
  sourceAddress: string;
  destinations: string[];
  ids: number[];
  active: boolean;
  createdAt: string;
  isCatchall: boolean;
}

interface AliasTableProps {
  aliases: GroupedAlias[];
  isLoading: boolean;
  onDelete: (alias: GroupedAlias) => void;
  onEdit: (alias: GroupedAlias) => void;
  onBulkDelete: (aliases: GroupedAlias[]) => Promise<void>;
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function AliasTable({ aliases, isLoading, onDelete, onEdit, onBulkDelete }: AliasTableProps) {
  const [rowSelection, setRowSelection] = useState<Record<string, boolean>>({});
  const [bulkDeleting, setBulkDeleting] = useState(false);
  const [bulkConfirming, setBulkConfirming] = useState(false);

  const columns: ColumnDef<GroupedAlias>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          data-testid="select-all"
          aria-label="Select all aliases"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          data-testid={`select-row-${row.original.sourceAddress}`}
          aria-label={`Select alias ${row.original.sourceAddress}`}
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "sourceAddress",
      header: "Source",
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <span className="font-medium text-foreground" data-testid="alias-source">
            {row.original.sourceAddress}
          </span>
          {row.original.isCatchall && (
            <Badge
              variant="outline"
              className="text-xs border-amber-500/50 text-amber-500"
              data-testid="catchall-badge"
            >
              catch-all
            </Badge>
          )}
        </div>
      ),
    },
    {
      id: "destinations",
      header: "Destinations",
      cell: ({ row }) => (
        <div className="flex flex-wrap gap-1" data-testid="alias-destinations">
          {row.original.destinations.map((dest) => (
            <Badge key={dest} variant="outline" className="mr-1">
              {dest}
            </Badge>
          ))}
        </div>
      ),
    },
    {
      accessorKey: "active",
      header: "Status",
      cell: ({ row }) => (
        <Badge
          variant={row.original.active ? "default" : "secondary"}
          className="text-xs"
          data-testid="alias-status"
        >
          {row.original.active ? "Active" : "Inactive"}
        </Badge>
      ),
    },
    {
      accessorKey: "createdAt",
      header: "Created",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {formatDate(row.original.createdAt)}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex justify-end gap-1">
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() => onEdit(row.original)}
            aria-label={`Edit alias ${row.original.sourceAddress}`}
            data-testid={`edit-alias-${row.original.sourceAddress}`}
          >
            <Pencil className="size-4 text-muted-foreground hover:text-foreground" />
          </Button>
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() => onDelete(row.original)}
            aria-label={`Delete alias ${row.original.sourceAddress}`}
            data-testid={`delete-alias-${row.original.sourceAddress}`}
          >
            <Trash2 className="size-4 text-muted-foreground hover:text-destructive" />
          </Button>
        </div>
      ),
    },
  ];

  const table = useReactTable({
    data: aliases,
    columns,
    getCoreRowModel: getCoreRowModel(),
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    getRowId: (row) => row.sourceAddress,
    state: { rowSelection },
  });

  const selectedRows = table.getFilteredSelectedRowModel().rows;
  const selectedCount = selectedRows.length;

  async function handleBulkDelete() {
    if (selectedCount === 0) return;
    const selectedAliases = selectedRows.map((row) => row.original);
    setBulkDeleting(true);
    try {
      await onBulkDelete(selectedAliases);
      table.resetRowSelection();
      setBulkConfirming(false);
    } finally {
      setBulkDeleting(false);
    }
  }

  if (isLoading) {
    return <AliasTableSkeleton />;
  }

  if (aliases.length === 0) {
    return (
      <div
        className="flex flex-col items-center justify-center rounded-lg border border-border bg-card p-12"
        data-testid="alias-empty-state"
      >
        <div className="flex size-12 items-center justify-center rounded-full bg-muted">
          <ArrowLeftRight className="size-6 text-muted-foreground" />
        </div>
        <h3 className="mt-4 text-base font-medium text-foreground">
          No aliases yet
        </h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Add an alias to forward mail to another address.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-2">
      {selectedCount > 0 && (
        <div
          className="flex items-center gap-3 rounded-lg border border-border bg-muted/50 px-4 py-2"
          data-testid="bulk-action-bar"
        >
          <span className="text-sm text-foreground">
            {selectedCount} {selectedCount === 1 ? "alias" : "aliases"} selected
          </span>
          {bulkConfirming ? (
            <div className="flex items-center gap-2">
              <Button
                variant="destructive"
                size="sm"
                onClick={handleBulkDelete}
                disabled={bulkDeleting}
                data-testid="bulk-delete-button"
              >
                {bulkDeleting ? "Deleting..." : `Delete ${selectedCount} selected`}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setBulkConfirming(false)}
                disabled={bulkDeleting}
              >
                Cancel
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <Button
                variant="destructive"
                size="sm"
                onClick={() => setBulkConfirming(true)}
                data-testid="bulk-delete-button"
              >
                <Trash2 className="size-4" />
                Delete {selectedCount} selected
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => table.resetRowSelection()}
              >
                <X className="size-4" />
                Clear selection
              </Button>
            </div>
          )}
        </div>
      )}
      <div className="rounded-lg border border-border" data-testid="alias-table">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => (
              <TableRow key={row.id} data-state={row.getIsSelected() ? "selected" : undefined}>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id}>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

function AliasTableSkeleton() {
  return (
    <div className="rounded-lg border border-border" data-testid="alias-table-skeleton">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-10"><Skeleton className="size-4" /></TableHead>
            <TableHead>Source</TableHead>
            <TableHead>Destinations</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 5 }).map((_, index) => (
            <TableRow key={`skeleton-${index}`}>
              <TableCell><Skeleton className="size-4" /></TableCell>
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
