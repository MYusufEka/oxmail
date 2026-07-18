"use client";

import { useState, useCallback, useEffect } from "react";
import { useAuth } from "@/contexts/auth";
import { redirect } from "next/navigation";
import { Signature, Save, Loader2 } from "lucide-react";
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

interface StoredSignature {
  enabled: boolean;
  content: string;
}

function getSignatureKey(email: string): string {
  return `signature:${email}`;
}

function loadSignature(email: string): StoredSignature {
  if (typeof window === "undefined") return { enabled: false, content: "" };
  try {
    const raw = localStorage.getItem(getSignatureKey(email));
    if (!raw) return { enabled: false, content: "" };
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    return {
      enabled: Boolean(parsed.enabled),
      content: typeof parsed.content === "string" ? parsed.content : "",
    };
  } catch {
    return { enabled: false, content: "" };
  }
}

function saveSignature(email: string, data: StoredSignature) {
  localStorage.setItem(getSignatureKey(email), JSON.stringify(data));
}

export function getSignatureForEmail(email: string): StoredSignature | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = localStorage.getItem(`signature:${email}`);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    if (!parsed.enabled || typeof parsed.content !== "string" || parsed.content.length === 0) {
      return null;
    }
    return { enabled: true, content: parsed.content };
  } catch {
    return null;
  }
}

export default function SignaturePage() {
  const { user, isLoading: authLoading } = useAuth();
  const userEmail = user?.email ?? "";
  const [enabled, setEnabled] = useState(false);
  const [content, setContent] = useState("");
  const [saving, setSaving] = useState(false);
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    if (userEmail && !initialized) {
      const sig = loadSignature(userEmail);
      setEnabled(sig.enabled);
      setContent(sig.content);
      setInitialized(true);
    }
  }, [userEmail, initialized]);

  const handleSave = useCallback(() => {
    setSaving(true);
    try {
      saveSignature(userEmail, { enabled, content });
      toast.success("Signature saved");
    } catch (err) {
      toast.error("Failed to save", {
        description: err instanceof Error ? err.message : "Unknown error",
      });
    } finally {
      setSaving(false);
    }
  }, [enabled, content, userEmail]);

  if (authLoading) {
    return (
      <div className="flex h-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <Signature className="size-5 text-primary" />
          <h2 className="text-lg font-semibold text-foreground">Email Signature</h2>
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="flex flex-col gap-3">
              <div className="h-8 w-48 animate-pulse rounded bg-muted" />
              <div className="h-24 w-full animate-pulse rounded bg-muted" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!user) {
    redirect("/login");
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
                onCheckedChange={setEnabled}
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
                onChange={(e) => setContent(e.target.value)}
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
