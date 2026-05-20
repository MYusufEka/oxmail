"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import type { DKIMKey } from "@/types/api";

export function useDkim(domain: string) {
  return useQuery({
    queryKey: ["dkim", domain],
    queryFn: () => apiClient.getDkim(domain),
    enabled: domain.length > 0,
  });
}

export function useGenerateDkim() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (domain: string) => apiClient.generateDkim(domain),
    onSuccess: (newKey: DKIMKey) => {
      queryClient.setQueryData(["dkim", newKey.domain], newKey);
    },
  });
}
