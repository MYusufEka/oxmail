"use client";

import { useState, useEffect, useCallback } from "react";
import { Luggage, Save, Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api-client";
import {
  generateVacationScript,
  parseVacationSettings,
} from "@/lib/sieve-utils";

const DEFAULT_EMAIL = "alice@local.test";

export default function VacationPage() {
  const queryClient = useQueryClient();
  const { data: sieveData, isLoading, isError, refetch } = useQuery({
    queryKey: ["vacation", DEFAULT_EMAIL],
    queryFn: () => apiClient.getVacationScript(DEFAULT_EMAIL),
    enabled: DEFAULT_EMAIL.length > 0,
  });
  const setVacationMutation = useMutation({
    mutationFn: ({ email, script }: { email: string; script: string }) =>
      apiClient.setVacationScript(email, script),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ["vacation", variables.email] });
    },
  });
  const deleteVacationMutation = useMutation({
    mutationFn: (email: string) => apiClient.deleteVacationScript(email),
    onSuccess: (_data, email) => {
      queryClient.invalidateQueries({ queryKey: ["vacation", email] });
    },
  });

  const [enabled, setEnabled] = useState(false);
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [days, setDays] = useState(7);
  const [initialized, setInitialized] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (sieveData && !initialized) {
      const script = sieveData.script ?? "";
      const settings = parseVacationSettings(script);
      setEnabled(settings.enabled);
      setSubject(settings.subject);
      setBody(settings.body);
      setDays(settings.days);
      setInitialized(true);
    }
  }, [sieveData, initialized]);

  const handleSave = useCallback(async () => {
    setSaving(true);
    try {
      if (!enabled) {
        await deleteVacationMutation.mutateAsync(DEFAULT_EMAIL);
        toast.success("Vacation auto-reply disabled");
      } else {
        const vacationScript = generateVacationScript({
          enabled: true,
          subject: subject.trim(),
          body: body.trim(),
          days,
        });
        await setVacationMutation.mutateAsync({ email: DEFAULT_EMAIL, script: vacationScript });
        toast.success("Vacation auto-reply saved");
      }
    } catch (err) {
      toast.error("Failed to save", { description: err instanceof Error ? err.message : "Unknown error" });
    } finally {
      setSaving(false);
    }
  }, [enabled, subject, body, days, setVacationMutation, deleteVacationMutation]);

  if (isLoading) {
    return (
      <div className="flex h-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <Luggage className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Vacation / Auto-reply</h2>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col gap-4">
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-24 w-full" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex h-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <Luggage className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Vacation / Auto-reply</h2>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center p-12">
            <div className="flex size-12 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircle className="size-6 text-destructive" />
            </div>
            <h3 className="mt-4 text-base font-medium text-foreground">Failed to load settings</h3>
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
        <Luggage className="size-5 text-primary" />
        <h2 className="text-lg font-semibold text-foreground">Vacation / Auto-reply</h2>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Auto-reply settings</CardTitle>
              <CardDescription>
                Send automated replies when you&apos;re away.
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="vacation-enabled"
                checked={enabled}
                onCheckedChange={setEnabled}
                aria-label="Toggle vacation auto-reply"
              />
              <Label htmlFor="vacation-enabled" className="text-sm">
                {enabled ? "Enabled" : "Disabled"}
              </Label>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-5">
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">
                Reply subject
              </label>
              <Input
                value={subject}
                onChange={(e) => setSubject(e.target.value)}
                placeholder="e.g. Out of office"
                disabled={!enabled}
                data-testid="vacation-subject"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">
                Reply body
              </label>
              <textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder="I'm currently away from my desk and will respond when I return."
                disabled={!enabled}
                rows={4}
                data-testid="vacation-body"
                className="min-h-[100px] w-full resize-y rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs transition-[color,box-shadow] outline-none selection:bg-primary selection:text-primary-foreground placeholder:text-muted-foreground disabled:pointer-events-none disabled:opacity-50 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted-foreground">
                Days between replies
              </label>
              <Input
                type="number"
                min={1}
                max={90}
                value={days}
                onChange={(e) => setDays(Math.max(1, Math.min(90, Number(e.target.value) || 1)))}
                disabled={!enabled}
                className="w-28"
                data-testid="vacation-days"
              />
              <p className="text-xs text-muted-foreground">
                How often the same sender receives your reply (min 1, max 90).
              </p>
            </div>

            <div className="flex justify-end">
              <Button
                onClick={handleSave}
                disabled={saving || (!enabled && !sieveData?.active)}
                data-testid="save-vacation"
              >
                {saving && <Loader2 className="size-4 animate-spin" />}
                <Save className="size-4" />
                {!enabled ? "Disable & Save" : "Save"}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
