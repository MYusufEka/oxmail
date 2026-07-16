"use client";

import { Popover, PopoverTrigger, PopoverContent } from "@/components/ui/popover";
import { useDomainHealth } from "@/hooks/use-domains";
import type { DomainCheckResult } from "@/types/api";

interface DomainHealthPopoverProps {
  name: string;
}

const CHECK_NAMES = ["mx", "spf", "dmarc", "dkim", "dovecot"] as const;

function statusIcon(status: DomainCheckResult["status"]): { icon: string; className: string } {
  if (status === "pass") return { icon: "✓", className: "text-green-500" };
  if (status === "warn") return { icon: "⚠", className: "text-amber-500" };
  return { icon: "✗", className: "text-red-500" };
}

function aggregateLabel(status: "healthy" | "degraded" | "unhealthy"): string {
  if (status === "healthy") return "Healthy";
  if (status === "degraded") return "Degraded";
  return "Unhealthy";
}

export function DomainHealthPopover({ name }: DomainHealthPopoverProps) {
  const { data, isLoading, isError } = useDomainHealth(name);

  const dotClass = isError
    ? "inline-block size-2.5 rounded-full bg-red-500/60"
    : isLoading || !data
      ? "inline-block size-2.5 rounded-full bg-muted-foreground/40 animate-pulse"
      : data.status === "healthy"
        ? "inline-block size-2.5 rounded-full bg-green-500"
        : data.status === "degraded"
          ? "inline-block size-2.5 rounded-full bg-amber-500"
          : "inline-block size-2.5 rounded-full bg-red-500";

  const ariaLabel = isError
    ? "Health check failed"
    : isLoading || !data
      ? "Loading health status"
      : `Health: ${data.status}`;

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="inline-flex items-center justify-center cursor-pointer"
          aria-label={ariaLabel}
          data-testid={`domain-health-trigger-${name}`}
        >
          <span
            className={dotClass}
            data-testid={`domain-health-badge-${name}`}
          />
        </button>
      </PopoverTrigger>
      <PopoverContent data-testid="domain-health-popover" className="w-80 p-0">
        {!data ? (
          <div className="p-4 text-sm text-muted-foreground">Loading…</div>
        ) : (
          <div>
            <div className="flex items-center gap-2 border-b border-border/50 px-4 py-3">
              <span
                className={
                  data.status === "healthy"
                    ? "inline-block size-2 rounded-full bg-green-500"
                    : data.status === "degraded"
                      ? "inline-block size-2 rounded-full bg-amber-500"
                      : "inline-block size-2 rounded-full bg-red-500"
                }
              />
              <span className="text-sm font-medium text-foreground">
                {name} — {aggregateLabel(data.status)}
              </span>
            </div>
            <ul className="px-4 py-3 space-y-2">
              {CHECK_NAMES.map((checkName) => {
                const check = data.checks.find((c) => c.name === checkName);
                const checkStatus = check?.status ?? "fail";
                const detail = check?.detail ?? "No data";
                const { icon, className } = statusIcon(checkStatus);
                return (
                  <li key={checkName} className="flex items-start gap-2 text-sm">
                    <span className={`${className} shrink-0 w-4 text-center font-medium`}>
                      {icon}
                    </span>
                    <span className="font-medium text-foreground capitalize w-14 shrink-0">
                      {checkName}
                    </span>
                    <span className="text-muted-foreground truncate">{detail}</span>
                  </li>
                );
              })}
            </ul>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
