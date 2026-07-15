"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, ApiError } from "@/lib/api-client";
import type { DKIMKey } from "@/types/api";

export function useDkim(domain: string) {
  return useQuery({
    queryKey: ["dkim", domain],
    queryFn: () => apiClient.getDkim(domain),
    enabled: domain.length > 0,
    retry: (_failCount: number, err: Error) =>
      !(err instanceof ApiError && err.status === 404),
  });
}

export function useGenerateDkim() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (domain: string) => apiClient.generateDkim(domain),
    onSuccess: (newKey: DKIMKey) => {
      queryClient.setQueryData(["dkim", newKey.domain], newKey);
      queryClient.invalidateQueries({ queryKey: ["dkim", newKey.domain] });
    },
  });
}
