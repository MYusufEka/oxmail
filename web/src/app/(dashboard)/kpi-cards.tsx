"use client";

import { Globe, Users, Mail, Clock } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useDomains } from "@/hooks/use-domains";
import { useLogs } from "@/hooks/use-logs";

interface KpiCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  trend?: string;
}

function KpiCard({ icon, label, value, trend }: KpiCardProps) {
  return (
    <Card className="gap-4 py-4">
      <CardContent className="flex items-center gap-4 px-4">
        <div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
          {icon}
        </div>
        <div className="flex flex-col gap-0.5">
          <span className="text-xs text-muted-foreground">{label}</span>
          <span className="text-2xl font-bold tracking-tight text-foreground">
            {value}
          </span>
          {trend && (
            <span className="text-xs text-muted-foreground">{trend}</span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function KpiCardSkeleton() {
  return (
    <Card className="gap-4 py-4">
      <CardContent className="flex items-center gap-4 px-4">
        <Skeleton className="size-10 rounded-lg" />
        <div className="flex flex-col gap-1.5">
          <Skeleton className="h-3 w-20" />
          <Skeleton className="h-7 w-14" />
        </div>
      </CardContent>
    </Card>
  );
}

export function KpiCards() {
  const domains = useDomains({ limit: 1 });
  const logs = useLogs({ limit: 1 });

  const isLoading = domains.isLoading || logs.isLoading;

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KpiCardSkeleton />
        <KpiCardSkeleton />
        <KpiCardSkeleton />
        <KpiCardSkeleton />
      </div>
    );
  }

  const totalDomains = domains.data?.pagination.total ?? 0;
  const totalEmails = logs.data?.pagination.total ?? 0;

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <KpiCard
        icon={<Globe className="size-5" />}
        label="Total Domains"
        value={totalDomains}
        trend="Active"
      />
      <KpiCard
        icon={<Users className="size-5" />}
        label="Total Users"
        value="—"
        trend="Across all domains"
      />
      <KpiCard
        icon={<Mail className="size-5" />}
        label="Emails Today"
        value={totalEmails}
        trend="Log entries"
      />
      <KpiCard
        icon={<Clock className="size-5" />}
        label="Uptime"
        value="99.9%"
        trend="Last 30 days"
      />
    </div>
  );
}
