"use client";

import { useState } from "react";
import {
  CheckCircle2,
  AlertTriangle,
  Key,
  RotateCw,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useDkim, useGenerateDkim } from "@/hooks/use-dkim";
import { DnsRecordDisplay } from "./dns-record-display";

interface DkimDomainCardProps {
  domain: string;
}

export function DkimDomainCard({ domain }: DkimDomainCardProps) {
  const { data: dkimKey, isLoading } = useDkim(domain);
  const generateDkim = useGenerateDkim();
  const [rotateOpen, setRotateOpen] = useState(false);

  const hasKey = Boolean(dkimKey?.publicKey);

  function handleGenerate() {
    generateDkim.mutate(domain);
  }

  function handleRotate() {
    generateDkim.mutate(domain);
    setRotateOpen(false);
  }

  if (isLoading) {
    return (
      <div className="rounded-lg border border-border bg-card p-6">
        <div className="flex items-center gap-3">
          <div className="size-4 animate-pulse rounded bg-muted" />
          <div className="h-4 w-32 animate-pulse rounded bg-muted" />
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {hasKey ? (
            <CheckCircle2 className="size-4 text-emerald-500" />
          ) : (
            <AlertTriangle className="size-4 text-amber-500" />
          )}
          <h3 className="text-sm font-medium text-foreground">{domain}</h3>
          <span
            className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
              hasKey
                ? "bg-emerald-500/10 text-emerald-500"
                : "bg-amber-500/10 text-amber-500"
            }`}
          >
            {hasKey ? "Active" : "Not Generated"}
          </span>
        </div>

        <div className="flex items-center gap-2">
          {!hasKey && (
            <Button
              size="sm"
              onClick={handleGenerate}
              disabled={generateDkim.isPending}
            >
              {generateDkim.isPending ? (
                <Loader2 className="size-3 animate-spin" />
              ) : (
                <Key className="size-3" />
              )}
              Generate Key
            </Button>
          )}

          {hasKey && (
            <Dialog open={rotateOpen} onOpenChange={setRotateOpen}>
              <DialogTrigger asChild>
                <Button variant="outline" size="sm">
                  <RotateCw className="size-3" />
                  Rotate Key
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Rotate DKIM Key</DialogTitle>
                  <DialogDescription>
                    Rotating the key will invalidate the current DNS record.
                    You&apos;ll need to update your DNS TXT record.
                  </DialogDescription>
                </DialogHeader>
                <div className="rounded-md border border-amber-500/20 bg-amber-500/5 p-3">
                  <p className="text-xs text-amber-500">
                    DNS propagation may take up to 48 hours. During this time,
                    some recipients may reject emails signed with the new key.
                  </p>
                </div>
                <DialogFooter>
                  <Button
                    variant="outline"
                    onClick={() => setRotateOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    variant="destructive"
                    onClick={handleRotate}
                    disabled={generateDkim.isPending}
                  >
                    {generateDkim.isPending ? (
                      <Loader2 className="size-3 animate-spin" />
                    ) : (
                      <RotateCw className="size-3" />
                    )}
                    Confirm Rotate
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
        </div>
      </div>

      {hasKey && dkimKey && (
        <div className="mt-4 space-y-4">
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            <div>
              <p className="text-xs text-muted-foreground">Algorithm</p>
              <p className="text-sm font-medium text-foreground">RSA</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Bits</p>
              <p className="text-sm font-medium text-foreground">2048</p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Selector</p>
              <p className="text-sm font-medium text-foreground">
                {dkimKey.selector}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground">Created</p>
              <p className="text-sm font-medium text-foreground">
                {new Date(dkimKey.createdAt).toLocaleDateString()}
              </p>
            </div>
          </div>

          <div>
            <p className="mb-2 text-xs font-medium text-muted-foreground">
              DNS TXT Record
            </p>
            <DnsRecordDisplay
              selector={dkimKey.selector}
              domain={dkimKey.domain}
              dnsRecord={
                dkimKey.dnsRecord ??
                `v=DKIM1; k=rsa; p=${dkimKey.publicKey}`
              }
            />
          </div>
        </div>
      )}
    </div>
  );
}
