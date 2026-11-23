"use client";

import { useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { type ReactNode, useEffect } from "react";

import { endSession, roleLabels, useSession } from "../../entities/session";
import styles from "./shell.module.scss";

export function CabinetShell({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const queryClient = useQueryClient();
  const session = useSession((state) => state.session);

  useEffect(() => {
    if (!session) router.replace("/login");
  }, [router, session]);

  if (!session) return null;

  return (
    <div className={styles.shell}>
      <aside className={styles.sidebar}>
        <span className={styles.brand}><span aria-hidden="true">TS</span><b>Transfer Service</b></span>
        <nav aria-label="Кабинет">
          <Link href="/dashboard" aria-label={session.role === "passenger" ? "Поиск поездок" : "Главная"} aria-current={pathname === "/dashboard" ? "page" : undefined}>
            <span aria-hidden="true">⌂</span><b>{session.role === "passenger" ? "Поиск поездок" : "Главная"}</b>
          </Link>
          {session.role === "passenger" && (
            <Link href="/bookings" aria-label="Мои бронирования" aria-current={pathname === "/bookings" ? "page" : undefined}>
              <span aria-hidden="true">▤</span><b>Мои бронирования</b>
            </Link>
          )}
        </nav>
      </aside>
      <header className={styles.header}>
        <span>Кабинет</span>
        <div>
          <span className={styles.role}>{roleLabels[session.role]}</span>
          <button type="button" onClick={() => endSession(queryClient)}>Выйти</button>
        </div>
      </header>
      <main className={styles.content}>{children}</main>
    </div>
  );
}
