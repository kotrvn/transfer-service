"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { type FormEvent, useEffect, useRef } from "react";

import { useSession } from "../../entities/session";
import { createTrip, driverTripsKey } from "../../entities/trip";
import styles from "./create-trip.module.scss";

export function CreateTripDialog({ driverId, token, onDismiss }: {
  driverId: string;
  token: string;
  onDismiss: () => void;
}) {
  const dialog = useRef<HTMLDialogElement>(null);
  const submitting = useRef(false);
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: (draft: Parameters<typeof createTrip>[0]) => createTrip(draft, token),
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
    onDismiss();
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting.current) return;

    const form = event.currentTarget;
    const fields = new FormData(form);
    const fromCity = String(fields.get("from_city") ?? "").trim();
    const toCity = String(fields.get("to_city") ?? "").trim();
    const timeInput = form.elements.namedItem("departure_at") as HTMLInputElement;
    const departure = new Date(timeInput.value);
    timeInput.setCustomValidity(Number.isNaN(departure.getTime()) ? "Проверьте дату и время" : "");
    if (!timeInput.reportValidity()) return;

    for (const [name, value] of [["from_city", fromCity], ["to_city", toCity]]) {
      const input = form.elements.namedItem(name) as HTMLInputElement;
      input.setCustomValidity(value ? "" : "Введите название города");
      if (!input.reportValidity()) return;
    }

    submitting.current = true;
    try {
      const trip = await mutation.mutateAsync({
        from_city: fromCity,
        to_city: toCity,
        departure_at: departure.toISOString(),
        seats_total: Number(fields.get("seats_total")),
        price_per_seat: Number(fields.get("price_per_seat")),
      });
      if (useSession.getState().session?.token !== token) return;
      const key = driverTripsKey(driverId);
      queryClient.setQueryData(key, (current: typeof trip[] | undefined) =>
        [...(current ?? []).filter((item) => item.id !== trip.id), trip]
          .sort((first, second) => Date.parse(first.departure_at) - Date.parse(second.departure_at)),
      );
      void queryClient.invalidateQueries({ queryKey: key });
      close();
    } catch {
      // The mutation exposes the error for the form to display.
    } finally {
      submitting.current = false;
    }
  }

  const status = isAxiosError(mutation.error) ? mutation.error.response?.status : undefined;
  const errorMessage = status === 400 ? "Проверьте данные поездки."
    : status === 401 || status === 403 ? "Создавать поездки может только авторизованный водитель."
    : "Не удалось создать поездку. Попробуйте ещё раз.";

  return (
    <dialog
      className={styles.dialog}
      ref={dialog}
      aria-labelledby="trip-dialog-title"
      onCancel={(event) => {
        event.preventDefault();
        if (!submitting.current) close();
      }}
    >
      <h2 id="trip-dialog-title">Создать поездку</h2>
      <form method="post" onSubmit={submit} className={styles.form}>
        <fieldset disabled={mutation.isPending} className={styles.fields}>
          <label>
            Откуда
            <input name="from_city" required maxLength={100} onInput={(e) => e.currentTarget.setCustomValidity("")} />
          </label>
          <label>
            Куда
            <input name="to_city" required maxLength={100} onInput={(e) => e.currentTarget.setCustomValidity("")} />
          </label>
          <label className={styles.full}>
            Дата и время отправления
            <input name="departure_at" type="datetime-local" required onInput={(e) => e.currentTarget.setCustomValidity("")} />
          </label>
          <label>
            Количество мест
            <input name="seats_total" type="number" min="1" max="2147483647" step="1" required />
          </label>
          <label>
            Цена за место
            <input name="price_per_seat" type="number" min="0.01" max="99999999.99" step="0.01" required />
          </label>
        </fieldset>
        {mutation.isError && <p role="alert">{errorMessage}</p>}
        <div className={styles.actions}>
          <button type="button" disabled={mutation.isPending} onClick={close}>Отмена</button>
          <button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? "Создаём…" : "Создать поездку"}
          </button>
        </div>
      </form>
    </dialog>
  );
}
