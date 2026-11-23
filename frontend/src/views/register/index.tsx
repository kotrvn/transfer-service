import Link from "next/link";
import { SignUpForm } from "../../features/sign-up";
import { AuthCard } from "../../shared/ui/auth-card";

export function RegisterView() {
  return (
    <AuthCard title="Регистрация" footer={<>Уже есть аккаунт? <Link href="/login">Войти</Link></>}>
      <SignUpForm />
    </AuthCard>
  );
}
