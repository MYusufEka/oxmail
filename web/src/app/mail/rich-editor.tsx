"use client";

import { useRef, useCallback } from "react";
import {
  Bold,
  Italic,
  Underline,
  Link,
  List,
  ListOrdered,
} from "lucide-react";
import { cn } from "@/lib/utils";

interface RichEditorProps {
  value: string;
  onChange: (html: string) => void;
  placeholder?: string;
  className?: string;
}

interface ToolbarAction {
  icon: typeof Bold;
  command: string;
  argument?: string;
  label: string;
}

const TOOLBAR_ACTIONS: ToolbarAction[] = [
  { icon: Bold, command: "bold", label: "Bold" },
  { icon: Italic, command: "italic", label: "Italic" },
  { icon: Underline, command: "underline", label: "Underline" },
  { icon: Link, command: "createLink", label: "Insert link" },
  { icon: List, command: "insertUnorderedList", label: "Bullet list" },
  { icon: ListOrdered, command: "insertOrderedList", label: "Numbered list" },
];

export function RichEditor({
  value,
  onChange,
  placeholder = "Write your message...",
  className,
}: RichEditorProps) {
  const editorRef = useRef<HTMLDivElement>(null);

  const execCommand = useCallback((command: string) => {
    if (command === "createLink") {
      const url = prompt("Enter URL:");
      if (url) {
        document.execCommand(command, false, url);
      }
    } else {
      document.execCommand(command, false);
    }
    editorRef.current?.focus();
  }, []);

  const handleInput = useCallback(() => {
    const html = editorRef.current?.innerHTML ?? "";
    onChange(html);
  }, [onChange]);

  return (
    <div
      className={cn(
        "flex flex-col rounded-md border border-input focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50",
        className
      )}
    >
      <div className="flex items-center gap-0.5 border-b border-border px-2 py-1.5">
        {TOOLBAR_ACTIONS.map((action) => (
          <button
            key={action.command}
            type="button"
            onClick={() => execCommand(action.command)}
            className="rounded-sm p-1.5 text-muted-foreground hover:bg-secondary hover:text-foreground"
            aria-label={action.label}
            title={action.label}
          >
            <action.icon className="size-4" />
          </button>
        ))}
      </div>
      <div
        ref={editorRef}
        contentEditable
        role="textbox"
        aria-label="Email body"
        aria-multiline="true"
        className="min-h-[200px] flex-1 overflow-y-auto px-3 py-2 text-sm text-foreground outline-none [&:empty]:before:pointer-events-none [&:empty]:before:text-muted-foreground [&:empty]:before:content-[attr(data-placeholder)]"
        data-placeholder={placeholder}
        onInput={handleInput}
        dangerouslySetInnerHTML={{ __html: value }}
      />
    </div>
  );
}
