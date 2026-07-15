"use client";

import { Copy, Check, Inbox, Send, Key, Shield, Download } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { User } from "@/types/api";

interface MailSetupDialogProps {
  user: User | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function CopyableField({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard not available
    }
  }

  return (
    <div className="flex items-center justify-between py-2 border-b border-border last:border-0">
      <span className="text-xs text-muted-foreground">{label}</span>
      <button
        type="button"
        onClick={handleCopy}
        className="flex items-center gap-1.5 rounded bg-muted px-2 py-0.5 font-mono text-xs text-foreground hover:bg-muted/80 transition-colors"
      >
        {value}
        {copied ? (
          <Check className="size-3 text-emerald-500" />
        ) : (
          <Copy className="size-3 text-muted-foreground" />
        )}
      </button>
    </div>
  );
}

function SectionHeader({ icon: Icon, title }: { icon: React.ComponentType<{ className?: string }>; title: string }) {
  return (
    <p className="text-xs font-medium text-muted-foreground mb-1 flex items-center gap-1.5">
      <Icon className="size-3" />
      {title}
    </p>
  );
}

function isDevMode(): boolean {
  if (typeof window === "undefined") return true;
  return (
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1" ||
    process.env.NEXT_PUBLIC_API_URL?.includes("localhost") === true
  );
}

function extractDomain(email: string): string {
  const parts = email.split("@");
  return parts.length === 2 ? parts[1] : "";
}

export function MailSetupDialog({ user, open, onOpenChange }: MailSetupDialogProps) {
  const dev = isDevMode();
  const smtpPort = dev ? "1025" : "25";
  const imapPort = dev ? "1143" : "143";
  const pop3Port = dev ? "1100" : "110";

  if (!user) return null;

  const domain = extractDomain(user.email);
  const hostname = `mail.${domain}`;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Mail Setup</DialogTitle>
          <DialogDescription>
            Connection settings for {user.email}.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-3">
          <div className="rounded-lg border border-border bg-card p-4">
            <SectionHeader icon={Send} title="SMTP (Outgoing)" />
            <div className="space-y-0">
              <CopyableField label="Server" value={hostname} />
              <CopyableField label={`Port ${smtpPort}`} value="STARTTLS" />
              <CopyableField label="Auth" value="Normal password" />
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-4">
            <SectionHeader icon={Inbox} title="IMAP (Incoming)" />
            <div className="space-y-0">
              <CopyableField label="Server" value={hostname} />
              <CopyableField label={`Port ${imapPort}`} value="STARTTLS" />
              <CopyableField label="Auth" value="Normal password" />
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-4">
            <SectionHeader icon={Download} title="POP3 (Incoming)" />
            <div className="space-y-0">
              <CopyableField label="Server" value={hostname} />
              <CopyableField label={`Port ${pop3Port}`} value="STARTTLS" />
              <CopyableField label="Port 995" value="POP3S (Implicit TLS)" />
              <CopyableField label="Auth" value="PLAIN / LOGIN" />
            </div>
          </div>

          <div className="rounded-lg border border-border bg-card p-4">
            <SectionHeader icon={Key} title="Your Credentials" />
            <div className="space-y-0">
              <CopyableField label="Username" value={user.email} />
              <div className="flex items-center justify-between py-2">
                <span className="text-xs text-muted-foreground">Password</span>
                <span className="text-xs text-muted-foreground">
                  Your mailbox password
                </span>
              </div>
            </div>
          </div>

          <div className="rounded-lg border border-border/50 bg-muted/30 p-3">
            <div className="flex items-start gap-2">
              <Shield className="size-3.5 text-muted-foreground mt-0.5 shrink-0" />
              <p className="text-xs text-muted-foreground">
                STARTTLS recommended. Use full email address as username.
                {dev && " Dev mode: ports differ from production."}
              </p>
            </div>
          </div>
        </div>

        <div className="flex justify-end">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
