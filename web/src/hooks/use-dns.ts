"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import type { DNSRecord, DNSCheckResult } from "@/types/api";

export function useDnsRecords() {
  return useQuery({
    queryKey: ["dns", "records"],
    queryFn: () =>
      apiClient.getDnsRecords() as Promise<{ records: DNSRecord[] }>,
  });
}

interface UseDnsCheckOptions {
  enabled?: boolean;
}

export function useDnsCheck({ enabled = false }: UseDnsCheckOptions = {}) {
  return useQuery({
    queryKey: ["dns", "check"],
    queryFn: () =>
      apiClient.getDnsCheck() as Promise<{ results: DNSCheckResult[] }>,
    enabled,
  });
}
