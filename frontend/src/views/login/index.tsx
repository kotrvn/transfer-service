import Link from "next/link";
import { SignInForm } from "../../features/sign-in";
import { AuthCard } from "../../shared/ui/auth-card";

export function LoginView() {
  return (
    <AuthCard title="Вход" footer={<>Нет аккаунта? <Link href="/register">Зарегистрироваться</Link></>}>
      <SignInForm />
    </AuthCard>
  );
}
