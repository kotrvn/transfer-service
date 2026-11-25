"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { type FormEvent, useState } from "react";

import { formatRating } from "../../entities/booking";
import { listTrips, type Trip, tripDateFormat, tripPriceFormat, tripSearchKey } from "../../entities/trip";
import { BookTripDialog } from "../../features/book-trip";
import styles from "./dashboard.module.scss";

export function PassengerDashboard({ token, userId }: { token: string; userId: string }) {
  const [{ from, to }, setSearch] = useState({ from: "", to: "" });
  const [selection, setSelection] = useState<Trip>();
  const [booked, setBooked] = useState(false);
  const trips = useQuery({
    queryKey: tripSearchKey(from, to),
    queryFn: ({ signal }) => listTrips(from, to, signal),
    retry: 1,
  });
  const availableTrips = trips.data?.filter((trip) => !trip.completed_at);
  const selectedTrip = availableTrips?.find((trip) => trip.id === selection?.id);

  function search(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const filters = { from: String(data.get("from") ?? "").trim(), to: String(data.get("to") ?? "").trim() };
    if (filters.from === from && filters.to === to) void trips.refetch();
    else setSearch(filters);
  }

  return (
    <>
      <div className={styles.heading}><h1>Поиск поездок</h1></div>
      <form className={styles.filters} onSubmit={search} role="search" aria-label="Поиск поездок">
        <label>Откуда<input name="from" maxLength={100} /></label>
        <label>Куда<input name="to" maxLength={100} /></label>
        <button type="submit" className={styles.primary}>Найти</button>
      </form>
      {booked && <div role="status" className={styles.notice}>Заявка отправлена водителю. <Link href="/bookings">Мои бронирования</Link></div>}
      <section className={styles.panel} aria-label="Результаты поиска">
        {trips.isPending ? <p role="status">Загружаем поездки…</p>
          : trips.isError && !trips.data ? (
            <div className={styles.failure}>
              <p role="alert">Не удалось загрузить поездки.</p>
              <button type="button" onClick={() => void trips.refetch()} disabled={trips.isFetching}>Повторить</button>
            </div>
          ) : availableTrips?.length ? (
            <table>
              <thead><tr>
                <th scope="col">Маршрут</th><th scope="col">Отправление</th>
                <th scope="col">Свободно мест</th><th scope="col">Цена за место</th>
                <th scope="col">Водитель</th><th scope="col">Бронирование</th>
              </tr></thead>
              <tbody>
                {availableTrips.map((trip) => (
                  <tr key={trip.id}>
                    <th scope="row">{trip.from_city} → {trip.to_city}</th>
                    <td>{tripDateFormat.format(new Date(trip.departure_at))}</td>
                    <td>{trip.seats_available}</td>
                    <td>{tripPriceFormat.format(trip.price_per_seat)}</td>
                    <td>{trip.driver_name} · {formatRating(trip.driver_rating, trip.driver_rating_count)}</td>
                    <td>
                      {trip.seats_available > 0 ? (
                        <button type="button" className={styles.action} onClick={() => { setBooked(false); setSelection(trip); }}>
                          Забронировать
                        </button>
                      ) : "Нет мест"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          ) : <p>По этому запросу поездок нет.</p>}
      </section>
      {selectedTrip && (
        <BookTripDialog key={selectedTrip.id} trip={selectedTrip} userId={userId} token={token}
          onClose={() => setSelection(undefined)} onBooked={() => setBooked(true)} />
      )}
    </>
  );
}
