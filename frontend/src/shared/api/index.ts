import axios from "axios";

export const api = axios.create({ timeout: 10_000 });

export function loginErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    if (error.response?.status === 400) return "Проверьте email и пароль";
    if (error.response?.status === 401) return "Неверный email или пароль";
  }
  return "Не удалось войти. Попробуйте ещё раз";
}
