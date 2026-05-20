"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Globe,
  Users,
  ArrowLeftRight,
  Key,
  LayoutDashboard,
  ScrollText,
  Mail,
  Plus,
} from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
} from "@/components/ui/command";

const NAV_COMMANDS = [
  { label: "Dashboard", href: "/", icon: LayoutDashboard, shortcut: "⌘1" },
  { label: "Domains", href: "/domains", icon: Globe, shortcut: "⌘2" },
  { label: "Users", href: "/users", icon: Users, shortcut: "⌘3" },
  { label: "Aliases", href: "/aliases", icon: ArrowLeftRight, shortcut: "⌘4" },
  { label: "DKIM", href: "/dkim", icon: Key, shortcut: "⌘5" },
  { label: "Logs", href: "/logs", icon: ScrollText, shortcut: "⌘6" },
  { label: "Webmail", href: "/mail", icon: Mail, shortcut: "⌘7" },
] as const;

const ACTION_COMMANDS = [
  { label: "Add Domain", href: "/domains", icon: Plus },
  { label: "Add User", href: "/users", icon: Plus },
  { label: "Add Alias", href: "/aliases", icon: Plus },
  { label: "Generate DKIM Key", href: "/dkim", icon: Plus },
] as const;

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const router = useRouter();

  const runCommand = useCallback(
    (command: () => void) => {
      onOpenChange(false);
      command();
    },
    [onOpenChange]
  );

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Command Palette"
      description="Search for pages and actions"
    >
      <div data-testid="command-palette">
        <CommandInput placeholder="Type a command or search..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Navigation">
            {NAV_COMMANDS.map((item) => (
              <CommandItem
                key={item.href}
                onSelect={() => runCommand(() => router.push(item.href))}
              >
                <item.icon className="size-4" />
                <span>{item.label}</span>
                <CommandShortcut>{item.shortcut}</CommandShortcut>
              </CommandItem>
            ))}
          </CommandGroup>
          <CommandSeparator />
          <CommandGroup heading="Actions">
            {ACTION_COMMANDS.map((item) => (
              <CommandItem
                key={item.label}
                onSelect={() => runCommand(() => router.push(item.href))}
              >
                <item.icon className="size-4" />
                <span>{item.label}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </div>
    </CommandDialog>
  );
}
