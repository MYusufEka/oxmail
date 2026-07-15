"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";

export function useSieveScript(email: string) {
  return useQuery({
    queryKey: ["sieve", email],
    queryFn: () => apiClient.getSieveScript(email),
    enabled: email.length > 0,
  });
}

export function useSetSieveScript() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ email, script }: { email: string; script: string }) =>
      apiClient.setSieveScript(email, script),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["sieve", variables.email] });
    },
  });
}

export function useDeleteSieveScript() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (email: string) => apiClient.deleteSieveScript(email),
    onSuccess: (_data, email) => {
      queryClient.invalidateQueries({ queryKey: ["sieve", email] });
    },
  });
}
