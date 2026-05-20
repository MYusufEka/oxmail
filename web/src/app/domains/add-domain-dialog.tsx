"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

const addDomainSchema = z.object({
  name: z
    .string()
    .min(3, "Domain must be at least 3 characters")
    .regex(
      /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/,
      "Enter a valid domain name (e.g. example.com)"
    ),
});

type AddDomainFormValues = z.infer<typeof addDomainSchema>;

interface AddDomainDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (name: string) => void;
  isPending: boolean;
}

export function AddDomainDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
}: AddDomainDialogProps) {
  const form = useForm<AddDomainFormValues>({
    resolver: zodResolver(addDomainSchema),
    defaultValues: { name: "" },
  });

  function handleSubmit(values: AddDomainFormValues) {
    onSubmit(values.name);
    form.reset();
  }

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      form.reset();
    }
    onOpenChange(nextOpen);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent data-testid="add-domain-dialog">
        <DialogHeader>
          <DialogTitle>Add Domain</DialogTitle>
          <DialogDescription>
            Add a new mail domain to your server.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="grid gap-4"
            data-testid="add-domain-form"
          >
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Domain name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="example.com"
                      autoFocus
                      disabled={isPending}
                      data-testid="domain-name-input"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage data-testid="domain-name-error" />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
                disabled={isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isPending} data-testid="add-domain-submit">
                {isPending && <Loader2 className="animate-spin" />}
                Add Domain
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
