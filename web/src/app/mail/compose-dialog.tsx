"use client";

import { useState, useCallback } from "react";
import { Paperclip, Send, Loader2, X } from "lucide-react";
import { toast } from "sonner";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useSendMail } from "@/hooks/use-mail";
import { RecipientInput } from "./recipient-input";
import { RichEditor } from "./rich-editor";

interface ComposeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialTo?: string[];
  initialSubject?: string;
  currentUserEmail: string;
}

interface Attachment {
  name: string;
  size: number;
}

export function ComposeDialog({
  open,
  onOpenChange,
  initialTo = [],
  initialSubject = "",
  currentUserEmail,
}: ComposeDialogProps) {
  const [to, setTo] = useState<string[]>(initialTo);
  const [cc, setCc] = useState<string[]>([]);
  const [bcc, setBcc] = useState<string[]>([]);
  const [subject, setSubject] = useState(initialSubject);
  const [body, setBody] = useState("");
  const [showCcBcc, setShowCcBcc] = useState(false);
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [showDiscardConfirm, setShowDiscardConfirm] = useState(false);

  const sendMail = useSendMail();

  const hasContent = body.trim().length > 0 || subject.trim().length > 0 || to.length > 0;

  const resetForm = useCallback(() => {
    setTo([]);
    setCc([]);
    setBcc([]);
    setSubject("");
    setBody("");
    setShowCcBcc(false);
    setAttachments([]);
  }, []);

  const handleClose = useCallback(() => {
    if (hasContent) {
      setShowDiscardConfirm(true);
    } else {
      onOpenChange(false);
    }
  }, [hasContent, onOpenChange]);

  const handleDiscard = useCallback(() => {
    setShowDiscardConfirm(false);
    resetForm();
    onOpenChange(false);
  }, [resetForm, onOpenChange]);

  const handleSend = useCallback(() => {
    if (to.length === 0) {
      toast.error("Please add at least one recipient");
      return;
    }

    sendMail.mutate(
      {
        from: currentUserEmail,
        to,
        cc: cc.length > 0 ? cc : undefined,
        subject,
        bodyHtml: body,
        bodyText: body.replace(/<[^>]*>/g, ""),
      },
      {
        onSuccess: () => {
          toast.success("Message sent");
          resetForm();
          onOpenChange(false);
        },
        onError: () => {
          toast.error("Failed to send message");
        },
      }
    );
  }, [to, cc, subject, body, sendMail, resetForm, onOpenChange, currentUserEmail]);

  const handleAttachment = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const files = event.target.files;
      if (!files) return;

      const newAttachments: Attachment[] = Array.from(files).map((file) => ({
        name: file.name,
        size: file.size,
      }));

      setAttachments((prev) => [...prev, ...newAttachments]);
      event.target.value = "";
    },
    []
  );

  const removeAttachment = useCallback((index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index));
  }, []);

  return (
    <>
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent
          className="flex h-[90vh] max-h-[700px] w-full max-w-2xl flex-col gap-0 p-0"
          showCloseButton={false}
        >
          <DialogHeader className="flex flex-row items-center justify-between border-b border-border px-4 py-3">
            <DialogTitle className="text-base">New Message</DialogTitle>
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={handleClose}
              aria-label="Close compose"
            >
              <X className="size-4" />
            </Button>
          </DialogHeader>

          <div className="flex flex-1 flex-col gap-3 overflow-y-auto px-4 py-3">
            <RecipientInput
              label="To"
              recipients={to}
              onChange={setTo}
            />

            {showCcBcc ? (
              <>
                <RecipientInput
                  label="Cc"
                  recipients={cc}
                  onChange={setCc}
                />
                <RecipientInput
                  label="Bcc"
                  recipients={bcc}
                  onChange={setBcc}
                />
              </>
            ) : (
              <button
                type="button"
                onClick={() => setShowCcBcc(true)}
                className="self-start text-xs text-muted-foreground hover:text-foreground"
              >
                Add Cc/Bcc
              </button>
            )}

            <div className="flex flex-col gap-1">
              <label className="text-xs font-medium text-muted-foreground">
                Subject
              </label>
              <Input
                value={subject}
                onChange={(event) => setSubject(event.target.value)}
                placeholder="Subject"
              />
            </div>

            <RichEditor
              value={body}
              onChange={setBody}
              className="flex-1"
            />

            {attachments.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {attachments.map((file, index) => (
                  <span
                    key={`${file.name}-${index}`}
                    className="inline-flex items-center gap-1.5 rounded-sm bg-secondary px-2 py-1 text-xs text-secondary-foreground"
                  >
                    <Paperclip className="size-3" />
                    {file.name}
                    <button
                      type="button"
                      onClick={() => removeAttachment(index)}
                      className="text-muted-foreground hover:text-foreground"
                      aria-label={`Remove ${file.name}`}
                    >
                      <X className="size-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>

          <div className="flex items-center justify-between border-t border-border px-4 py-3">
            <div className="flex items-center gap-2">
              <Button
                onClick={handleSend}
                disabled={sendMail.isPending || to.length === 0}
                size="sm"
              >
                {sendMail.isPending ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Send className="size-4" />
                )}
                Send
              </Button>

              <label className="cursor-pointer">
                <input
                  type="file"
                  multiple
                  className="hidden"
                  onChange={handleAttachment}
                />
                <span className="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-sm text-muted-foreground hover:bg-secondary hover:text-foreground">
                  <Paperclip className="size-4" />
                  Attach
                </span>
              </label>
            </div>

            <Button
              variant="ghost"
              size="sm"
              onClick={handleClose}
              className="text-muted-foreground"
            >
              Discard
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={showDiscardConfirm} onOpenChange={setShowDiscardConfirm}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Discard draft?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Your message has unsaved content. Are you sure you want to discard
            it?
          </p>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowDiscardConfirm(false)}
            >
              Keep editing
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={handleDiscard}
            >
              Discard
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
