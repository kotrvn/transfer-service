import type { QueryClient } from "@tanstack/react-query";
import { create } from "zustand";

export const roleLabels = {
  passenger: "Пассажир",
  driver: "Водитель",
} as const;

export type Role = keyof typeof roleLabels;

type Session = {
  token: string;
  userId: string;
  role: Role;
  expiresAt: number;
};

type SessionState = {
  session: Session | null;
  setSession: (session: Session | null) => void;
};

export const useSession = create<SessionState>((set) => ({
  session: null,
  setSession: (session) => set({ session }),
}));

export function endSession(queryClient: QueryClient) {
  queryClient.removeQueries({ predicate: (query) => query.meta?.private === true });
  useSession.getState().setSession(null);
}

export function readSession(token: string): Session | null {
  const parts = token.split(".");
  if (parts.length !== 3 || parts.some((part) => !/^[A-Za-z0-9_-]+$/.test(part))) {
    return null;
  }

  try {
    const payload: unknown = JSON.parse(atob(parts[1].replace(/-/g, "+").replace(/_/g, "/")));
    if (
      !payload || typeof payload !== "object" ||
      !("sub" in payload) || typeof payload.sub !== "string" || !payload.sub.trim() ||
      !("role" in payload) || (payload.role !== "passenger" && payload.role !== "driver") ||
      !("exp" in payload) || typeof payload.exp !== "number"
    ) {
      return null;
    }

    const expiresAt = payload.exp * 1000;
    if (!Number.isSafeInteger(expiresAt) || expiresAt <= Date.now()) return null;
    return { token, userId: payload.sub, role: payload.role, expiresAt };
  } catch {
    return null;
  }
}
