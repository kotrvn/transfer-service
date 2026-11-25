"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { bookingStatusLabel, decideBooking, driverBookingsKey, formatRating, listDriverBookings } from "../../entities/booking";
import type { Trip } from "../../entities/trip";
import { RateBooking } from "../../features/rate-booking";
import styles from "./dashboard.module.scss";

export function DriverBookings({ driverId, token, trips }: { driverId: string; token: string; trips: Trip[] }) {
  const client = useQueryClient();
  const queryKey = driverBookingsKey(driverId);
  const bookings = useQuery({
    queryKey,
    queryFn: ({ signal }) => listDriverBookings(token, signal),
    refetchInterval: 5_000,
    meta: { private: true },
  });
  const decision = useMutation({
    mutationFn: ({ id, action }: { id: string; action: "approve" | "reject" }) => decideBooking(id, action, token),
    onSettled: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey }),
        client.invalidateQueries({ queryKey: ["trips"] }),
      ]);
    },
  });
  const pendingCount = bookings.data?.filter((booking) => booking.decision_available).length ?? 0;
  const sortedBookings = [...(bookings.data ?? [])].sort(
    (first, second) => Number(second.decision_available) - Number(first.decision_available),
  );
  const byId = new Map(trips.map((trip) => [trip.id, trip]));

  return (
    <section className={styles.panel} aria-label="Бронирования пассажиров">
      <h2 className={styles.sectionTitle}>Бронирования пассажиров{pendingCount > 0 && ` · Ожидают решения: ${pendingCount}`}</h2>
      {bookings.data && pendingCount === 0 && <p>Сейчас нет заявок, которые можно принять или отклонить.</p>}
      {bookings.isPending ? <p role="status">Загружаем бронирования…</p>
        : !bookings.data ? <p role="alert">Не удалось загрузить бронирования.</p>
        : !bookings.data.length ? <p>Заявок пока нет.</p> : (
          <table>
            <thead><tr><th scope="col">Пассажир</th><th scope="col">Поездка</th><th scope="col">Мест</th><th scope="col">Статус</th><th scope="col">Действия</th></tr></thead>
            <tbody>
              {sortedBookings.map((booking) => {
                const trip = byId.get(booking.trip_id);
                return (
                  <tr key={booking.id}>
                    <th scope="row">{booking.passenger_name} · {formatRating(booking.passenger_rating, booking.passenger_rating_count)}</th>
                    <td>{trip ? `${trip.from_city} → ${trip.to_city}` : "Поездка"}</td>
                    <td>{booking.seats}</td>
                    <td>{bookingStatusLabel(booking.status)}</td>
                    <td>
                      {booking.decision_available && (
                        <div className={styles.tripActions}>
                          <button className={styles.action} disabled={decision.isPending}
                            onClick={() => decision.mutate({ id: booking.id, action: "approve" })}>Принять</button>
                          <button className={styles.action} disabled={decision.isPending}
                            onClick={() => decision.mutate({ id: booking.id, action: "reject" })}>Отклонить</button>
                        </div>
                      )}
                      {decision.isError && decision.variables.id === booking.id && <span role="alert">Не удалось изменить заявку.</span>}
                      {booking.status === "completed" && (
                        <RateBooking id={booking.id} token={token} rating={booking.rating_given} review={booking.review_given} subject="пассажира"
                          onRated={() => client.invalidateQueries({ queryKey })} />
                      )}
                      {!booking.decision_available && booking.status !== "completed" && "Действия недоступны"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
    </section>
  );
}
