"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { useEffect, useState } from "react";

import { useSession } from "../../entities/session";
import { completeTrip, driverBalanceKey, driverTripsKey, getDriverBalance, listDriverTrips, tripDateFormat, tripPriceFormat } from "../../entities/trip";
import { CreateTripDialog } from "../../features/create-trip";
import styles from "./dashboard.module.scss";
import { DriverBookings } from "./driver-bookings";
import { PassengerDashboard } from "./passenger";

function DriverTrips({ driverId, token }: { driverId: string; token: string }) {
  const [creating, setCreating] = useState(false);
  const [confirming, setConfirming] = useState<string>();
  const [now, setNow] = useState(Date.now);
  const queryClient = useQueryClient();
  const trips = useQuery({
    queryKey: driverTripsKey(driverId),
    queryFn: ({ signal }) => listDriverTrips(driverId, signal),
    retry: 1,
    meta: { private: true },
  });
  const balance = useQuery({
    queryKey: driverBalanceKey(driverId),
    queryFn: ({ signal }) => getDriverBalance(token, signal),
    refetchInterval: 5_000,
    meta: { private: true },
  });
  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 15_000);
    return () => window.clearInterval(timer);
  }, []);
  const completion = useMutation({
    mutationFn: (id: string) => completeTrip(id, token),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["trips"] }),
        queryClient.invalidateQueries({ queryKey: ["bookings"] }),
        queryClient.invalidateQueries({ queryKey: driverBalanceKey(driverId) }),
      ]);
      setConfirming(undefined);
    },
  });
  const completeError = isAxiosError(completion.error) && completion.error.response?.status === 409
    ? "Поездку можно завершить после подтверждения прибытия всеми пассажирами."
    : "Не удалось завершить поездку. Попробуйте ещё раз.";
  const money = new Intl.NumberFormat("ru-RU", { style: "currency", currency: balance.data?.currency ?? "USD" });

  return (
    <>
      <div className={styles.heading}>
        <h1>Мои поездки</h1>
        <button type="button" onClick={() => setCreating(true)}>Создать поездку</button>
      </div>
      <section className={styles.notice} aria-label="Тестовый баланс">
        {balance.data ? (
          <>Тестовый баланс · Доступно {money.format(balance.data.available_cents / 100)} · Заблокировано {money.format(balance.data.locked_cents / 100)}</>
        ) : balance.isError ? <p role="alert">Не удалось загрузить баланс.</p> : <p role="status">Загружаем баланс…</p>}
      </section>
      <DriverBookings driverId={driverId} token={token} trips={trips.data ?? []} />
      <section className={styles.panel} aria-label="Список поездок">
        {trips.isPending ? <p role="status">Загружаем поездки…</p>
          : trips.isError && !trips.data ? (
            <div className={styles.failure}>
              <p role="alert">Не удалось загрузить поездки.</p>
              <button type="button" onClick={() => void trips.refetch()} disabled={trips.isFetching}>Повторить</button>
            </div>
          ) : trips.data?.length ? (
            <table>
              <thead><tr>
                <th scope="col">Маршрут</th>
                <th scope="col">Отправление</th>
                <th scope="col">Свободно мест</th>
                <th scope="col">Цена за место</th>
                <th scope="col">Статус и действия</th>
              </tr></thead>
              <tbody>
                {trips.data.map((trip) => (
                  <tr key={trip.id}>
                    <th scope="row">{trip.from_city} → {trip.to_city}</th>
                    <td>{tripDateFormat.format(new Date(trip.departure_at))}</td>
                    <td>{trip.seats_available} / {trip.seats_total}</td>
                    <td>{tripPriceFormat.format(trip.price_per_seat)}</td>
                    <td>
                      {trip.completed_at ? "Завершена"
                        : confirming === trip.id ? (
                          <div className={styles.tripActions}>
                            <span>Подтвердить завершение?</span>
                            <button className={styles.action} type="button" disabled={completion.isPending}
                              onClick={() => completion.mutate(trip.id)}>
                              {completion.isPending ? "Завершаем…" : "Да, завершить"}
                            </button>
                            <button className={styles.action} type="button" disabled={completion.isPending}
                              onClick={() => { setConfirming(undefined); completion.reset(); }}>Назад</button>
                            {completion.isError && <span role="alert">{completeError}</span>}
                          </div>
                        ) : balance.data?.ready_trip_ids.includes(trip.id) ? (
                          <button className={styles.action} type="button" disabled={completion.isPending}
                            onClick={() => { completion.reset(); setConfirming(trip.id); }}>Завершить поездку</button>
                        ) : Date.parse(trip.departure_at) > now ? "Запланирована" : "Ожидает подтверждения пассажиров"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : <p>У вас пока нет поездок.</p>}
      </section>
      {creating && <CreateTripDialog driverId={driverId} token={token} onDismiss={() => setCreating(false)} />}
    </>
  );
}

export function DashboardView() {
  const session = useSession((state) => state.session);
  if (!session) return null;
  if (session.role === "passenger") return <PassengerDashboard userId={session.userId} token={session.token} />;
  return <DriverTrips driverId={session.userId} token={session.token} />;
}
