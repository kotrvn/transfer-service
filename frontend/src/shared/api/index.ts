import axios from "axios";

export const api = axios.create({ timeout: 10_000 });

export function loginErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    if (error.response?.status === 400) return "Проверьте email и пароль";
    if (error.response?.status === 401) return "Неверный email или пароль";
  }
  return "Не удалось войти. Попробуйте ещё раз";
}

export function registrationErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    if (error.response?.status === 400) return "Проверьте данные для регистрации";
    if (error.response?.status === 409) {
      return "Не удалось создать аккаунт. Возможно, этот email уже используется";
    }
  }
  return "Не удалось зарегистрироваться. Попробуйте ещё раз";
}
