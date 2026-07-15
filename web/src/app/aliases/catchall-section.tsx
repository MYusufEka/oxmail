"use client";

import { useState } from "react";
import { Mail } from "lucide-react";
import { toast } from "sonner";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import type { GroupedAlias } from "./alias-table";

interface CatchallSectionProps {
  domainId: number;
  domainName: string;
  catchallAlias: GroupedAlias | null;
}

export function CatchallSection({
  domainId,
  domainName,
  catchallAlias,
}: CatchallSectionProps) {
  const queryClient = useQueryClient();
  const isEnabled = catchallAlias !== null;

  const [destination, setDestination] = useState(
    catchallAlias?.destinations[0] ?? "",
  );
  const [pending, setPending] = useState(false);

  async function handleToggleOn() {
    const trimmed = destination.trim();
    if (!trimmed) {
      toast.error("Enter a destination email address first.");
      return;
    }

    setPending(true);
    try {
      await apiClient.createAlias(domainId, {
        sourceAddress: `@${domainName}`,
        destinationAddress: trimmed,
      });
      queryClient.invalidateQueries({ queryKey: ["aliases", domainId] });
      toast.success("Catch-all enabled", {
        description: `@${domainName} → ${trimmed}`,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Something went wrong.";
      toast.error("Failed to enable catch-all", { description: message });
    } finally {
      setPending(false);
    }
  }

  async function handleToggleOff() {
    if (!catchallAlias) return;

    setPending(true);
    try {
      await Promise.all(
        catchallAlias.ids.map((id) => apiClient.deleteAlias(domainId, id)),
      );
      queryClient.invalidateQueries({ queryKey: ["aliases", domainId] });
      setDestination("");
      toast.success("Catch-all disabled", {
        description: `@${domainName} no longer catches unmatched mail.`,
      });
    } catch (err) {
      const message = err instanceof Error ? err.message : "Something went wrong.";
      toast.error("Failed to disable catch-all", { description: message });
    } finally {
      setPending(false);
    }
  }

  function handleCheckedChange(checked: boolean) {
    if (checked) {
      handleToggleOn();
    } else {
      handleToggleOff();
    }
  }

  return (
    <div
      className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4"
      data-testid="catchall-section"
    >
      <div className="flex items-center gap-2">
        <Mail className="size-4 text-muted-foreground" />
        <span className="text-sm font-medium text-foreground">Catch-all</span>
        <span className="text-xs text-muted-foreground">
          — route unmatched mail for{" "}
          <span className="font-mono text-foreground">@{domainName}</span>
        </span>
      </div>

      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2">
          <Switch
            id="catchall-toggle"
            checked={isEnabled}
            onCheckedChange={handleCheckedChange}
            disabled={pending}
            data-testid="catchall-toggle"
            aria-label={`Toggle catch-all for ${domainName}`}
          />
          <Label
            htmlFor="catchall-toggle"
            className="cursor-pointer select-none text-sm"
          >
            {isEnabled ? "Enabled" : "Disabled"}
          </Label>
        </div>

        {isEnabled ? (
          <p className="text-sm text-muted-foreground" data-testid="catchall-destination">
            Delivering to{" "}
            <span className="font-mono font-medium text-foreground">
              {catchallAlias.destinations[0]}
            </span>
          </p>
        ) : (
          <div className="flex items-center gap-2">
            <Input
              type="email"
              placeholder="destination@example.com"
              value={destination}
              onChange={(e) => setDestination(e.target.value)}
              className="h-8 w-56 text-sm"
              data-testid="catchall-destination-input"
              aria-label="Catch-all destination email"
              disabled={pending}
              onKeyDown={(e) => {
                if (e.key === "Enter") handleToggleOn();
              }}
            />
            <Button
              size="sm"
              variant="outline"
              onClick={handleToggleOn}
              disabled={pending || !destination.trim()}
              data-testid="catchall-enable-btn"
            >
              Enable
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
