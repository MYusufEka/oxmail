"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";

export function useStats(days: number = 7) {
  return useQuery({
    queryKey: ["stats", days],
    queryFn: () => apiClient.getStats(days),
  });
}

export function useStatsSummary() {
  return useQuery({
    queryKey: ["stats-summary"],
    queryFn: () => apiClient.getStatsSummary(),
  });
}
