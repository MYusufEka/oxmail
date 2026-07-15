"use client";

import { useEffect, useState } from "react";
import { Globe, Server, Shield, Gauge, Mail, Inbox, Send, Download } from "lucide-react";
import { Badge } from "@/components/ui/badge";

interface ProductionSettingsProps {
  hostname: string;
  publicIp: string;
  tlsEnabled: boolean;
  outboundRateLimit: number;
  isDevMode?: boolean;
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
  isDevMode,
}: ProductionSettingsProps) {
  const smtpPort = isDevMode ? "1025" : "25";
  const imapPort = isDevMode ? "1143" : "143";
  const pop3Port = isDevMode ? "1100" : "110";
  const imapsPort = "993";
  const submissionPort = "587";
  const [webUrl, setWebUrl] = useState("");

  useEffect(() => {
    setWebUrl(window.location.origin);
  }, []);

  const emailDomain = hostname.replace("mail.", "");

  return (
    <div className="flex flex-col gap-4">
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

      <div
        className="rounded-lg border border-border bg-card p-5"
        data-testid="mail-server-settings"
      >
        <h3 className="text-sm font-semibold text-foreground mb-4">
          Mail Server Settings
        </h3>

        <div className="flex flex-col">
          <div className="py-2 border-b border-border">
            <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1.5">
              <Send className="size-3" />
              SMTP (Outgoing)
            </p>
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Server</span>
                <code className="font-mono text-foreground">{hostname || "—"}</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port {smtpPort}</span>
                <code className="font-mono text-foreground">Plain / STARTTLS</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port {submissionPort}</span>
                <code className="font-mono text-foreground">Submission (STARTTLS)</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port 465</span>
                <code className="font-mono text-foreground">SMTPS (Implicit TLS)</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Auth</span>
                <code className="font-mono text-foreground">PLAIN / LOGIN</code>
              </div>
            </div>
          </div>

          <div className="py-2 border-b border-border">
            <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1.5">
              <Inbox className="size-3" />
              IMAP (Incoming)
            </p>
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Server</span>
                <code className="font-mono text-foreground">{hostname || "—"}</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port {imapPort}</span>
                <code className="font-mono text-foreground">Plain (STARTTLS)</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port {imapsPort}</span>
                <code className="font-mono text-foreground">IMAPS (Implicit TLS)</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Auth</span>
                <code className="font-mono text-foreground">PLAIN / LOGIN</code>
              </div>
            </div>
          </div>

          <div className="py-2 border-b border-border">
            <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1.5">
              <Download className="size-3" />
              POP3 (Incoming)
            </p>
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Server</span>
                <code className="font-mono text-foreground">{hostname || "—"}</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port {pop3Port}</span>
                <code className="font-mono text-foreground">Plain / STARTTLS</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Port 995</span>
                <code className="font-mono text-foreground">POP3S (Implicit TLS)</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Auth</span>
                <code className="font-mono text-foreground">PLAIN / LOGIN</code>
              </div>
            </div>
          </div>

          <div className="py-2">
            <p className="text-xs font-medium text-muted-foreground mb-2 flex items-center gap-1.5">
              <Mail className="size-3" />
              Email Format
            </p>
            <div className="space-y-1">
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Default User</span>
                <code className="font-mono text-foreground">user@{emailDomain}</code>
              </div>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Webmail</span>
                <code className="font-mono text-foreground">{webUrl || "—"}</code>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
