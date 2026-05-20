"use client";

import { Shield, AlertTriangle } from "lucide-react";
import { DnsWizard } from "./dns-wizard";
import { ProductionSettings } from "./production-settings";

const OXMAIL_MODE = process.env.NEXT_PUBLIC_OXMAIL_MODE ?? "dev";
const OXMAIL_DOMAIN = process.env.NEXT_PUBLIC_OXMAIL_DOMAIN ?? "local.test";
const OXMAIL_PUBLIC_IP = process.env.NEXT_PUBLIC_OXMAIL_PUBLIC_IP ?? "";
const OXMAIL_TLS_ENABLED = process.env.NEXT_PUBLIC_OXMAIL_TLS_ENABLED === "true";
const OXMAIL_RATE_LIMIT = Number(process.env.NEXT_PUBLIC_OXMAIL_RATE_LIMIT ?? "0");

export default function ProductionPage() {
  const isDevMode = OXMAIL_MODE === "dev";

  return (
    <div className="flex flex-col gap-6" data-testid="production-page">
      <div className="flex items-center gap-3">
        <Shield className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Production</h2>
      </div>

      {isDevMode && (
        <div
          className="flex items-center gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 px-4 py-3"
          role="alert"
          data-testid="dev-mode-warning"
        >
          <AlertTriangle className="size-4 shrink-0 text-amber-500" />
          <p className="text-sm text-amber-500">
            You&apos;re viewing production settings in dev mode
          </p>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
        <div className="flex flex-col gap-6">
          <div>
            <h3 className="text-sm font-semibold text-foreground mb-1">
              DNS Setup Wizard
            </h3>
            <p className="text-xs text-muted-foreground mb-4">
              Configure these DNS records to enable email delivery for your domain.
            </p>
            <DnsWizard domain={OXMAIL_DOMAIN} />
          </div>
        </div>

        <div className="flex flex-col gap-4">
          <ProductionSettings
            hostname={`mail.${OXMAIL_DOMAIN}`}
            publicIp={OXMAIL_PUBLIC_IP}
            tlsEnabled={OXMAIL_TLS_ENABLED}
            outboundRateLimit={OXMAIL_RATE_LIMIT}
          />
        </div>
      </div>
    </div>
  );
}
