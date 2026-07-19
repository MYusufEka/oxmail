"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Loader2 } from "lucide-react";
import { useAuth } from "@/contexts/auth";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Sidebar, NAV_ITEMS } from "@/components/layout/sidebar";
import { Topbar } from "@/components/layout/topbar";
import { CommandPalette } from "@/components/command-palette";

export function AppShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const { user, isLoading, mustChangePassword } = useAuth();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);

  const handleToggleSidebar = useCallback(() => {
    setSidebarCollapsed((prev) => !prev);
  }, []);

  const isLoginRoute = pathname === "/login";
  const isChangePasswordRoute = pathname === "/account/change-password";

  useEffect(() => {
    if (isLoading) {
      return;
    }

    if (!user && !isLoginRoute) {
      router.replace("/login");
      return;
    }

    if (user && mustChangePassword && !isChangePasswordRoute) {
      router.replace("/account/change-password");
    }
  }, [isChangePasswordRoute, isLoading, isLoginRoute, mustChangePassword, router, user]);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      // Cmd+K or Ctrl+K → open command palette
      if ((event.metaKey || event.ctrlKey) && event.key === "k") {
        event.preventDefault();
        setCommandPaletteOpen((prev) => !prev);
        return;
      }

      // Cmd+1-7 or Ctrl+1-7 → navigate
      if ((event.metaKey || event.ctrlKey) && event.key >= "1" && event.key <= "7") {
        event.preventDefault();
        const index = parseInt(event.key, 10) - 1;
        const navItem = NAV_ITEMS[index];
        if (navItem) {
          router.push(navItem.href);
        }
        return;
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [router]);

  if (isLoginRoute) {
    return <>{children}</>;
  }

  if (isLoading || !user) {
    return (
      <div className="flex min-h-dvh items-center justify-center bg-background">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <TooltipProvider delayDuration={0}>
      <div className="flex h-screen overflow-hidden bg-background">
        <Sidebar collapsed={sidebarCollapsed} onToggle={handleToggleSidebar} />
        <div className="flex flex-1 flex-col overflow-hidden">
          <Topbar onOpenCommandPalette={() => setCommandPaletteOpen(true)} />
          <main className="flex-1 overflow-y-auto p-6">{children}</main>
        </div>
      </div>
      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
      />
    </TooltipProvider>
  );
}
