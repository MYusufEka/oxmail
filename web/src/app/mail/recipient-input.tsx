"use client";

import { useState, useCallback, type KeyboardEvent } from "react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

interface RecipientInputProps {
  label: string;
  recipients: string[];
  onChange: (recipients: string[]) => void;
  placeholder?: string;
  className?: string;
}

export function RecipientInput({
  label,
  recipients,
  onChange,
  placeholder = "email@example.com",
  className,
}: RecipientInputProps) {
  const [inputValue, setInputValue] = useState("");
  const [error, setError] = useState<string | null>(null);

  const addRecipient = useCallback(
    (value: string) => {
      const trimmed = value.trim();
      if (!trimmed) return;

      if (!EMAIL_REGEX.test(trimmed)) {
        setError(`Invalid email: ${trimmed}`);
        return;
      }

      if (recipients.includes(trimmed)) {
        setError(`Already added: ${trimmed}`);
        return;
      }

      setError(null);
      onChange([...recipients, trimmed]);
      setInputValue("");
    },
    [recipients, onChange]
  );

  const removeRecipient = useCallback(
    (index: number) => {
      onChange(recipients.filter((_, i) => i !== index));
    },
    [recipients, onChange]
  );

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter" || event.key === "," || event.key === "Tab") {
      event.preventDefault();
      addRecipient(inputValue);
    }

    if (
      event.key === "Backspace" &&
      inputValue === "" &&
      recipients.length > 0
    ) {
      removeRecipient(recipients.length - 1);
    }
  };

  const handleBlur = () => {
    if (inputValue.trim()) {
      addRecipient(inputValue);
    }
  };

  return (
    <div className={cn("flex flex-col gap-1", className)}>
      <label className="text-xs font-medium text-muted-foreground">
        {label}
      </label>
      <div className="flex flex-wrap items-center gap-1.5 rounded-md border border-input bg-transparent px-2 py-1.5 focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50">
        {recipients.map((email, index) => (
          <span
            key={`${email}-${index}`}
            className="inline-flex items-center gap-1 rounded-sm bg-secondary px-2 py-0.5 text-xs text-secondary-foreground"
          >
            {email}
            <button
              type="button"
              onClick={() => removeRecipient(index)}
              className="rounded-xs text-muted-foreground hover:text-foreground"
              aria-label={`Remove ${email}`}
            >
              <X className="size-3" />
            </button>
          </span>
        ))}
        <input
          type="email"
          value={inputValue}
          onChange={(event) => {
            setInputValue(event.target.value);
            setError(null);
          }}
          onKeyDown={handleKeyDown}
          onBlur={handleBlur}
          placeholder={recipients.length === 0 ? placeholder : ""}
          className="min-w-[120px] flex-1 bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
        />
      </div>
      {error && (
        <p className="text-xs text-destructive" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}
