"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, type PaginationParams } from "@/lib/api-client";
import type { User, CreateUserRequest, UpdateUserRequest, PaginatedResponse, UserImportResult } from "@/types/api";

export function useUsers(domainId: number, params?: PaginationParams) {
  return useQuery({
    queryKey: ["users", domainId, params],
    queryFn: () => apiClient.getUsers(domainId, params),
    enabled: domainId > 0,
  });
}

export function useImportUsers(domainId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (file: File) => apiClient.importUsers(domainId, file),
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users", domainId] });
    },
  });
}

export function useCreateUser(domainId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateUserRequest) => apiClient.createUser(domainId, payload),
    onMutate: async (newUser) => {
      await queryClient.cancelQueries({ queryKey: ["users", domainId] });

      const previousUsers = queryClient.getQueriesData<PaginatedResponse<User>>({
        queryKey: ["users", domainId],
      });

      queryClient.setQueriesData<PaginatedResponse<User>>(
        { queryKey: ["users", domainId] },
        (old) => {
          if (!old) return old;
          const optimistic: User = {
            id: Date.now(),
            email: newUser.email,
            domainId,
            displayName: newUser.displayName,
            quota: newUser.quota ?? 0,
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

      return { previousUsers };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousUsers) {
        for (const [queryKey, data] of context.previousUsers) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users", domainId] });
    },
  });
}

export function useDeleteUser(domainId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: number) => apiClient.deleteUser(domainId, userId),
    onMutate: async (userId) => {
      await queryClient.cancelQueries({ queryKey: ["users", domainId] });

      const previousUsers = queryClient.getQueriesData<PaginatedResponse<User>>({
        queryKey: ["users", domainId],
      });

      queryClient.setQueriesData<PaginatedResponse<User>>(
        { queryKey: ["users", domainId] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.filter((user) => user.id !== userId),
            pagination: { ...old.pagination, total: old.pagination.total - 1 },
          };
        },
      );

      return { previousUsers };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousUsers) {
        for (const [queryKey, data] of context.previousUsers) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users", domainId] });
    },
  });
}

export function useUpdateUser(domainId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ userId, payload }: { userId: number; payload: UpdateUserRequest }) =>
      apiClient.updateUser(domainId, userId, payload),
    onMutate: async ({ userId, payload }) => {
      await queryClient.cancelQueries({ queryKey: ["users", domainId] });

      const previousUsers = queryClient.getQueriesData<PaginatedResponse<User>>({
        queryKey: ["users", domainId],
      });

      queryClient.setQueriesData<PaginatedResponse<User>>(
        { queryKey: ["users", domainId] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            data: old.data.map((user) =>
              user.id === userId
                ? {
                    ...user,
                    displayName: payload.displayName ?? user.displayName,
                    quota: payload.quota ?? user.quota,
                    updatedAt: new Date().toISOString(),
                  }
                : user,
            ),
          };
        },
      );

      return { previousUsers };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousUsers) {
        for (const [queryKey, data] of context.previousUsers) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users", domainId] });
    },
  });
}
