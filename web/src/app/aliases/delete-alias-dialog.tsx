"use client";

import { Loader2, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface DeleteAliasDialogProps {
  sourceAddress: string;
  destinations: string[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: () => void;
  isPending: boolean;
}

export function DeleteAliasDialog({
  sourceAddress,
  destinations,
  open,
  onOpenChange,
  onConfirm,
  isPending,
}: DeleteAliasDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-testid="delete-alias-dialog">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertTriangle className="size-5 text-destructive" />
            Delete Alias
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to delete the alias{" "}
            <span className="font-medium text-foreground">
              {sourceAddress}
            </span>
            ? Mail forwarding to{" "}
            <span className="font-medium text-foreground">
              {destinations.join(", ")}
            </span>{" "}
            will stop immediately. This action cannot be undone.
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
            data-testid="confirm-delete-alias"
          >
            {isPending && <Loader2 className="animate-spin" />}
            Delete Alias
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
