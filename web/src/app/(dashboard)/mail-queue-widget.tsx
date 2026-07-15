"use client";

import { RefreshCw } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { useMailQueue } from "@/hooks/use-mail";

function queueIndicatorColor(total: number, deferred: number): string {
  if (deferred > 10) return "bg-red-500";
  if (deferred > 0) return "bg-amber-500";
  return "bg-emerald-500";
}

function queueBadgeVariant(deferred: number): "destructive" | "secondary" {
  if (deferred > 10) return "destructive";
  if (deferred > 0) return "secondary";
  return "secondary";
}

interface StatItemProps {
  label: string;
  value: number;
}

function StatItem({ label, value }: StatItemProps) {
  return (
    <div className="flex flex-col items-center gap-1">
      <span className="text-2xl font-bold tabular-nums text-foreground">
        {value}
      </span>
      <span className="text-xs text-muted-foreground">{label}</span>
    </div>
  );
}

function WidgetSkeleton() {
  return (
    <Card className="gap-3 py-4">
      <CardHeader className="flex-row items-center justify-between gap-2 px-4 py-0">
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-5 w-16 rounded-full" />
      </CardHeader>
      <CardContent className="flex items-center justify-between px-4">
        <div className="flex gap-6">
          <Skeleton className="h-12 w-12" />
          <Skeleton className="h-12 w-12" />
          <Skeleton className="h-12 w-12" />
        </div>
        <Skeleton className="size-8 rounded-md" />
      </CardContent>
    </Card>
  );
}

export function MailQueueWidget() {
  const { data, isLoading, refetch, isFetching } = useMailQueue();

  if (isLoading || !data) {
    return <WidgetSkeleton />;
  }

  const indicatorColor = queueIndicatorColor(data.total, data.deferred);

  return (
    <Card data-testid="mail-queue-widget" className="gap-3 py-4">
      <CardHeader className="flex-row items-center justify-between gap-2 px-4 py-0">
        <CardTitle className="text-sm font-medium">Mail Queue</CardTitle>
        <Badge variant={queueBadgeVariant(data.deferred)} className="text-xs">
          {data.deferred > 10
            ? "Backed up"
            : data.deferred > 0
              ? "Deferred"
              : "Healthy"}
        </Badge>
      </CardHeader>
      <CardContent className="flex items-center justify-between px-4">
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-3">
            <span className="relative flex size-2.5">
              <span
                className={`absolute inline-flex size-full rounded-full opacity-75 ${indicatorColor}`}
                style={{ animation: "pulse-dot 2s ease-in-out infinite" }}
              />
              <span
                className={`relative inline-flex size-2.5 rounded-full ${indicatorColor}`}
              />
            </span>
            <StatItem label="Total" value={data.total} />
          </div>
          <StatItem label="Active" value={data.active} />
          <StatItem label="Deferred" value={data.deferred} />
        </div>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => refetch()}
          disabled={isFetching}
          data-testid="mail-queue-refresh"
          aria-label="Refresh mail queue"
        >
          <RefreshCw className={`size-4 ${isFetching ? "animate-spin" : ""}`} />
        </Button>
      </CardContent>
    </Card>
  );
}
