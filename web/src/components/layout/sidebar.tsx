"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Globe,
  Users,
  ArrowLeftRight,
  Key,
  ScrollText,
  Mail,
  Shield,
  PanelLeftClose,
  PanelLeft,
  Lock,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";

const NAV_ITEMS = [
  { label: "Dashboard", href: "/", icon: LayoutDashboard, shortcut: "1" },
  { label: "Domains", href: "/domains", icon: Globe, shortcut: "2" },
  { label: "Users", href: "/users", icon: Users, shortcut: "3" },
  { label: "Aliases", href: "/aliases", icon: ArrowLeftRight, shortcut: "4" },
  { label: "DKIM", href: "/dkim", icon: Key, shortcut: "5" },
  { label: "Logs", href: "/logs", icon: ScrollText, shortcut: "6" },
  { label: "Webmail", href: "/mail", icon: Mail, shortcut: "7" },
  { label: "Production", href: "/production", icon: Shield, shortcut: "8" },
] as const;

export { NAV_ITEMS };

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const pathname = usePathname();

  return (
    <aside
      data-testid="sidebar"
      className={cn(
        "flex h-full flex-col border-r border-sidebar-border bg-sidebar transition-[width] duration-200 ease-in-out",
        collapsed ? "w-14" : "w-56"
      )}
    >
      <div className="flex h-14 items-center justify-between px-3">
        {!collapsed && (
          <Link
            href="/"
            className="flex items-center gap-2 text-sm font-semibold text-sidebar-foreground"
          >
            <Mail className="size-5 text-primary" />
            <span>Oxmail</span>
          </Link>
        )}
        <Button
          variant="ghost"
          size="icon-sm"
          onClick={onToggle}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          className={cn(
            "text-muted-foreground hover:text-sidebar-foreground",
            collapsed && "mx-auto"
          )}
        >
          {collapsed ? (
            <PanelLeft className="size-4" />
          ) : (
            <PanelLeftClose className="size-4" />
          )}
        </Button>
      </div>

      <Separator className="bg-sidebar-border" />

      <ScrollArea className="flex-1 py-2">
        <nav className="flex flex-col gap-1 px-2">
          {NAV_ITEMS.map((item) => {
            const isActive =
              item.href === "/"
                ? pathname === "/"
                : pathname.startsWith(item.href);

            const linkContent = (
              <Link
                href={item.href}
                aria-current={isActive ? "page" : undefined}
                className={cn(
                  "relative flex items-center gap-3 rounded-md px-2.5 py-2 text-sm font-medium transition-all duration-150",
                  isActive
                    ? "bg-sidebar-accent text-primary"
                    : "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground"
                )}
              >
                {isActive && (
                  <span className="absolute left-0.5 top-1/2 h-4 w-0.5 -translate-y-1/2 rounded-full bg-primary animate-[fade-in_150ms_ease-out]" />
                )}
                <item.icon className="size-4 shrink-0" />
                {!collapsed && <span>{item.label}</span>}
              </Link>
            );

            if (collapsed) {
              return (
                <Tooltip key={item.href} delayDuration={0}>
                  <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
                  <TooltipContent side="right" sideOffset={8}>
                    {item.label}
                  </TooltipContent>
                </Tooltip>
              );
            }

            return <div key={item.href}>{linkContent}</div>;
          })}
        </nav>
      </ScrollArea>

      <Separator className="bg-sidebar-border" />

      <div className="px-2 py-2">
        {(() => {
          const isActive = pathname === "/account/change-password";
          const linkContent = (
            <Link
              href="/account/change-password"
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "relative flex items-center gap-3 rounded-md px-2.5 py-2 text-sm font-medium transition-all duration-150",
                isActive
                  ? "bg-sidebar-accent text-primary"
                  : "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground"
              )}
            >
              {isActive && (
                <span className="absolute left-0.5 top-1/2 h-4 w-0.5 -translate-y-1/2 rounded-full bg-primary animate-[fade-in_150ms_ease-out]" />
              )}
              <Lock className="size-4 shrink-0" />
              {!collapsed && <span>Change Password</span>}
            </Link>
          );

          if (collapsed) {
            return (
              <Tooltip delayDuration={0}>
                <TooltipTrigger asChild>{linkContent}</TooltipTrigger>
                <TooltipContent side="right" sideOffset={8}>
                  Change Password
                </TooltipContent>
              </Tooltip>
            );
          }

          return linkContent;
        })()}
      </div>
    </aside>
  );
}
