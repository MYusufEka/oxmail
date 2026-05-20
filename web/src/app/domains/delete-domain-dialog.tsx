"use client";

import { Loader2, AlertTriangle } from "lucide-react";
import type { Domain } from "@/types/api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface DeleteDomainDialogProps {
  domain: Domain | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isPending: boolean;
}

export function DeleteDomainDialog({
  domain,
  open,
  onOpenChange,
  onConfirm,
  isPending,
}: DeleteDomainDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-testid="delete-domain-dialog">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="size-5 text-destructive" />
            Delete Domain
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to delete{" "}
            <span className="font-medium text-foreground">{domain?.name}</span>?
            This will remove all associated users, aliases, and DKIM keys. This
            action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isPending}
            data-testid="confirm-delete-domain"
          >
            {isPending && <Loader2 className="animate-spin" />}
            Delete Domain
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
