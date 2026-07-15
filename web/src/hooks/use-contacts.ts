"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import type {
  Contact,
  CreateContactRequest,
  UpdateContactRequest,
} from "@/types/api";

export function useContacts(userEmail: string) {
  return useQuery({
    queryKey: ["contacts", userEmail],
    queryFn: () => apiClient.getContacts(userEmail),
    enabled: userEmail.length > 0,
  });
}

export function useCreateContact(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateContactRequest) =>
      apiClient.createContact(userEmail, payload),
    onMutate: async (newContact) => {
      await queryClient.cancelQueries({ queryKey: ["contacts", userEmail] });

      const previousContacts = queryClient.getQueryData<Contact[]>([
        "contacts",
        userEmail,
      ]);

      queryClient.setQueryData<Contact[]>(["contacts", userEmail], (old) => {
        if (!old) return old;
        const optimistic: Contact = {
          id: Date.now(),
          userEmail,
          name: newContact.name,
          email: newContact.email,
          phone: newContact.phone,
          createdAt: new Date().toISOString(),
        };
        return [...old, optimistic];
      });

      return { previousContacts };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousContacts) {
        queryClient.setQueryData(["contacts", userEmail], context.previousContacts);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts", userEmail] });
    },
  });
}

export function useUpdateContact(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      contactId,
      payload,
    }: {
      contactId: number;
      payload: UpdateContactRequest;
    }) => apiClient.updateContact(userEmail, contactId, payload),
    onMutate: async ({ contactId, payload }) => {
      await queryClient.cancelQueries({ queryKey: ["contacts", userEmail] });

      const previousContacts = queryClient.getQueryData<Contact[]>([
        "contacts",
        userEmail,
      ]);

      queryClient.setQueryData<Contact[]>(["contacts", userEmail], (old) => {
        if (!old) return old;
        return old.map((c) =>
          c.id === contactId
            ? {
                ...c,
                name: payload.name ?? c.name,
                email: payload.email ?? c.email,
                phone: payload.phone ?? c.phone,
              }
            : c,
        );
      });

      return { previousContacts };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousContacts) {
        queryClient.setQueryData(["contacts", userEmail], context.previousContacts);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts", userEmail] });
    },
  });
}

export function useDeleteContact(userEmail: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (contactId: number) =>
      apiClient.deleteContact(userEmail, contactId),
    onMutate: async (contactId) => {
      await queryClient.cancelQueries({ queryKey: ["contacts", userEmail] });

      const previousContacts = queryClient.getQueryData<Contact[]>([
        "contacts",
        userEmail,
      ]);

      queryClient.setQueryData<Contact[]>(["contacts", userEmail], (old) => {
        if (!old) return old;
        return old.filter((c) => c.id !== contactId);
      });

      return { previousContacts };
    },
    onError: (_error, _variables, context) => {
      if (context?.previousContacts) {
        queryClient.setQueryData(["contacts", userEmail], context.previousContacts);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts", userEmail] });
    },
  });
}
