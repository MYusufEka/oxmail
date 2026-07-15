"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useHealth } from "@/hooks/use-health";
import type { ServiceHealthEntry } from "@/types/api";

const SERVICE_LABELS: Record<string, string> = {
  postfix: "Postfix",
  dovecot: "Dovecot",
  rspamd: "Rspamd",
  redis: "Redis",
  api: "API",
};

function statusColor(status: ServiceHealthEntry["status"]): string {
  switch (status) {
    case "healthy":
      return "bg-emerald-500";
    default:
      return "bg-red-500";
  }
}

function statusBadgeVariant(status: ServiceHealthEntry["status"]) {
  switch (status) {
    case "healthy":
      return "secondary" as const;
    default:
      return "destructive" as const;
  }
}

interface ServiceCardProps {
  name: string;
  health: ServiceHealthEntry;
}

function ServiceCard({ name, health }: ServiceCardProps) {
  const isHealthy = health.status === "healthy";
  return (
    <Card data-testid="service-card" className="gap-3 py-4">
      <CardHeader className="flex-row items-center justify-between gap-2 px-4 py-0">
        <CardTitle className="text-sm font-medium">
          {SERVICE_LABELS[name] ?? name}
        </CardTitle>
        <Badge variant={statusBadgeVariant(health.status)} className="text-xs">
          {isHealthy ? "Healthy" : "Down"}
        </Badge>
      </CardHeader>
      <CardContent className="flex items-center gap-3 px-4">
        <span className="relative flex size-2.5">
          {isHealthy && (
            <span
              className={`absolute inline-flex size-full rounded-full opacity-75 ${statusColor(health.status)}`}
              style={{ animation: "pulse-dot 2s ease-in-out infinite" }}
            />
          )}
          <span
            className={`relative inline-flex size-2.5 rounded-full ${statusColor(health.status)}`}
          />
        </span>
        <span className="text-xs text-muted-foreground">
          {health.latencyMs !== undefined ? `${health.latencyMs}ms` : "—"}
        </span>
      </CardContent>
    </Card>
  );
}

function ServiceCardSkeleton() {
  return (
    <Card className="gap-3 py-4">
      <CardHeader className="flex-row items-center justify-between gap-2 px-4 py-0">
        <Skeleton className="h-4 w-16" />
        <Skeleton className="h-5 w-14 rounded-full" />
      </CardHeader>
      <CardContent className="flex items-center gap-3 px-4">
        <Skeleton className="size-2.5 rounded-full" />
        <Skeleton className="h-3 w-10" />
      </CardContent>
    </Card>
  );
}

export function ServiceHealthGrid() {
  const { data, isLoading } = useHealth();

  if (isLoading || !data) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
        {Array.from({ length: 5 }).map((_, index) => (
          <ServiceCardSkeleton key={index} />
        ))}
      </div>
    );
  }

  const allServices: ServiceHealthEntry[] = [
    ...data.services,
    { name: "api", status: "healthy" as const, latencyMs: 0 },
  ];

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
      {allServices.map((svc) => (
        <ServiceCard key={svc.name} name={svc.name} health={svc} />
      ))}
    </div>
  );
}
