"use client";

import { useState, useCallback } from "react";
import { useAuth } from "@/contexts/auth";
import { redirect } from "next/navigation";
import { Signature, Save, Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import type { SignatureResponse } from "@/types/api";

export default function SignaturePage() {
  const { user, isLoading: authLoading } = useAuth();
  const userEmail = user?.email ?? "";
  const queryClient = useQueryClient();
  const [enabledOverride, setEnabledOverride] = useState<boolean | null>(null);
  const [contentOverride, setContentOverride] = useState<string | null>(null);

  const {
    data: signatureData,
    isLoading: signatureLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ["signature", userEmail],
    queryFn: () => apiClient.getSignature(userEmail),
    enabled: userEmail.length > 0,
  });

  const upsertSignatureMutation = useMutation({
    mutationFn: ({ email, signature }: { email: string; signature: SignatureResponse }) =>
      apiClient.upsertSignature(email, {
        content: signature.content,
        enabled: signature.enabled,
      }),
    onSuccess: (savedSignature) => {
      queryClient.setQueryData(["signature", savedSignature.email], savedSignature);
    },
  });

  const deleteSignatureMutation = useMutation({
    mutationFn: (email: string) => apiClient.deleteSignature(email),
    onSuccess: (_deleted, email) => {
      queryClient.setQueryData(["signature", email], {
        email,
        content: "",
        enabled: false,
      } satisfies SignatureResponse);
    },
  });

  const enabled = enabledOverride ?? signatureData?.enabled ?? false;
  const content = contentOverride ?? signatureData?.content ?? "";

  const handleSave = useCallback(async () => {
    try {
      if (!enabled) {
        await deleteSignatureMutation.mutateAsync(userEmail);
        setContentOverride("");
        setEnabledOverride(false);
        toast.success("Signature disabled");
        return;
      }

      await upsertSignatureMutation.mutateAsync({
        email: userEmail,
        signature: { email: userEmail, enabled, content },
      });
      toast.success("Signature saved");
    } catch (err) {
      toast.error("Failed to save", {
        description: err instanceof Error ? err.message : "Unknown error",
      });
    }
  }, [enabled, content, userEmail, upsertSignatureMutation, deleteSignatureMutation]);

  const saving = upsertSignatureMutation.isPending || deleteSignatureMutation.isPending;

  if (authLoading || signatureLoading) {
    return (
      <div className="flex h-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <Signature className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Email Signature</h2>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col gap-3">
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-24 w-full" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!user) {
    redirect("/login");
  }

  if (isError) {
    return (
      <div className="flex h-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <Signature className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Email Signature</h2>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center p-12">
            <div className="flex size-12 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="size-6 text-destructive" />
            </div>
            <h3 className="mt-4 text-base font-medium text-foreground">Failed to load signature</h3>
            <p className="mt-1 text-sm text-muted-foreground">Something went wrong. Please try again.</p>
            <Button variant="outline" className="mt-6" onClick={() => refetch()}>
              <RefreshCw className="size-4" />
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="flex items-center gap-3">
        <Signature className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Email Signature</h2>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Signature settings</CardTitle>
              <CardDescription>
                Automatically append a signature to outgoing messages.
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="signature-enabled"
                checked={enabled}
                onCheckedChange={setEnabledOverride}
                data-testid="signature-toggle"
                aria-label="Toggle email signature"
              />
              <Label htmlFor="signature-enabled" className="text-sm">
                {enabled ? "Enabled" : "Disabled"}
              </Label>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-5">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">
                Signature content
              </label>
              <textarea
                value={content}
                onChange={(event) => setContentOverride(event.target.value)}
                placeholder={"-- \nJohn Doe\nEngineering Lead\nAcme Inc."}
                disabled={!enabled}
                rows={4}
                data-testid="signature-textarea"
                className="min-h-[100px] w-full resize-y rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none selection:bg-primary selection:text-primary-foreground placeholder:text-muted-foreground disabled:pointer-events-none disabled:opacity-50 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
              />
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSave}
                disabled={saving}
                data-testid="signature-save-btn"
              >
                {saving ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Save className="size-4" />
                )}
                Save
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
