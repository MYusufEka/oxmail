"use client";

import { useRef, useState } from "react";
import { Download, Upload } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useImportUsers } from "@/hooks/use-users";
import type { UserImportError, UserImportResult } from "@/types/api";

const TEMPLATE_CSV = "email,password,quota_mb\nuser@example.com,SecurePass123,1024\n";
const PREVIEW_ROW_LIMIT = 5;

interface ImportCsvDialogProps {
  domainId: number;
}

function parseCsvPreview(text: string): string[][] {
  const lines = text.split(/\r?\n/).filter((line) => line.trim() !== "");
  return lines.slice(0, PREVIEW_ROW_LIMIT + 1).map((line) =>
    line.split(",").map((cell) => cell.trim()),
  );
}

function downloadTemplate() {
  const blob = new Blob([TEMPLATE_CSV], { type: "text/csv" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = "users-import-template.csv";
  anchor.click();
  URL.revokeObjectURL(url);
}

export function ImportCsvDialog({ domainId }: ImportCsvDialogProps) {
  const [open, setOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewRows, setPreviewRows] = useState<string[][]>([]);
  const [importResult, setImportResult] = useState<UserImportResult | null>(null);
  const [rowErrors, setRowErrors] = useState<UserImportError[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const importUsers = useImportUsers(domainId);

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0] ?? null;
    setSelectedFile(file);
    setRowErrors([]);
    setImportResult(null);
    setPreviewRows([]);

    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result;
      if (typeof text === "string") {
        setPreviewRows(parseCsvPreview(text));
      }
    };
    reader.readAsText(file);
  }

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      setSelectedFile(null);
      setPreviewRows([]);
      setImportResult(null);
      setRowErrors([]);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
    setOpen(nextOpen);
  }

  function handleSubmit() {
    if (!selectedFile) return;

    importUsers.mutate(selectedFile, {
      onSuccess: (result) => {
        setImportResult(result);
        if (result.errors && result.errors.length > 0) {
          setRowErrors(result.errors);
        }
        toast.success(
          `Import complete: ${result.created} created, ${result.skipped} skipped, ${result.errors?.length ?? 0} errors`,
        );
      },
      onError: (err) => {
        toast.error(err instanceof Error ? err.message : "Import failed");
      },
    });
  }

  const headerRow = previewRows[0] ?? [];
  const dataRows = previewRows.slice(1);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button variant="outline" size="sm" data-testid="import-csv-btn">
          <Upload className="mr-2 h-4 w-4" />
          Import CSV
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl" data-testid="import-csv-dialog">
        <DialogHeader>
          <DialogTitle>Import Users from CSV</DialogTitle>
          <DialogDescription>
            Upload a CSV with columns: <code>email,password,quota_mb</code>. Header row required.
            Max 100 rows.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* File picker + template download */}
          <div className="flex items-center gap-3">
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv,text/csv"
              onChange={handleFileChange}
              className="flex-1 text-sm file:mr-3 file:cursor-pointer file:rounded file:border-0 file:bg-muted file:px-3 file:py-1.5 file:text-sm file:font-medium hover:file:bg-muted/80"
              data-testid="import-csv-file-input"
            />
            <Button
              variant="ghost"
              size="sm"
              onClick={downloadTemplate}
              data-testid="import-csv-template-btn"
            >
              <Download className="mr-2 h-4 w-4" />
              Template
            </Button>
          </div>

          {/* CSV preview */}
          {previewRows.length > 0 && !importResult && (
            <div className="rounded-md border">
              <p className="border-b px-3 py-2 text-xs font-medium text-muted-foreground">
                Preview (first {Math.min(dataRows.length, PREVIEW_ROW_LIMIT)} rows)
              </p>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="border-b bg-muted/40">
                      {headerRow.map((col, i) => (
                        <th key={i} className="px-3 py-2 text-left font-medium">
                          {col}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {dataRows.slice(0, PREVIEW_ROW_LIMIT).map((row, ri) => (
                      <tr key={ri} className="border-b last:border-0">
                        {row.map((cell, ci) => (
                          <td key={ci} className="px-3 py-2 text-muted-foreground">
                            {cell}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Import result summary */}
          {importResult && (
            <div className="rounded-md border bg-muted/30 p-3 text-sm" data-testid="import-csv-result">
              <p className="font-medium">Import complete</p>
              <ul className="mt-1 space-y-0.5 text-muted-foreground">
                <li>Created: <span className="font-medium text-foreground">{importResult.created}</span></li>
                <li>Skipped: <span className="font-medium text-foreground">{importResult.skipped}</span></li>
                <li>
                  Errors:{" "}
                  <span className="font-medium text-foreground">
                    {importResult.errors?.length ?? 0}
                  </span>
                </li>
              </ul>
            </div>
          )}

          {/* Row-level errors */}
          {rowErrors.length > 0 && (
            <div className="max-h-40 overflow-y-auto rounded-md border border-destructive/40 bg-destructive/5 p-3">
              <p className="mb-2 text-xs font-medium text-destructive">Row errors:</p>
              <ul className="space-y-1">
                {rowErrors.map((e) => (
                  <li key={e.row} className="text-xs text-muted-foreground">
                    Row {e.row}
                    {e.email ? ` (${e.email})` : ""}: {e.error}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => handleOpenChange(false)}>
            {importResult ? "Close" : "Cancel"}
          </Button>
          {!importResult && (
            <Button
              onClick={handleSubmit}
              disabled={!selectedFile || importUsers.isPending}
              data-testid="import-csv-submit"
            >
              {importUsers.isPending ? "Importing…" : "Import"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
