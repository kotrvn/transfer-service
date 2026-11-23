import type { ReactNode } from "react";
import styles from "./auth-card.module.scss";

export function AuthCard({ title, children, footer }: {
  title: string;
  children: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <main className={styles.screen}>
      <section className={styles.card} aria-labelledby="auth-title">
        <p className={styles.brand}>Transfer Service</p>
        <h1 className={styles.title} id="auth-title">{title}</h1>
        {children}
        {footer && <p className={styles.footer}>{footer}</p>}
      </section>
    </main>
  );
}
