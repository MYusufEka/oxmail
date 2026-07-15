"use client";

import { useState, useCallback } from "react";
import { toast } from "sonner";
import { Copy, Check, KeyRound } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useUpdateUser } from "@/hooks/use-users";
import type { User } from "@/types/api";

const PASSWORD_LENGTH = 16;
const PASSWORD_CHARS = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%&*";

function generatePassword(): string {
  const randomValues = new Uint32Array(PASSWORD_LENGTH);
  crypto.getRandomValues(randomValues);
  let password = "";
  for (let i = 0; i < PASSWORD_LENGTH; i++) {
    password += PASSWORD_CHARS[randomValues[i] % PASSWORD_CHARS.length];
  }
  return password;
}

interface ResetPasswordDialogProps {
  user: User | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function ResetPasswordDialog({
  user,
  open,
  onOpenChange,
  onSuccess,
}: ResetPasswordDialogProps) {
  const [generatedPassword, setGeneratedPassword] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const updateUser = useUpdateUser(user?.domainId ?? 0);

  const handleReset = useCallback(() => {
    if (!user) return;

    const newPassword = generatePassword();

    updateUser.mutate(
      { userId: user.id, payload: { password: newPassword } },
      {
        onSuccess: () => {
          setGeneratedPassword(newPassword);
          toast.success(`Password reset for ${user.email}`);
        },
        onError: (error) => {
          toast.error(`Failed to reset password: ${error.message}`);
        },
      },
    );
  }, [user, updateUser]);

  async function handleCopy() {
    if (!generatedPassword) return;
    try {
      await navigator.clipboard.writeText(generatedPassword);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Failed to copy to clipboard");
    }
  }

  function handleClose() {
    setGeneratedPassword(null);
    setCopied(false);
    onOpenChange(false);
    onSuccess();
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          handleClose();
        }
      }}
    >
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Password Reset</DialogTitle>
          <DialogDescription>
            New password for {user?.email}
          </DialogDescription>
        </DialogHeader>

        {generatedPassword === null ? (
          <div className="flex flex-col items-center gap-4 py-4">
            <p className="text-sm text-muted-foreground text-center">
              This will generate a random password and set it immediately.
              The new password will only be shown once.
            </p>
            <Button
              onClick={handleReset}
              disabled={updateUser.isPending}
            >
              <KeyRound className="mr-2 size-4" />
              {updateUser.isPending ? "Resetting..." : "Generate & Reset"}
            </Button>
          </div>
        ) : (
          <div className="flex flex-col gap-4 py-4">
            <div className="rounded-lg border border-border bg-muted/50 p-4">
              <p className="text-xs text-muted-foreground mb-2">
                Copy this password now. It will not be shown again.
              </p>
              <div className="flex items-center gap-2">
                <code className="flex-1 break-all rounded bg-muted px-3 py-2 font-mono text-sm text-foreground">
                  {generatedPassword}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={handleCopy}
                  aria-label="Copy password to clipboard"
                >
                  {copied ? (
                    <Check className="size-4 text-emerald-500" />
                  ) : (
                    <Copy className="size-4" />
                  )}
                </Button>
              </div>
            </div>
          </div>
        )}

        <DialogFooter>
          {generatedPassword !== null && (
            <Button onClick={handleClose}>
              Done
            </Button>
          )}
          {generatedPassword === null && (
            <Button variant="outline" onClick={handleClose}>
              Cancel
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
