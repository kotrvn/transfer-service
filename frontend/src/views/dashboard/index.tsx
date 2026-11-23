"use client";

import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

import { useSession } from "../../entities/session";
import { driverTripsKey, listDriverTrips } from "../../entities/trip";
import { CreateTripDialog } from "../../features/create-trip";
import styles from "./dashboard.module.scss";

const dateFormat = new Intl.DateTimeFormat("ru-RU", { dateStyle: "medium", timeStyle: "short" });
const priceFormat = new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 2 });

function DriverTrips({ driverId, token }: { driverId: string; token: string }) {
  const [creating, setCreating] = useState(false);
  const trips = useQuery({
    queryKey: driverTripsKey(driverId),
    queryFn: ({ signal }) => listDriverTrips(driverId, signal),
    retry: 1,
    meta: { private: true },
  });

  return (
    <>
      <div className={styles.heading}>
        <h1>Мои поездки</h1>
        <button type="button" onClick={() => setCreating(true)}>Создать поездку</button>
      </div>
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
              </tr></thead>
              <tbody>
                {trips.data.map((trip) => (
                  <tr key={trip.id}>
                    <th scope="row">{trip.from_city} → {trip.to_city}</th>
                    <td>{dateFormat.format(new Date(trip.departure_at))}</td>
                    <td>{trip.seats_available} / {trip.seats_total}</td>
                    <td>{priceFormat.format(trip.price_per_seat)}</td>
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
  if (session.role !== "driver") return <h1 className={styles.title}>Кабинет пассажира</h1>;
  return <DriverTrips driverId={session.userId} token={session.token} />;
}
