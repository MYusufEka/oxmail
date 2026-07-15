"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useStats } from "@/hooks/use-stats";
import { Send, Inbox, AlertTriangle, ShieldCheck } from "lucide-react";

function formatShortDate(iso: string): string {
  const [year, month, day] = iso.split("-").map(Number);
  return new Date(year, month - 1, day).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

function MiniBar({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? Math.round((value / max) * 100) : 0;
  return (
    <div className="flex h-16 items-end">
      <div
        className="w-full rounded-sm bg-primary/60 transition-all duration-300"
        style={{ height: `${Math.max(pct, 2)}%` }}
      />
    </div>
  );
}

function SkeletonCard() {
  return (
    <Card data-testid="mail-stats-card">
      <CardHeader className="pb-2">
        <Skeleton className="h-4 w-40" />
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="grid grid-cols-4 gap-2">
          {[0, 1, 2, 3].map((i) => (
            <div key={i} className="flex flex-col gap-1.5">
              <Skeleton className="h-3 w-12" />
              <Skeleton className="h-6 w-8" />
            </div>
          ))}
        </div>
        <Skeleton className="h-16 w-full rounded-sm" />
      </CardContent>
    </Card>
  );
}

export function MailStatsCard() {
  const { data: stats, isLoading } = useStats(7);

  if (isLoading) {
    return <SkeletonCard />;
  }

  const days = stats ?? [];

  const totalSent = days.reduce((sum, d) => sum + d.sent, 0);
  const totalReceived = days.reduce((sum, d) => sum + d.received, 0);
  const totalBounced = days.reduce((sum, d) => sum + d.bounced, 0);
  const totalSpamCaught = days.reduce((sum, d) => sum + d.spamCaught, 0);

  const allValues = days.flatMap((d) => [d.sent, d.received, d.bounced]);
  const maxValue = Math.max(...allValues, 1);

  const isAllZero = totalSent === 0 && totalReceived === 0 && totalBounced === 0 && totalSpamCaught === 0;

  return (
    <Card data-testid="mail-stats-card">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">
          Mail Activity (7 days)
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="grid grid-cols-4 gap-2">
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Send className="size-3" />
              Sent
            </span>
            <span className="text-lg font-bold text-foreground">{totalSent}</span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <Inbox className="size-3" />
              Received
            </span>
            <span className="text-lg font-bold text-foreground">{totalReceived}</span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <AlertTriangle className="size-3" />
              Bounced
            </span>
            <span className="text-lg font-bold text-foreground">{totalBounced}</span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="flex items-center gap-1 text-xs text-muted-foreground">
              <ShieldCheck className="size-3" />
              Spam
            </span>
            <span className="text-lg font-bold text-foreground">{totalSpamCaught}</span>
          </div>
        </div>

        {isAllZero ? (
          <div className="flex h-16 items-center justify-center">
            <span className="text-xs text-muted-foreground">
              No mail activity in the last 7 days
            </span>
          </div>
        ) : (
          <div className="grid grid-cols-7 gap-1">
            {[...days].reverse().map((day) => (
              <div key={day.date} className="flex flex-col gap-1">
                <MiniBar value={day.sent + day.received + day.bounced} max={maxValue} />
                <span className="text-center text-[10px] text-muted-foreground">
                  {formatShortDate(day.date)}
                </span>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
