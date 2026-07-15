"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useLogs } from "@/hooks/use-logs";
import { ScrollArea } from "@/components/ui/scroll-area";

const LEVEL_COLORS: Record<string, string> = {
  debug: "text-muted-foreground",
  info: "text-foreground",
  warn: "text-yellow-400",
  error: "text-red-400",
};

const SERVICE_COLORS: Record<string, string> = {
  postfix: "text-blue-400",
  dovecot: "text-emerald-400",
  rspamd: "text-orange-400",
  api: "text-primary",
};

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  });
}

function truncateMessage(message: string, maxLength = 60): string {
  if (message.length <= maxLength) return message;
  return `${message.slice(0, maxLength)}…`;
}

export function RecentActivity() {
  const { data, isLoading } = useLogs({ limit: 10 });

  return (
    <Card className="gap-0 py-0">
      <CardHeader className="px-4 py-3">
        <CardTitle className="text-sm font-medium">Recent Activity</CardTitle>
      </CardHeader>
      <CardContent className="px-0 pb-0">
        <ScrollArea className="h-[280px]">
          {isLoading ? (
            <div className="flex flex-col gap-2 px-4 pb-4">
              {Array.from({ length: 6 }).map((_, index) => (
                <div key={index} className="flex items-center gap-3">
                  <Skeleton className="h-3 w-14" />
                  <Skeleton className="h-3 w-12" />
                  <Skeleton className="h-3 flex-1" />
                </div>
              ))}
            </div>
          ) : !data?.entries?.length ? (
            <div className="flex h-[240px] items-center justify-center">
              <span className="text-xs text-muted-foreground">
                No recent activity
              </span>
            </div>
          ) : (
            <div className="flex flex-col">
              {data.entries?.map((entry) => (
                <div
                  key={entry.id}
                  className="flex items-center gap-3 border-t border-border px-4 py-2 font-mono text-xs"
                >
                  <span className="shrink-0 text-muted-foreground">
                    {formatTimestamp(entry.timestamp)}
                  </span>
                  <span
                    className={`shrink-0 w-14 ${SERVICE_COLORS[entry.service] ?? "text-muted-foreground"}`}
                  >
                    {entry.service}
                  </span>
                  <span
                    className={`truncate ${LEVEL_COLORS[entry.level] ?? "text-foreground"}`}
                  >
                    {truncateMessage(entry.message)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
