"use client";

import { useState } from "react";
import { Check, Copy, X, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface DnsRecordStepProps {
  stepNumber: number;
  title: string;
  recordType: string;
  recordName: string;
  recordValue: string;
  verified: boolean | null;
  verifying: boolean;
  onVerify: () => void;
  manualOnly?: boolean;
  description?: string;
}

export function DnsRecordStep({
  stepNumber,
  title,
  recordType,
  recordName,
  recordValue,
  verified,
  verifying,
  onVerify,
  manualOnly = false,
  description,
}: DnsRecordStepProps) {
  const [copied, setCopied] = useState(false);

  const fullRecord = `${recordName} IN ${recordType} ${recordValue}`;

  async function handleCopy() {
    await navigator.clipboard.writeText(fullRecord);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div
      className="rounded-lg border border-border bg-card p-5"
      data-testid={`dns-step-${stepNumber}`}
    >
      <div className="flex items-center gap-3 mb-3">
        <span
          className={cn(
            "flex size-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold",
            verified === true
              ? "bg-emerald-500/15 text-emerald-500"
              : verified === false
                ? "bg-destructive/15 text-destructive"
                : "bg-muted text-muted-foreground"
          )}
        >
          {verified === true ? (
            <Check className="size-3.5" />
          ) : verified === false ? (
            <X className="size-3.5" />
          ) : (
            stepNumber
          )}
        </span>
        <div className="flex flex-col">
          <h3 className="text-sm font-medium text-foreground">{title}</h3>
          <span className="text-xs text-muted-foreground">{recordType} Record</span>
        </div>
      </div>

      {description && (
        <p className="mb-3 text-xs text-muted-foreground">{description}</p>
      )}

      <div className="relative rounded-md border border-border bg-muted/50 p-4">
        <div className="flex items-start justify-between gap-3">
          <pre className="flex-1 overflow-x-auto whitespace-pre-wrap break-all font-mono text-xs text-foreground">
            {fullRecord}
          </pre>
          <Button
            variant="ghost"
            size="icon-xs"
            onClick={handleCopy}
            aria-label={copied ? "Copied" : `Copy ${title} record`}
            className="shrink-0"
          >
            {copied ? (
              <Check className="size-3 text-emerald-500" />
            ) : (
              <Copy className="size-3" />
            )}
          </Button>
        </div>
        {copied && (
          <span className="absolute right-2 top-10 text-xs text-emerald-500">
            Copied!
          </span>
        )}
      </div>

      {!manualOnly && (
        <div className="mt-3 flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={onVerify}
            disabled={verifying}
            data-testid={`verify-step-${stepNumber}`}
          >
            {verifying ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <Check className="size-3.5" />
            )}
            Check DNS
          </Button>
          {verified === true && (
            <span className="text-xs font-medium text-emerald-500">Verified</span>
          )}
          {verified === false && (
            <span className="text-xs font-medium text-destructive">Not found</span>
          )}
        </div>
      )}

      {manualOnly && (
        <p className="mt-3 text-xs text-amber-500">
          This record must be configured by your hosting provider. Verification is manual.
        </p>
      )}
    </div>
  );
}
