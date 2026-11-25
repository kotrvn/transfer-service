"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";

import { bookingErrorStatus, bookingsKey, bookingStatusLabel, cancelBooking, listBookings, startBookingPayment } from "../../entities/booking";
import { useSession } from "../../entities/session";
import { listTrips, tripDateFormat, tripSearchKey } from "../../entities/trip";
import styles from "./bookings.module.scss";

const checkoutPrice = new Intl.NumberFormat("ru-RU", {
  style: "currency",
  currency: "USD",
  currencyDisplay: "code",
});

function PassengerBookings({ token, userId }: { token: string; userId: string }) {
  const queryClient = useQueryClient();
  const [confirming, setConfirming] = useState<string>();
  const bookings = useQuery({
    queryKey: bookingsKey(userId),
    queryFn: ({ signal }) => listBookings(token, signal),
    retry: 1,
        refetchInterval: (query) => {
      const bookings = query.state.data;
      if (bookings?.some((booking) => booking.status === "pending")) return 5_000;
      return bookings?.some((booking) => booking.status === "confirmed") ? 30_000 : false;
    },
    meta: { private: true },
  });
  const trips = useQuery({
    queryKey: tripSearchKey("", ""),
    queryFn: ({ signal }) => listTrips("", "", signal),
    retry: 1,
    enabled: !!bookings.data?.length,
  });
  const tripsById = new Map(trips.data?.map((trip) => [trip.id, trip]));
  const payment = useMutation({
    mutationFn: (id: string) => startBookingPayment(id, token),
    onError: (error) => {
      if (bookingErrorStatus(error) === 409) {
        void queryClient.invalidateQueries({ queryKey: bookingsKey(userId) });
      }
    },
  });
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

  async function openCheckout(id: string) {
    const tab = window.open("about:blank", "_blank");
    if (tab) {
      tab.opener = null;
      tab.document.title = "Открываем Stripe Checkout";
      tab.document.body.textContent = "Открываем Stripe Checkout…";
    }
    try {
      const url = await payment.mutateAsync(id);
      if (tab && !tab.closed) tab.location.replace(url);
    } catch {
      tab?.close();
    }
  }

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
                const active = booking.status === "pending";
                return (
                  <li key={booking.id} className={styles.booking}>
                    <div>
                      <h2>{trip ? `${trip.from_city} → ${trip.to_city}` : "Поездка"}</h2>
                      {trip && <time dateTime={trip.departure_at}>{tripDateFormat.format(new Date(trip.departure_at))}</time>}
                      <p>Мест: {booking.seats} · {bookingStatusLabel(booking.status)}{booking.paid && " · Тестовая оплата принята"}</p>
                      {(booking.payment_available || booking.paid) && trip && (
                        <p>{checkoutPrice.format(trip.price_per_seat * booking.seats)}</p>
                      )}
                    </div>
                    <div className={styles.actions}>
                      {booking.payment_available && confirming !== booking.id && (
                        payment.isSuccess && payment.variables === booking.id ? (
                          <>
                            <a className={styles.pay} href={payment.data} target="_blank" rel="noopener noreferrer">
                              Открыть Stripe Checkout ↗
                            </a>
                            <p className={styles.paymentStatus} role="status">Статус обновится после подтверждения Stripe.</p>
                          </>
                        ) : (
                          <button className={styles.pay} type="button" disabled={payment.isPending || cancellation.isPending}
                            onClick={() => void openCheckout(booking.id)}>
                            {payment.isPending && payment.variables === booking.id ? "Открываем Stripe…" : "Оплатить в тестовом режиме ↗"}
                          </button>
                        )
                      )}
                      {payment.isError && payment.variables === booking.id && (
                        <p className={styles.error} role="alert">
                          {bookingErrorStatus(payment.error) === 409 ? "Бронь больше недоступна для оплаты." : "Не удалось начать оплату. Попробуйте ещё раз."}
                        </p>
                      )}
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
