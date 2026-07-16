"use client";

import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from "@tanstack/react-table";
import { Trash2 } from "lucide-react";
import type { Domain } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { DomainHealthPopover } from "@/app/domains/domain-health-popover";

interface DomainTableProps {
  domains: Domain[];
  onDelete: (domain: Domain) => void;
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}


function createColumns(onDelete: (domain: Domain) => void): ColumnDef<Domain>[] {
  return [
    {
      accessorKey: "name",
      header: "Domain",
      cell: ({ row }) => (
        <span className="font-medium text-foreground" data-testid="domain-name">
          {row.getValue("name")}
        </span>
      ),
    },
    {
      id: "users",
      header: "Users",
      cell: () => (
        <span className="text-muted-foreground">—</span>
      ),
    },
    {
      id: "aliases",
      header: "Aliases",
      cell: () => (
        <span className="text-muted-foreground">—</span>
      ),
    },
    {
      id: "health",
      header: "Health",
      cell: ({ row }) => <DomainHealthPopover name={row.original.name} />,
    },
    {
      id: "dkim",
      header: "DKIM",
      cell: ({ row }) => (
        <Badge
          variant={row.original.active ? "default" : "secondary"}
          className="text-xs"
          data-testid="dkim-status"
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
          {formatDate(row.getValue("createdAt"))}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex justify-end">
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={() => onDelete(row.original)}
            aria-label={`Delete ${row.original.name}`}
            data-testid={`delete-domain-${row.original.id}`}
          >
            <Trash2 className="size-3.5 text-muted-foreground hover:text-destructive" />
          </Button>
        </div>
      ),
    },
  ];
}

export function DomainTable({ domains, onDelete }: DomainTableProps) {
  const columns = createColumns(onDelete);

  const table = useReactTable({
    data: domains,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div className="rounded-lg border border-border" data-testid="domain-table">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id} className="border-border hover:bg-transparent">
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id} className="text-muted-foreground">
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
            <TableRow key={row.id} className="border-border">
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
  );
}
