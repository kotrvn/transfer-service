"use client";

import { useMutation } from "@tanstack/react-query";
import { type FormEvent, useRef, useState } from "react";

import { api, loginErrorMessage } from "../../shared/api";
import styles from "./sign-in.module.scss";

const roleLabels = { passenger: "Пассажир", driver: "Водитель" } as const;
type Role = keyof typeof roleLabels;

type Errors = Partial<Record<"role" | "email" | "password", string>>;

export function SignInForm() {
  const [role, setRole] = useState<Role>();
  const [errors, setErrors] = useState<Errors>({});
  const [message, setMessage] = useState("");
  const submitting = useRef(false);

  const { mutateAsync, isPending, reset } = useMutation({
    mutationFn: async (credentials: { email: string; password: string }) => {
      await api.post("/api/auth/login", credentials);
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
    if (!role) nextErrors.role = "Выберите роль";
    if (!email || emailInput.validity.typeMismatch) nextErrors.email = "Введите корректный email";
    if (!password) nextErrors.password = "Введите пароль";

    setErrors(nextErrors);
    setMessage("");
    if (!role || nextErrors.email || nextErrors.password) {
      const firstInvalid = Object.keys(nextErrors)[0];
      form.querySelector<HTMLInputElement>(`input[name="${firstInvalid}"]`)?.focus();
      return;
    }

    submitting.current = true;
    try {
      await mutateAsync({ email, password });
      passwordInput.value = "";
      setMessage("Данные подтверждены. Сессия появится на следующем этапе");
    } catch (error) {
      setMessage(loginErrorMessage(error));
    } finally {
      reset();
      submitting.current = false;
    }
  }

  return (
    <form className={styles.form} method="post" onSubmit={submit} noValidate>
      <fieldset
        className={styles.roleGroup}
        disabled={isPending}
        aria-invalid={!!errors.role}
        aria-describedby={errors.role ? "role-error" : undefined}
      >
        <legend className={styles.label}>Роль</legend>
        <div className={styles.roles}>
          {(Object.keys(roleLabels) as Role[]).map((option) => (
            <label className={styles.role} key={option}>
              <input
                type="radio"
                name="role"
                value={option}
                checked={role === option}
                required
                aria-describedby={errors.role ? "role-error" : undefined}
                onChange={() => {
                  setRole(option);
                  setErrors((current) => ({ ...current, role: undefined }));
                }}
              />
              {roleLabels[option]}
            </label>
          ))}
        </div>
        {errors.role && <p className={styles.error} id="role-error">{errors.role}</p>}
      </fieldset>

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
