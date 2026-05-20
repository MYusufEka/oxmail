"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, type PaginationParams } from "@/lib/api-client";
import type { SendMailRequest } from "@/types/api";

export function useInbox(userId: number, params?: PaginationParams) {
  return useQuery({
    queryKey: ["inbox", userId, params],
    queryFn: () => apiClient.getInbox(userId, params),
    enabled: userId > 0,
  });
}

export function useMessage(userId: number, messageId: number) {
  return useQuery({
    queryKey: ["message", userId, messageId],
    queryFn: () => apiClient.getMessage(userId, messageId),
    enabled: userId > 0 && messageId > 0,
  });
}

export function useSendMail() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: SendMailRequest) => apiClient.sendMail(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inbox"] });
    },
  });
}
