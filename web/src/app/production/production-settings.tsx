"use client";

import { Globe, Server, Shield, Gauge } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface ProductionSettingsProps {
  hostname: string;
  publicIp: string;
  tlsEnabled: boolean;
  outboundRateLimit: number;
}

function SettingRow({
  icon: Icon,
  label,
  children,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between py-3 border-b border-border last:border-0">
      <div className="flex items-center gap-3">
        <Icon className="size-4 text-muted-foreground" />
        <span className="text-sm text-foreground">{label}</span>
      </div>
      <div className="text-sm">{children}</div>
    </div>
  );
}

export function ProductionSettings({
  hostname,
  publicIp,
  tlsEnabled,
  outboundRateLimit,
}: ProductionSettingsProps) {
  return (
    <div
      className="rounded-lg border border-border bg-card p-5"
      data-testid="production-settings"
    >
      <h3 className="text-sm font-semibold text-foreground mb-4">
        Server Configuration
      </h3>

      <div className="flex flex-col">
        <SettingRow icon={Globe} label="Hostname">
          <code className="rounded bg-muted px-2 py-0.5 font-mono text-xs text-foreground">
            {hostname || "—"}
          </code>
        </SettingRow>

        <SettingRow icon={Server} label="Public IP">
          <code className="rounded bg-muted px-2 py-0.5 font-mono text-xs text-foreground">
            {publicIp || "Not configured"}
          </code>
        </SettingRow>

        <SettingRow icon={Shield} label="TLS Status">
          {tlsEnabled ? (
            <Badge variant="default" className="bg-emerald-500/15 text-emerald-500 border-emerald-500/20">
              Active
            </Badge>
          ) : (
            <Badge variant="secondary" className="text-muted-foreground">
              Inactive
            </Badge>
          )}
        </SettingRow>

        <SettingRow icon={Gauge} label="Outbound Rate Limit">
          <span className="text-muted-foreground">
            {outboundRateLimit > 0
              ? `${outboundRateLimit} msg/hour`
              : "Unlimited"}
          </span>
        </SettingRow>
      </div>
    </div>
  );
}
