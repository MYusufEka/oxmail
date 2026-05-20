"use client";

import { useState, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import { DnsRecordStep } from "./dns-record-step";
import type { DNSCheckResult } from "@/types/api";

interface DnsStepConfig {
  title: string;
  recordType: string;
  recordName: string;
  recordValue: string;
  recordKey: string;
  manualOnly?: boolean;
  description?: string;
}

const DOMAIN_PLACEHOLDER = "yourdomain.com";

function getDnsSteps(domain: string): DnsStepConfig[] {
  const host = domain || DOMAIN_PLACEHOLDER;
  return [
    {
      title: "MX Record",
      recordType: "MX",
      recordName: host,
      recordValue: `10 mail.${host}`,
      recordKey: "mx",
      description: "Routes incoming email to your mail server.",
    },
    {
      title: "SPF Record",
      recordType: "TXT",
      recordName: host,
      recordValue: `"v=spf1 mx a:mail.${host} ~all"`,
      recordKey: "spf",
      description: "Authorizes your server to send email for this domain.",
    },
    {
      title: "DKIM Record",
      recordType: "TXT",
      recordName: `default._domainkey.${host}`,
      recordValue: `"v=DKIM1; k=rsa; p=..."`,
      recordKey: "dkim",
      description:
        "Cryptographically signs outgoing email. The full key is generated in the DKIM page.",
    },
    {
      title: "DMARC Record",
      recordType: "TXT",
      recordName: `_dmarc.${host}`,
      recordValue: `"v=DMARC1; p=quarantine; rua=mailto:postmaster@${host}"`,
      recordKey: "dmarc",
      description: "Tells receivers how to handle unauthenticated email from your domain.",
    },
    {
      title: "rDNS / PTR Record",
      recordType: "PTR",
      recordName: "your.public.ip",
      recordValue: `mail.${host}`,
      recordKey: "rdns",
      manualOnly: true,
      description:
        "Reverse DNS must be set by your hosting provider to map your IP back to your mail hostname.",
    },
  ];
}

export function DnsWizard({ domain }: { domain: string }) {
  const queryClient = useQueryClient();
  const [verificationState, setVerificationState] = useState<
    Record<string, boolean | null>
  >({
    mx: null,
    spf: null,
    dkim: null,
    dmarc: null,
    rdns: null,
  });
  const [verifyingStep, setVerifyingStep] = useState<string | null>(null);

  const steps = getDnsSteps(domain);

  const verifiedCount = Object.values(verificationState).filter(
    (v) => v === true
  ).length;
  const totalSteps = steps.length;
  const allVerified = verifiedCount === totalSteps;
  const hasPartial = verifiedCount > 0 && !allVerified;

  const handleVerify = useCallback(
    async (recordKey: string) => {
      setVerifyingStep(recordKey);
      try {
        const response = await apiClient.getDnsCheck();
        const results = response.results;

        const newState: Record<string, boolean | null> = { ...verificationState };
        for (const result of results) {
          newState[result.record] = result.valid;
        }

        // If the specific record wasn't in the response, mark as failed
        if (!results.some((r: DNSCheckResult) => r.record === recordKey)) {
          newState[recordKey] = false;
        }

        setVerificationState(newState);
        queryClient.invalidateQueries({ queryKey: ["dns", "check"] });
      } catch {
        setVerificationState((prev) => ({ ...prev, [recordKey]: false }));
      } finally {
        setVerifyingStep(null);
      }
    },
    [verificationState, queryClient]
  );

  return (
    <div className="flex flex-col gap-5" data-testid="dns-wizard">
      {/* Progress indicator */}
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium text-foreground">
            DNS Readiness
          </span>
          <span
            className="text-xs font-medium"
            data-testid="dns-progress-text"
          >
            <span
              className={
                allVerified
                  ? "text-emerald-500"
                  : hasPartial
                    ? "text-amber-500"
                    : "text-muted-foreground"
              }
            >
              {verifiedCount}/{totalSteps} records verified
            </span>
          </span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
          <div
            className={`h-full rounded-full transition-all duration-300 ${
              allVerified
                ? "bg-emerald-500"
                : hasPartial
                  ? "bg-amber-500"
                  : "bg-muted-foreground/30"
            }`}
            style={{ width: `${(verifiedCount / totalSteps) * 100}%` }}
            role="progressbar"
            aria-valuenow={verifiedCount}
            aria-valuemin={0}
            aria-valuemax={totalSteps}
            aria-label="DNS verification progress"
            data-testid="dns-progress-bar"
          />
        </div>
      </div>

      {/* Steps */}
      <div className="flex flex-col gap-4">
        {steps.map((step, index) => (
          <DnsRecordStep
            key={step.recordKey}
            stepNumber={index + 1}
            title={step.title}
            recordType={step.recordType}
            recordName={step.recordName}
            recordValue={step.recordValue}
            verified={verificationState[step.recordKey]}
            verifying={verifyingStep === step.recordKey}
            onVerify={() => handleVerify(step.recordKey)}
            manualOnly={step.manualOnly}
            description={step.description}
          />
        ))}
      </div>
    </div>
  );
}
