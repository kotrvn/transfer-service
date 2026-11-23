import { SignInForm } from "../../features/sign-in";
import styles from "./login.module.scss";

export function LoginView() {
  return (
    <main className={styles.screen}>
      <section className={styles.card} aria-labelledby="login-title">
        <p className={styles.brand}>Transfer Service</p>
        <h1 className={styles.title} id="login-title">Вход</h1>
        <SignInForm />
      </section>
    </main>
  );
}
