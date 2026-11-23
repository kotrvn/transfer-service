"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { type ReactNode, useEffect, useState } from "react";

import { endSession, useSession } from "../entities/session";

export function Providers({ children }: { children: ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());
  const expiresAt = useSession((state) => state.session?.expiresAt);

  useEffect(() => {
    const expiration = expiresAt;
    if (expiration === undefined) return;
    let timer: ReturnType<typeof setTimeout>;

    const checkExpiration = () => {
      clearTimeout(timer);
      if (useSession.getState().session?.expiresAt !== expiration) return;
      const remaining = expiration - Date.now();
      if (remaining <= 0) {
        endSession(queryClient);
      } else {
        timer = setTimeout(checkExpiration, Math.min(remaining, 2_147_483_647));
      }
    };

    checkExpiration();
    window.addEventListener("focus", checkExpiration);
    return () => {
      clearTimeout(timer);
      window.removeEventListener("focus", checkExpiration);
    };
  }, [expiresAt, queryClient]);

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
