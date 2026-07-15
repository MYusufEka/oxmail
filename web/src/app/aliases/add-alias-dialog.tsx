"use client";

import { useState } from "react";
import { Plus, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { RecipientInput } from "@/app/mail/recipient-input";
import { apiClient } from "@/lib/api-client";
import { useQueryClient } from "@tanstack/react-query";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

interface AddAliasDialogProps {
  domainId: number;
}

export function AddAliasDialog({ domainId }: AddAliasDialogProps) {
  const [open, setOpen] = useState(false);
  const [sourceAddress, setSourceAddress] = useState("");
  const [destinations, setDestinations] = useState<string[]>([]);
  const [sourceError, setSourceError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const queryClient = useQueryClient();

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();

    if (!sourceAddress.trim()) {
      setSourceError("Source address is required");
      return;
    }

    if (!EMAIL_REGEX.test(sourceAddress.trim())) {
      setSourceError("Enter a valid email address");
      return;
    }

    if (destinations.length === 0) {
      toast.error("Add at least one destination address");
      return;
    }

    setSourceError(null);
    setIsSaving(true);

    try {
      await Promise.all(
        destinations.map((dest) =>
          apiClient.createAlias(domainId, {
            sourceAddress: sourceAddress.trim(),
            destinationAddress: dest,
          }),
        ),
      );

      queryClient.invalidateQueries({ queryKey: ["aliases", domainId] });

      toast.success("Aliases created", {
        description: `${sourceAddress} → ${destinations.length} destination(s)`,
      });
      setSourceAddress("");
      setDestinations([]);
      setOpen(false);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "Failed to create aliases";
      toast.error("Failed to create aliases", { description: message });
    } finally {
      setIsSaving(false);
    }
  }

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen && !isSaving) {
      setSourceAddress("");
      setDestinations([]);
      setSourceError(null);
    }
    setOpen(nextOpen);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button data-testid="add-alias-trigger">
          <Plus className="size-4" />
          Add Alias
        </Button>
      </DialogTrigger>
      <DialogContent data-testid="add-alias-dialog">
        <DialogHeader>
          <DialogTitle>Add Alias</DialogTitle>
          <DialogDescription>
            Forward mail from a source address to one or more destinations.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={handleSubmit}
          className="grid gap-6"
          data-testid="add-alias-form"
        >
          <div>
            <label className="text-xs font-medium text-muted-foreground mb-1.5 block">
              Source address
            </label>
            <Input
              type="email"
              placeholder="alias@example.com"
              value={sourceAddress}
              onChange={(event) => {
                setSourceAddress(event.target.value);
                setSourceError(null);
              }}
              autoFocus
              disabled={isSaving}
              data-testid="alias-source-input"
              className={sourceError ? "border-destructive" : ""}
            />
            {sourceError && (
              <p
                className="text-xs text-destructive mt-1"
                data-testid="alias-source-error"
              >
                {sourceError}
              </p>
            )}
          </div>

          <RecipientInput
            label="Destination addresses"
            recipients={destinations}
            onChange={setDestinations}
            placeholder="user@example.com"
          />

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isSaving}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isSaving}
              data-testid="add-alias-submit"
            >
              {isSaving && <Loader2 className="size-4 animate-spin" />}
              Add Alias
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
