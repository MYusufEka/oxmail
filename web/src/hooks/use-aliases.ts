"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, type PaginationParams } from "@/lib/api-client";
import type { Alias, CreateAliasRequest, PaginatedResponse } from "@/types/api";

export function useAliases(domainId: number, params?: PaginationParams) {
  return useQuery({
    queryKey: ["aliases", domainId, params],
    queryFn: () => apiClient.getAliases(domainId, params),
    enabled: domainId > 0,
  });
}

export function useCreateAlias(domainId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateAliasRequest) => apiClient.createAlias(domainId, payload),
    onMutate: async (newAlias) => {
      await queryClient.cancelQueries({ queryKey: ["aliases", domainId] });

      const previousAliases = queryClient.getQueriesData<PaginatedResponse<Alias>>({
        queryKey: ["aliases", domainId],
      });

      queryClient.setQueriesData<PaginatedResponse<Alias>>(
        { queryKey: ["aliases", domainId] },
        (old) => {
          if (!old) return old;
          const optimistic: Alias = {
            id: Date.now(),
            sourceAddress: newAlias.sourceAddress,
            destinationAddress: newAlias.destinationAddress,
            active: true,
            createdAt: new Date().toISOString(),
          };
          return {
            ...old,
            data: [...old.data, optimistic],
            pagination: { ...old.pagination, total: old.pagination.total + 1 },
          };
        },
      );

      return { previousAliases };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousAliases) {
        for (const [queryKey, data] of context.previousAliases) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["aliases", domainId] });
    },
  });
}

export function useDeleteAlias(domainId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (aliasId: number) => apiClient.deleteAlias(domainId, aliasId),
    onMutate: async (aliasId) => {
      await queryClient.cancelQueries({ queryKey: ["aliases", domainId] });

      const previousAliases = queryClient.getQueriesData<PaginatedResponse<Alias>>({
        queryKey: ["aliases", domainId],
      });

      queryClient.setQueriesData<PaginatedResponse<Alias>>(
        { queryKey: ["aliases", domainId] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.filter((alias) => alias.id !== aliasId),
            pagination: { ...old.pagination, total: old.pagination.total - 1 },
          };
        },
      );

      return { previousAliases };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousAliases) {
        for (const [queryKey, data] of context.previousAliases) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["aliases", domainId] });
    },
  });
}
