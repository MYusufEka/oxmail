"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Eye, EyeOff, Plus } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCreateUser } from "@/hooks/use-users";
import { useDomains } from "@/hooks/use-domains";
import type { Domain } from "@/types/api";

const addUserSchema = z.object({
  localPart: z
    .string()
    .min(1, "Email local part is required")
    .regex(/^[a-zA-Z0-9._%+-]+$/, "Invalid email characters"),
  domainId: z.string().min(1, "Domain is required"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  quota: z.string().optional(),
});

type AddUserFormValues = z.infer<typeof addUserSchema>;

function getPasswordStrength(password: string): {
  label: string;
  color: string;
  width: string;
} {
  if (password.length === 0) return { label: "", color: "", width: "w-0" };

  let score = 0;
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[0-9]/.test(password)) score++;
  if (/[^A-Za-z0-9]/.test(password)) score++;

  if (score <= 2) return { label: "Weak", color: "bg-destructive", width: "w-1/3" };
  if (score <= 3) return { label: "Medium", color: "bg-yellow-500", width: "w-2/3" };
  return { label: "Strong", color: "bg-green-500", width: "w-full" };
}

interface AddUserDialogProps {
  selectedDomainId: number;
  domains: Domain[];
}

export function AddUserDialog({ selectedDomainId, domains }: AddUserDialogProps) {
  const [open, setOpen] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  const activeDomainId = selectedDomainId > 0 ? selectedDomainId : (domains[0]?.id ?? 0);
  const createUser = useCreateUser(activeDomainId);

  const form = useForm<AddUserFormValues>({
    resolver: zodResolver(addUserSchema),
    defaultValues: {
      localPart: "",
      domainId: activeDomainId > 0 ? String(activeDomainId) : "",
      password: "",
      quota: "",
    },
  });

  const watchedPassword = form.watch("password");
  const strength = getPasswordStrength(watchedPassword);

  function onSubmit(values: AddUserFormValues) {
    const domainId = Number(values.domainId);
    const domain = domains.find((d) => d.id === domainId);
    if (!domain) return;

    const email = `${values.localPart}@${domain.name}`;
    const quotaBytes = values.quota ? Number(values.quota) * 1024 * 1024 : undefined;

    createUser.mutate(
      { email, password: values.password, quota: quotaBytes },
      {
        onSuccess: () => {
          toast.success(`User ${email} created`);
          setOpen(false);
          form.reset();
          setShowPassword(false);
        },
        onError: (error) => {
          toast.error(`Failed to create user: ${error.message}`);
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Plus className="size-4" />
          Add User
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add User</DialogTitle>
          <DialogDescription>
            Create a new mailbox user for a domain.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="grid gap-4">
            <div className="grid gap-4 sm:grid-cols-[1fr_auto_1fr]">
              <FormField
                control={form.control}
                name="localPart"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Email</FormLabel>
                    <FormControl>
                      <Input placeholder="username" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <span className="flex items-end pb-2 text-muted-foreground">@</span>
              <FormField
                control={form.control}
                name="domainId"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Domain</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select domain" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {domains.map((domain) => (
                          <SelectItem
                            key={domain.id}
                            value={String(domain.id)}
                          >
                            {domain.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl>
                    <div className="relative">
                      <Input
                        type={showPassword ? "text" : "password"}
                        placeholder="Minimum 8 characters"
                        {...field}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        className="absolute top-1/2 right-2 -translate-y-1/2"
                        onClick={() => setShowPassword(!showPassword)}
                        aria-label={showPassword ? "Hide password" : "Show password"}
                      >
                        {showPassword ? (
                          <EyeOff className="size-4" />
                        ) : (
                          <Eye className="size-4" />
                        )}
                      </Button>
                    </div>
                  </FormControl>
                  {watchedPassword.length > 0 && (
                    <div className="flex items-center gap-2">
                      <div className="h-1 flex-1 rounded-full bg-muted">
                        <div
                          className={`h-full rounded-full transition-all ${strength.color} ${strength.width}`}
                        />
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {strength.label}
                      </span>
                    </div>
                  )}
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="quota"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Quota (MB)</FormLabel>
                  <FormControl>
                    <Input
                      type="number"
                      placeholder="Leave empty for unlimited"
                      min={0}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type="submit"
                disabled={createUser.isPending}
              >
                {createUser.isPending ? "Creating..." : "Create User"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
