"use client";

import { useEffect, useRef } from "react";
import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import UnderlineExtension from "@tiptap/extension-underline";
import {
  Bold,
  Italic,
  Underline as UnderlineIcon,
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

interface ToolbarButtonProps {
  isActive: boolean;
  onClick: () => void;
  icon: typeof Bold;
  label: string;
}

function ToolbarButton({ isActive, onClick, icon: Icon, label }: ToolbarButtonProps) {
  return (
    <button
      type="button"
      onMouseDown={(event) => event.preventDefault()}
      onClick={onClick}
      className={cn(
        "rounded-sm p-1.5 transition-colors",
        isActive
          ? "bg-secondary text-foreground"
          : "text-muted-foreground hover:bg-secondary hover:text-foreground"
      )}
      aria-label={label}
      title={label}
    >
      <Icon className="size-4" />
    </button>
  );
}

export function RichEditor({
  value,
  onChange,
  placeholder = "Write your message...",
  className,
}: RichEditorProps) {
  const isExternalChange = useRef(false);

  const editor = useEditor({
    extensions: [
      StarterKit.configure({
        link: { openOnClick: false },
        underline: false,
      }),
      UnderlineExtension,
      Placeholder.configure({ placeholder }),
    ],
    content: value,
    immediatelyRender: false,
    onUpdate: ({ editor: e }) => {
      isExternalChange.current = true;
      onChange(e.getHTML());
    },
  });

  // Sync external value changes (e.g. dialog reset) — skip when onChange fired above.
  useEffect(() => {
    if (editor && isExternalChange.current) {
      isExternalChange.current = false;
      return;
    }
    if (editor && value !== editor.getHTML()) {
      editor.commands.setContent(value, { emitUpdate: false });
    }
  }, [editor, value]);

  if (!editor) return null;

  function insertLink() {
    if (!editor) return;
    const url = window.prompt("Enter URL:");
    if (url) {
      editor.chain().focus().extendMarkRange("link").setLink({ href: url }).run();
    }
  }

  return (
    <div
      className={cn(
        "flex flex-col rounded-md border border-input focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50",
        className
      )}
    >
      <div className="flex items-center gap-0.5 border-b border-border px-2 py-1.5">
        <ToolbarButton
          isActive={editor.isActive("bold")}
          onClick={() => editor.chain().focus().toggleBold().run()}
          icon={Bold}
          label="Bold"
        />
        <ToolbarButton
          isActive={editor.isActive("italic")}
          onClick={() => editor.chain().focus().toggleItalic().run()}
          icon={Italic}
          label="Italic"
        />
        <ToolbarButton
          isActive={editor.isActive("underline")}
          onClick={() => editor.chain().focus().toggleUnderline().run()}
          icon={UnderlineIcon}
          label="Underline"
        />
        <ToolbarButton
          isActive={editor.isActive("link")}
          onClick={insertLink}
          icon={Link}
          label="Insert link"
        />
        <ToolbarButton
          isActive={editor.isActive("bulletList")}
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          icon={List}
          label="Bullet list"
        />
        <ToolbarButton
          isActive={editor.isActive("orderedList")}
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          icon={ListOrdered}
          label="Numbered list"
        />
      </div>
      <EditorContent
        editor={editor}
        className="min-h-[200px] flex-1 overflow-y-auto [&_.tiptap]:px-3 [&_.tiptap]:py-2 [&_.tiptap]:text-sm [&_.tiptap]:text-foreground [&_.tiptap]:outline-none [&_.tiptap]:empty:before:pointer-events-none [&_.tiptap]:empty:before:text-muted-foreground [&_.tiptap]:empty:before:content-[attr(data-placeholder)]"
      />
    </div>
  );
}
