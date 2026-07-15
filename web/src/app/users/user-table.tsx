"use client";

import { useState } from "react";
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from "@tanstack/react-table";
import { Trash2, Settings, Pencil, X, KeyRound } from "lucide-react";
import type { User } from "@/types/api";
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

interface UserTableProps {
  users: User[];
  isLoading: boolean;
  onDelete: (user: User) => void;
  onEdit: (user: User) => void;
  onMailSetup: (user: User) => void;
  onResetPassword: (user: User) => void;
  onBulkDelete: (users: User[]) => Promise<void>;
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function formatQuota(bytes: number): string {
  if (bytes === 0) return "Unlimited";
  const mb = bytes / (1024 * 1024);
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`;
  return `${Math.round(mb)} MB`;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const kb = bytes / 1024;
  if (kb < 1) return `${bytes} B`;
  const mb = kb / 1024;
  if (mb < 1) return `${Math.round(kb)} KB`;
  if (mb >= 1024) return `${(mb / 1024).toFixed(1)} GB`;
  return `${Math.round(mb)} MB`;
}

export function UserTable({ users, isLoading, onDelete, onEdit, onMailSetup, onResetPassword, onBulkDelete }: UserTableProps) {
  const [rowSelection, setRowSelection] = useState<Record<string, boolean>>({});
  const [bulkDeleting, setBulkDeleting] = useState(false);
  const [bulkConfirming, setBulkConfirming] = useState(false);

  const columns: ColumnDef<User>[] = [
    {
      id: "select",
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          data-testid="select-all"
          aria-label="Select all users"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          data-testid={`select-row-${row.original.id}`}
          aria-label={`Select user ${row.original.email}`}
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ row }) => (
        <span className="font-medium text-foreground">
          {row.original.email}
        </span>
      ),
    },
    {
      accessorKey: "domainId",
      header: "Domain",
      cell: ({ row }) => {
        const email = row.original.email;
        const domain = email.includes("@") ? email.split("@")[1] : "—";
        return <span className="text-muted-foreground">{domain}</span>;
      },
    },
    {
      accessorKey: "quota",
      header: "Quota",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {formatQuota(row.original.quota)}
        </span>
      ),
    },
    {
      accessorKey: "storageUsed",
      header: "Storage",
      cell: ({ row }) => {
        const used = row.original.storageUsed ?? 0;
        const quota = row.original.quota;
        if (used === 0) return null;
        const percentage = quota > 0
          ? Math.min(100, Math.round((used / quota) * 100))
          : 0;
        return (
          <div className="flex flex-col gap-1 min-w-[140px]" data-testid="storage-cell">
            <div
              className="h-2 w-full bg-muted rounded-full overflow-hidden"
              role="progressbar"
              aria-valuenow={percentage}
              aria-valuemin={0}
              aria-valuemax={100}
            >
              {quota > 0 && (
                <div
                  className="h-full bg-primary rounded-full transition-all"
                  style={{ width: `${percentage}%` }}
                />
              )}
            </div>
            <span className="text-xs text-muted-foreground">
              {formatBytes(used)}
              {quota > 0 ? ` / ${formatQuota(quota)}` : " / Unlimited"}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: "active",
      header: "Status",
      cell: ({ row }) => (
        <Badge variant={row.original.active ? "default" : "secondary"}>
          {row.original.active ? "Active" : "Inactive"}
        </Badge>
      ),
    },
    {
      accessorKey: "createdAt",
      header: "Created",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
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
            aria-label={`Edit user ${row.original.email}`}
          >
            <Pencil className="size-4 text-muted-foreground hover:text-foreground" />
          </Button>
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() => onResetPassword(row.original)}
            aria-label={`Reset password for ${row.original.email}`}
          >
            <KeyRound className="size-4 text-muted-foreground hover:text-foreground" />
          </Button>
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() => onMailSetup(row.original)}
            aria-label={`Mail setup for ${row.original.email}`}
          >
            <Settings className="size-4 text-muted-foreground hover:text-foreground" />
          </Button>
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() => onDelete(row.original)}
            aria-label={`Delete user ${row.original.email}`}
          >
            <Trash2 className="size-4 text-muted-foreground hover:text-destructive" />
          </Button>
        </div>
      ),
    },
  ];

  const table = useReactTable({
    data: users,
    columns,
    getCoreRowModel: getCoreRowModel(),
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    getRowId: (row) => String(row.id),
    state: { rowSelection },
  });

  const selectedRows = table.getFilteredSelectedRowModel().rows;
  const selectedCount = selectedRows.length;

  async function handleBulkDelete() {
    if (selectedCount === 0) return;
    const selectedUsers = selectedRows.map((row) => row.original);
    setBulkDeleting(true);
    try {
      await onBulkDelete(selectedUsers);
      table.resetRowSelection();
      setBulkConfirming(false);
    } finally {
      setBulkDeleting(false);
    }
  }

  if (isLoading) {
    return <UserTableSkeleton />;
  }

  if (users.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-12 text-center">
        <p className="text-sm text-muted-foreground">
          No users found. Create your first user to get started.
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
            {selectedCount} {selectedCount === 1 ? "user" : "users"} selected
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
      <div className="rounded-lg border border-border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
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

function UserTableSkeleton() {
  return (
    <div className="rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-10"><Skeleton className="size-4" /></TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Domain</TableHead>
            <TableHead>Quota</TableHead>
            <TableHead>Storage</TableHead>
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
              <TableCell><Skeleton className="h-4 w-24" /></TableCell>
              <TableCell><Skeleton className="h-4 w-16" /></TableCell>
              <TableCell><Skeleton className="h-4 w-32" /></TableCell>
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
