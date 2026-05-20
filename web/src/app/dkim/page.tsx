"use client";

import dynamic from "next/dynamic";
import { Key, Loader2 } from "lucide-react";
import { useDomains } from "@/hooks/use-domains";
import { Skeleton } from "@/components/ui/skeleton";

const DkimDomainCard = dynamic(
  () => import("./dkim-domain-card").then((mod) => ({ default: mod.DkimDomainCard })),
  { ssr: false }
);

export default function DkimPage() {
  const { data: domainsResponse, isLoading } = useDomains();

  const domains = domainsResponse?.data ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-3">
        <Key className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">DKIM</h2>
      </div>

      {isLoading && (
        <div className="rounded-lg border border-border" data-testid="dkim-skeleton">
          <div className="border-b border-border p-4">
            <Skeleton className="h-6 w-40" />
          </div>
          {Array.from({ length: 3 }).map((_, index) => (
            <div key={index} className="border-b border-border p-4 last:border-0">
              <Skeleton className="h-10 w-full" />
            </div>
          ))}
        </div>
      )}

      {!isLoading && domains.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-lg border border-border bg-card p-12">
          <div className="flex size-12 items-center justify-center rounded-full bg-muted">
            <Key className="size-6 text-muted-foreground" />
          </div>
          <h3 className="mt-4 text-base font-medium text-foreground">
            No domains configured
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Add a domain first to manage DKIM keys.
          </p>
        </div>
      )}

      {!isLoading && domains.length > 0 && (
        <div className="flex flex-col gap-3">
          {domains.map((domain) => (
            <DkimDomainCard key={domain.id} domain={domain.name} />
          ))}
        </div>
      )}
    </div>
  );
}
