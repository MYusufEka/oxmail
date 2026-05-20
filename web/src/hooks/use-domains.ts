"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, type PaginationParams } from "@/lib/api-client";
import type { Domain, CreateDomainRequest, PaginatedResponse } from "@/types/api";

export function useDomains(params?: PaginationParams) {
  return useQuery({
    queryKey: ["domains", params],
    queryFn: () => apiClient.getDomains(params),
  });
}

export function useCreateDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateDomainRequest) => apiClient.createDomain(payload),
    onMutate: async (newDomain) => {
      await queryClient.cancelQueries({ queryKey: ["domains"] });

      const previousDomains = queryClient.getQueriesData<PaginatedResponse<Domain>>({
        queryKey: ["domains"],
      });

      queryClient.setQueriesData<PaginatedResponse<Domain>>(
        { queryKey: ["domains"] },
        (old) => {
          if (!old) return old;
          const optimistic: Domain = {
            id: Date.now(),
            name: newDomain.name,
            active: true,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          };
          return {
            ...old,
            data: [...old.data, optimistic],
            pagination: { ...old.pagination, total: old.pagination.total + 1 },
          };
        },
      );

      return { previousDomains };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousDomains) {
        for (const [queryKey, data] of context.previousDomains) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["domains"] });
    },
  });
}

export function useDeleteDomain() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (domainId: number) => apiClient.deleteDomain(domainId),
    onMutate: async (domainId) => {
      await queryClient.cancelQueries({ queryKey: ["domains"] });

      const previousDomains = queryClient.getQueriesData<PaginatedResponse<Domain>>({
        queryKey: ["domains"],
      });

      queryClient.setQueriesData<PaginatedResponse<Domain>>(
        { queryKey: ["domains"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.filter((domain) => domain.id !== domainId),
            pagination: { ...old.pagination, total: old.pagination.total - 1 },
          };
        },
      );

      return { previousDomains };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousDomains) {
        for (const [queryKey, data] of context.previousDomains) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["domains"] });
    },
  });
}
