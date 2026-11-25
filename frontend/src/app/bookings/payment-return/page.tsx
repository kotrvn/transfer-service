import { AuthCard } from "../../../shared/ui/auth-card";

export default async function PaymentReturnPage({
  searchParams,
}: {
  searchParams: Promise<{ result?: string }>;
}) {
  const { result } = await searchParams;

  return (
    <AuthCard title={result === "success" ? "Проверяем оплату" : "Оплата прервана"}>
      <p>{result === "success"
        ? "Вернитесь к бронированиям в исходной вкладке. Статус обновится после подтверждения Stripe."
        : "Вернитесь к бронированиям в исходной вкладке. Если бронь ещё действует, оплату можно открыть снова."}</p>
    </AuthCard>
  );
}
