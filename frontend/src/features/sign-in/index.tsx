"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { type FormEvent, useRef, useState } from "react";

import { readSession, useSession } from "../../entities/session";
import { api, loginErrorMessage } from "../../shared/api";
import styles from "../../shared/ui/auth-form/auth-form.module.scss";

type Errors = Partial<Record<"email" | "password", string>>;

export function SignInForm() {
  const router = useRouter();
  const setSession = useSession((state) => state.setSession);
  const [errors, setErrors] = useState<Errors>({});
  const [message, setMessage] = useState("");
  const submitting = useRef(false);

  const { mutateAsync, isPending, reset } = useMutation({
    mutationFn: async (credentials: { email: string; password: string }) => {
      const response = await api.post<{ token?: unknown }>("/api/auth/login", credentials);
      const token = response.data?.token;
      return typeof token === "string" ? readSession(token) : null;
    },
    retry: false,
    networkMode: "always",
    gcTime: 0,
  });

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting.current) return;

    const form = event.currentTarget;
    const emailInput = form.elements.namedItem("email") as HTMLInputElement;
    const passwordInput = form.elements.namedItem("password") as HTMLInputElement;
    const email = emailInput.value.trim();
    const password = passwordInput.value;
    const nextErrors: Errors = {};
    if (!email || emailInput.validity.typeMismatch) nextErrors.email = "Введите корректный email";
    if (!password) nextErrors.password = "Введите пароль";

    setErrors(nextErrors);
    setMessage("");
    if (nextErrors.email || nextErrors.password) {
      const firstInvalid = Object.keys(nextErrors)[0];
      form.querySelector<HTMLInputElement>(`input[name="${firstInvalid}"]`)?.focus();
      return;
    }

    submitting.current = true;
    try {
      const session = await mutateAsync({ email, password });
      if (!session) {
        setMessage("Не удалось проверить сессию. Попробуйте ещё раз");
      } else {
        passwordInput.value = "";
        setSession(session);
        router.replace("/account");
      }
    } catch (error) {
      setMessage(loginErrorMessage(error));
    } finally {
      reset();
      submitting.current = false;
    }
  }

  return (
    <form className={styles.form} method="post" onSubmit={submit} noValidate>
      <div>
        <label className={styles.label} htmlFor="email">Email</label>
        <input
          className={styles.input}
          id="email"
          name="email"
          type="email"
          autoComplete="username"
          required
          readOnly={isPending}
          aria-invalid={!!errors.email}
          aria-describedby={errors.email ? "email-error" : undefined}
          onChange={() => {
            if (errors.email) setErrors((current) => ({ ...current, email: undefined }));
          }}
        />
        {errors.email && <p className={styles.error} id="email-error">{errors.email}</p>}
      </div>

      <div>
        <label className={styles.label} htmlFor="password">Пароль</label>
        <input
          className={styles.input}
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          required
          readOnly={isPending}
          aria-invalid={!!errors.password}
          aria-describedby={errors.password ? "password-error" : undefined}
          onChange={() => {
            if (errors.password) setErrors((current) => ({ ...current, password: undefined }));
          }}
        />
        {errors.password && <p className={styles.error} id="password-error">{errors.password}</p>}
      </div>

      <div>
        <button className={styles.submit} type="submit" disabled={isPending}>
          {isPending ? "Входим…" : "Войти"}
        </button>
        <div className={styles.status} role="status">{isPending ? "Входим…" : ""}</div>
        {message && <p className={styles.error} role="alert">{message}</p>}
      </div>
    </form>
  );
}
