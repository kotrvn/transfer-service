"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { endSession, roleLabels, useSession } from "../../entities/session";
import styles from "./account.module.scss";

export function AccountView() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const session = useSession((state) => state.session);

  useEffect(() => {
    if (!session) router.replace("/login");
  }, [router, session]);

  if (!session) return null;

  return (
    <main className={styles.screen}>
      <section className={styles.card} aria-labelledby="account-title">
        <h1 id="account-title">Аккаунт</h1>
        <p>Роль: {roleLabels[session.role]}</p>
        <button type="button" onClick={() => endSession(queryClient)}>Выйти</button>
      </section>
    </main>
  );
}
