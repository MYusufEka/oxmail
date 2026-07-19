"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useRouter } from "next/navigation";
import { apiClient, ApiError } from "@/lib/api-client";

interface AuthUser {
  email: string;
  role?: string;
  mustChangePassword: boolean;
}

interface AuthContextValue {
  user: AuthUser | null;
  email: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  mustChangePassword: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const session = await apiClient.me();
      setUser({
        email: session.email,
        role: session.role,
        mustChangePassword: session.mustChangePassword,
      });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        setUser(null);
        return;
      }
      setUser(null);
      throw error;
    }
  }, []);

  useEffect(() => {
    let active = true;

    async function loadSession() {
      try {
        const session = await apiClient.me();
        if (active) {
          setUser({
            email: session.email,
            role: session.role,
            mustChangePassword: session.mustChangePassword,
          });
        }
      } catch {
        if (active) {
          setUser(null);
        }
      } finally {
        if (active) {
          setIsLoading(false);
        }
      }
    }

    void loadSession();

    return () => {
      active = false;
    };
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const response = await apiClient.login({ email, password });
      setUser({
        email: response.email,
        role: response.role,
        mustChangePassword: response.mustChangePassword,
      });
      router.push(response.mustChangePassword ? "/account/change-password" : "/");
    },
    [router],
  );

  const logout = useCallback(async () => {
    await apiClient.logout();
    setUser(null);
    router.push("/login");
  }, [router]);

  const value = useMemo(
    () => ({
      user,
      email: user?.email ?? null,
      isAuthenticated: Boolean(user),
      isLoading,
      mustChangePassword: user?.mustChangePassword ?? false,
      login,
      logout,
      refresh,
    }),
    [user, isLoading, login, logout, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
