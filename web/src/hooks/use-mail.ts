"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient, type PaginationParams } from "@/lib/api-client";
import type { SendMailRequest } from "@/types/api";

export function useMailFolders(userEmail: string) {
  return useQuery({
    queryKey: ["mailFolders", userEmail],
    queryFn: () => apiClient.getMailFolders(userEmail),
    enabled: userEmail.length > 0,
  });
}

export function useInbox(userId: number, params?: PaginationParams, email?: string) {
  return useQuery({
    queryKey: ["inbox", userId, params, email],
    queryFn: () => apiClient.getInbox(userId, params, email),
    enabled: userId > 0,
  });
}

export function useFolderMessages(folder: string, userEmail: string, params?: PaginationParams) {
  return useQuery({
    queryKey: ["folderMessages", folder, userEmail, params],
    queryFn: () => apiClient.getFolderMessages(folder, userEmail, params),
    enabled: folder.length > 0 && userEmail.length > 0,
  });
}

export function useThreads(userEmail: string, folder: string, params?: PaginationParams) {
  return useQuery({
    queryKey: ["threads", folder, userEmail, params],
    queryFn: () => apiClient.getThreads(folder, userEmail, params),
    enabled: folder.length > 0 && userEmail.length > 0,
  });
}

export function useMessage(userId: number, messageId: number, email?: string) {
  return useQuery({
    queryKey: ["message", userId, messageId, email],
    queryFn: () => apiClient.getMessage(userId, messageId, email),
    enabled: userId > 0 && messageId > 0,
  });
}

export function useSendMail() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: SendMailRequest) => apiClient.sendMail(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inbox"] });
      queryClient.invalidateQueries({ queryKey: ["mailFolders"] });
    },
  });
}

export function useMailQueue() {
  return useQuery({
    queryKey: ["mailQueue"],
    queryFn: () => apiClient.getMailQueue(),
    refetchInterval: 30_000,
  });
}

export function useCreateFolder(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (folderName: string) => apiClient.createFolder(userEmail, folderName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["mailFolders", userEmail] });
    },
  });
}

export function useDeleteFolder(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (folderName: string) => apiClient.deleteFolder(userEmail, folderName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["mailFolders", userEmail] });
    },
  });
}

export function useRenameFolder(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ oldName, newName }: { oldName: string; newName: string }) =>
      apiClient.renameFolder(userEmail, oldName, newName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["mailFolders", userEmail] });
    },
  });
}

export function useMoveMessage(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ uid, fromFolder, toFolder }: { uid: number; fromFolder: string; toFolder: string }) =>
      apiClient.moveMessage(userEmail, uid, fromFolder, toFolder),
    onSuccess: (_data, { fromFolder }) => {
      queryClient.invalidateQueries({ queryKey: ["folderMessages", fromFolder, userEmail] });
      queryClient.invalidateQueries({ queryKey: ["mailFolders", userEmail] });
    },
  });
}
