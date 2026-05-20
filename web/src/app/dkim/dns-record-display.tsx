"use client";

import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { Button } from "@/components/ui/button";

interface DnsRecordDisplayProps {
  selector: string;
  domain: string;
  dnsRecord: string;
}

export function DnsRecordDisplay({
  selector,
  domain,
  dnsRecord,
}: DnsRecordDisplayProps) {
  const [copied, setCopied] = useState(false);

  const fullRecord = `${selector}._domainkey.${domain} IN TXT "${dnsRecord}"`;

  async function handleCopy() {
    await navigator.clipboard.writeText(fullRecord);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div className="relative rounded-md border border-border bg-muted/50 p-4">
      <div className="flex items-start justify-between gap-3">
        <pre className="flex-1 overflow-x-auto whitespace-pre-wrap break-all font-mono text-xs text-foreground">
          {fullRecord}
        </pre>
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={handleCopy}
          aria-label={copied ? "Copied" : "Copy DNS record"}
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
  );
}
