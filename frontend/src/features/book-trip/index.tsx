"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type FormEvent, useEffect, useRef, useState } from "react";

import { bookingErrorStatus, bookingsKey, bookTrip } from "../../entities/booking";
import { type Trip, tripPriceFormat } from "../../entities/trip";
import styles from "../../shared/ui/trip-dialog/trip-dialog.module.scss";

export function BookTripDialog({ trip, token, userId, onClose, onBooked }: {
  trip: Trip;
  token: string;
  userId: string;
  onClose: () => void;
  onBooked: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const submitting = useRef(false);
  const [quantity, setQuantity] = useState("1");
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: (seats: number) => bookTrip(trip.id, seats, token),
    retry: false,
    networkMode: "always",
  });

  useEffect(() => {
    const element = dialog.current;
    element?.showModal();
    return () => element?.close();
  }, []);

  function close() {
    dialog.current?.close();
    onClose();
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting.current) return;

    submitting.current = true;
    try {
      await mutation.mutateAsync(Number(quantity));
      void queryClient.invalidateQueries({ queryKey: ["trips"] });
      void queryClient.invalidateQueries({ queryKey: bookingsKey(userId) });
      close();
      onBooked();
    } catch (error) {
      if ([400, 404, 409].includes(bookingErrorStatus(error) ?? 0)) {
        void queryClient.invalidateQueries({ queryKey: ["trips"] });
      }
    } finally {
      submitting.current = false;
    }
  }

  const status = bookingErrorStatus(mutation.error);
  const error = status === 400 ? "Не удалось забронировать столько мест. Проверьте остаток и попробуйте снова."
    : status === 404 || status === 409 ? "Эта поездка больше недоступна."
    : status === 401 || status === 403 ? "Войдите в аккаунт пассажира, чтобы забронировать поездку."
    : "Не удалось создать бронь. Попробуйте ещё раз.";
  const seats = Number(quantity);

  return (
    <dialog
      ref={dialog}
      className={styles.dialog}
      aria-labelledby="book-trip-title"
      onCancel={(event) => {
        event.preventDefault();
        if (!submitting.current) close();
      }}
    >
      <h2 id="book-trip-title">Забронировать поездку</h2>
      <p>{trip.from_city} → {trip.to_city}</p>
      <form className={styles.form} method="post" onSubmit={submit}>
        <fieldset className={styles.fields} disabled={mutation.isPending}>
          <label className={styles.full}>
            Количество мест
            <input type="number" min="1" max={trip.seats_available} step="1" required value={quantity}
              onChange={(event) => setQuantity(event.target.value)} />
          </label>
        </fieldset>
        <div className={styles.summary}>
          <span>Свободно мест: {trip.seats_available}</span>
          <strong>Итого: {Number.isInteger(seats) && seats > 0 ? tripPriceFormat.format(seats * trip.price_per_seat) : "—"}</strong>
        </div>
        {mutation.isError && <p role="alert">{error}</p>}
        <div className={styles.actions}>
          <button type="button" disabled={mutation.isPending} onClick={close}>Закрыть</button>
          <button type="submit" disabled={mutation.isPending || trip.seats_available <= 0}>
            {mutation.isPending ? "Бронируем…" : "Подтвердить бронь"}
          </button>
        </div>
      </form>
    </dialog>
  );
}
