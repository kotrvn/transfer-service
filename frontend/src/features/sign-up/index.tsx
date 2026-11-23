"use client";

import { useMutation } from "@tanstack/react-query";
import { type FormEvent, useEffect, useRef, useState } from "react";

import { roleLabels, type Role } from "../../entities/session";
import { api, registrationErrorMessage } from "../../shared/api";
import styles from "../../shared/ui/auth-form/auth-form.module.scss";

type Field = "role" | "name" | "email" | "password";
type Errors = Partial<Record<Field, string>>;
type Registration = { role: Role; name: string; email: string; password: string };

export function SignUpForm() {
  const [errors, setErrors] = useState<Errors>({});
  const [message, setMessage] = useState("");
  const [created, setCreated] = useState(false);
  const submitting = useRef(false);
  const status = useRef<HTMLDivElement>(null);
  const { mutateAsync, isPending, reset } = useMutation({
    mutationFn: async (data: Registration) => {
      await api.post("/api/auth/register", data);
    },
    retry: false,
    networkMode: "always",
    gcTime: 0,
  });

  useEffect(() => {
    if (created) status.current?.focus();
  }, [created]);

  function clearError(field: Field) {
    if (errors[field]) setErrors((current) => ({ ...current, [field]: undefined }));
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting.current) return;

    const form = event.currentTarget;
    const data = new FormData(form);
    const choice = data.get("role");
    const role = choice === "passenger" || choice === "driver" ? choice : null;
    const name = String(data.get("name") ?? "").trim();
    const emailInput = form.elements.namedItem("email") as HTMLInputElement;
    const email = emailInput.value.trim();
    const password = String(data.get("password") ?? "");
    const nextErrors: Errors = {};

    if (!role) nextErrors.role = "Выберите роль";
    const nameLength = Array.from(name).length;
    if (nameLength < 2 || nameLength > 100) nextErrors.name = "Введите имя длиной от 2 до 100 символов";
    if (!email || emailInput.validity.typeMismatch) nextErrors.email = "Введите корректный email";
    if (Array.from(password).length < 6) nextErrors.password = "Введите не менее 6 символов";
    else if (new TextEncoder().encode(password).length > 72) nextErrors.password = "Введите более короткий пароль";

    setErrors(nextErrors);
    setMessage("");
    const firstInvalid = Object.keys(nextErrors)[0];
    if (firstInvalid || !role) {
      if (firstInvalid) form.querySelector<HTMLInputElement>(`input[name="${firstInvalid}"]`)?.focus();
      return;
    }

    submitting.current = true;
    try {
      await mutateAsync({ role, name, email, password });
      form.reset();
      setCreated(true);
    } catch (error) {
      setMessage(registrationErrorMessage(error));
    } finally {
      reset();
      submitting.current = false;
    }
  }

  return (
    <>
      <div ref={status} className={created ? undefined : styles.status} role="status" tabIndex={created ? -1 : undefined}>
        {created ? "Аккаунт создан. Теперь можно войти." : isPending ? "Создаём аккаунт…" : ""}
      </div>
      {!created && (
        <form className={styles.form} method="post" onSubmit={submit} noValidate>
          <fieldset
            className={styles.roleGroup}
            disabled={isPending}
            aria-invalid={!!errors.role}
            aria-describedby={errors.role ? "register-role-error" : undefined}
          >
            <legend className={styles.label}>Роль</legend>
            <div className={styles.roles}>
              {(Object.keys(roleLabels) as Role[]).map((role) => (
                <label className={styles.role} key={role}>
                  <input
                    type="radio"
                    name="role"
                    value={role}
                    required
                    aria-describedby={errors.role ? "register-role-error" : undefined}
                    onChange={() => clearError("role")}
                  />
                  {roleLabels[role]}
                </label>
              ))}
            </div>
            {errors.role && <p className={styles.error} id="register-role-error">{errors.role}</p>}
          </fieldset>

          <div>
            <label className={styles.label} htmlFor="register-name">Имя</label>
            <input
              className={styles.input}
              id="register-name"
              name="name"
              autoComplete="name"
              required
              readOnly={isPending}
              aria-invalid={!!errors.name}
              aria-describedby={errors.name ? "register-name-error" : undefined}
              onChange={() => clearError("name")}
            />
            {errors.name && <p className={styles.error} id="register-name-error">{errors.name}</p>}
          </div>

          <div>
            <label className={styles.label} htmlFor="register-email">Email</label>
            <input
              className={styles.input}
              id="register-email"
              name="email"
              type="email"
              autoComplete="username"
              required
              readOnly={isPending}
              aria-invalid={!!errors.email}
              aria-describedby={errors.email ? "register-email-error" : undefined}
              onChange={() => clearError("email")}
            />
            {errors.email && <p className={styles.error} id="register-email-error">{errors.email}</p>}
          </div>

          <div>
            <label className={styles.label} htmlFor="register-password">Пароль</label>
            <input
              className={styles.input}
              id="register-password"
              name="password"
              type="password"
              autoComplete="new-password"
              required
              readOnly={isPending}
              aria-invalid={!!errors.password}
              aria-describedby={errors.password ? "register-password-error" : undefined}
              onChange={() => clearError("password")}
            />
            {errors.password && <p className={styles.error} id="register-password-error">{errors.password}</p>}
          </div>

          <div>
            <button className={styles.submit} type="submit" disabled={isPending}>
              {isPending ? "Создаём аккаунт…" : "Создать аккаунт"}
            </button>
            {message && <p className={styles.error} role="alert">{message}</p>}
          </div>
        </form>
      )}
    </>
  );
}
