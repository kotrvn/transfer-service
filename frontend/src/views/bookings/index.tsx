"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";

import { bookingErrorStatus, bookingsKey, bookingStatusLabel, cancelBooking, listBookings } from "../../entities/booking";
import { useSession } from "../../entities/session";
import { listTrips, tripDateFormat, tripSearchKey } from "../../entities/trip";
import styles from "./bookings.module.scss";

function PassengerBookings({ token, userId }: { token: string; userId: string }) {
  const queryClient = useQueryClient();
  const [confirming, setConfirming] = useState<string>();
  const bookings = useQuery({
    queryKey: bookingsKey(userId),
    queryFn: ({ signal }) => listBookings(token, signal),
    retry: 1,
    meta: { private: true },
  });
  const trips = useQuery({
    queryKey: tripSearchKey("", ""),
    queryFn: ({ signal }) => listTrips("", "", signal),
    retry: 1,
    enabled: !!bookings.data?.length,
  });
  const tripsById = new Map(trips.data?.map((trip) => [trip.id, trip]));
  const cancellation = useMutation({
    mutationFn: (id: string) => cancelBooking(id, token),
    retry: false,
    networkMode: "always",
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: bookingsKey(userId) }),
        queryClient.invalidateQueries({ queryKey: ["trips"] }),
      ]);
      setConfirming(undefined);
    },
    onError: (error) => {
      if (bookingErrorStatus(error) === 400) {
        void queryClient.invalidateQueries({ queryKey: bookingsKey(userId) });
      }
    },
  });

  return (
    <>
      <h1 className={styles.title}>Мои бронирования</h1>
      <section className={styles.panel} aria-label="Список бронирований">
        {bookings.isPending ? <p role="status">Загружаем бронирования…</p>
          : bookings.isError && !bookings.data ? (
            <div>
              <p role="alert">Не удалось загрузить бронирования.</p>
              <button type="button" onClick={() => void bookings.refetch()} disabled={bookings.isFetching}>Повторить</button>
            </div>
          ) : bookings.data?.length ? (
            <ul className={styles.list}>
              {bookings.data.map((booking) => {
                const trip = tripsById.get(booking.trip_id);
                const active = booking.status === "pending" || booking.status === "confirmed";
                return (
                  <li key={booking.id} className={styles.booking}>
                    <div>
                      <h2>{trip ? `${trip.from_city} → ${trip.to_city}` : "Поездка"}</h2>
                      {trip && <time dateTime={trip.departure_at}>{tripDateFormat.format(new Date(trip.departure_at))}</time>}
                      <p>Мест: {booking.seats} · {bookingStatusLabel(booking.status)}</p>
                    </div>
                    <div className={styles.actions}>
                      {active && (confirming === booking.id ? (
                        <>
                          <span>Отменить бронь?</span>
                          <button type="button" disabled={cancellation.isPending}
                            onClick={() => cancellation.mutate(booking.id)}>
                            {cancellation.isPending ? "Отменяем…" : "Да, отменить"}
                          </button>
                          <button type="button" disabled={cancellation.isPending} onClick={() => { setConfirming(undefined); cancellation.reset(); }}>Оставить</button>
                        </>
                      ) : (
                        <button type="button" disabled={cancellation.isPending}
                          onClick={() => { cancellation.reset(); setConfirming(booking.id); }}>Отменить бронь</button>
                      ))}
                      {cancellation.isError && cancellation.variables === booking.id && (
                        <p className={styles.error} role="alert">Не удалось отменить бронь. Попробуйте ещё раз.</p>
                      )}
                    </div>
                  </li>
                );
              })}
            </ul>
          ) : <p>У вас пока нет бронирований. <Link href="/dashboard">Найти поездку</Link></p>}
      </section>
    </>
  );
}

export function BookingsView() {
  const session = useSession((state) => state.session);
  if (!session) return null;
  if (session.role !== "passenger") {
    return <p>Бронирования доступны пассажирам. <Link href="/dashboard">Вернуться в кабинет</Link></p>;
  }
  return <PassengerBookings userId={session.userId} token={session.token} />;
}
