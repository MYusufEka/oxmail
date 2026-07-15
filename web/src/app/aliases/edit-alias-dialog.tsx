"use client";

import { useState, useEffect } from "react";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { RecipientInput } from "@/app/mail/recipient-input";
import { apiClient } from "@/lib/api-client";
import { useQueryClient } from "@tanstack/react-query";
import type { GroupedAlias } from "./alias-table";

interface EditAliasDialogProps {
  alias: GroupedAlias | null;
  domainId: number;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditAliasDialog({
  alias,
  domainId,
  open,
  onOpenChange,
}: EditAliasDialogProps) {
  const queryClient = useQueryClient();
  const [currentDestinations, setCurrentDestinations] = useState<string[]>([]);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (alias && open) {
      setCurrentDestinations([...alias.destinations]);
    }
  }, [alias, open]);

  async function handleSave() {
    if (!alias) return;

    const original = alias.destinations;
    const current = currentDestinations;

    const toRemove = original.filter((d) => !current.includes(d));
    const toAdd = current.filter((d) => !original.includes(d));

    if (toRemove.length === 0 && toAdd.length === 0) {
      onOpenChange(false);
      return;
    }

    setIsSaving(true);
    let hasError = false;
    const log: string[] = [];

    for (const dest of toRemove) {
      const idx = alias.destinations.indexOf(dest);
      const id = alias.ids[idx];
      if (id === undefined) continue;
      try {
        await apiClient.deleteAlias(domainId, id);
        log.push(`Removed ${dest}`);
      } catch {
        hasError = true;
        log.push(`Failed to remove ${dest}`);
      }
    }

    for (const dest of toAdd) {
      try {
        await apiClient.createAlias(domainId, {
          sourceAddress: alias.sourceAddress,
          destinationAddress: dest,
        });
        log.push(`Added ${dest}`);
      } catch {
        hasError = true;
        log.push(`Failed to add ${dest}`);
      }
    }

    queryClient.invalidateQueries({ queryKey: ["aliases", domainId] });

    if (hasError) {
      toast.error("Some changes failed", { description: log.join(", ") });
    } else {
      toast.success("Alias updated", {
        description: `${log.length} change(s) applied.`,
      });
    }

    setIsSaving(false);
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-testid="edit-alias-dialog">
        <DialogHeader>
          <DialogTitle>Edit Alias</DialogTitle>
          <DialogDescription>
            Manage destinations for {alias?.sourceAddress ?? ""}.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div>
            <label className="text-xs font-medium text-muted-foreground">
              Source
            </label>
            <p
              className="text-sm font-medium text-foreground"
              data-testid="edit-alias-source"
            >
              {alias?.sourceAddress ?? ""}
            </p>
          </div>

          <RecipientInput
            label="Destinations"
            recipients={currentDestinations}
            onChange={setCurrentDestinations}
            placeholder="user@example.com"
          />
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button
            type="button"
            onClick={handleSave}
            disabled={isSaving}
            data-testid="save-edit-alias"
          >
            {isSaving && <Loader2 className="size-4 animate-spin" />}
            Save Changes
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
