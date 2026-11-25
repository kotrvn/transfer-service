"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { type FormEvent, useState } from "react";
import { rateBooking } from "../../entities/booking";
import styles from "./rate-booking.module.scss";

const scores = [1, 2, 3, 4, 5];

export function RateBooking({ id, token, rating, review, subject, onRated }: {
  id: string;
  token: string;
  rating: number | null;
  review: string | null;
  subject: string;
  onRated: () => Promise<void>;
}) {
  const [score, setScore] = useState(0);
  const queryClient = useQueryClient();
  const submit = useMutation({
    mutationFn: (text: string) => rateBooking(id, score, text, token),
    onSettled: async () => {
      await Promise.all([onRated(), queryClient.invalidateQueries({ queryKey: ["trips"] })]);
    },
  });

  if (rating != null) return (
    <div className={styles.saved}>
      <p>Ваша оценка: {rating} из 5</p>
      {review && <p className={styles.review}>{review}</p>}
    </div>
  );

  function send(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (score && !submit.isPending) submit.mutate(String(new FormData(event.currentTarget).get("review") ?? "").trim());
  }

  return (
    <form className={styles.form} onSubmit={send}>
      <fieldset className={styles.fieldset} disabled={submit.isPending}>
        <legend>Оценка {subject}</legend>
        <div className={styles.stars}>
          {scores.map((value) => (
            <label key={value} className={styles.star} data-filled={value <= score}>
              <input type="radio" name="score" value={value} aria-label={`${value} из 5`}
                checked={score === value} onChange={() => setScore(value)} />
              <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
                <path d="m12 2 3.1 6.3 6.9 1-5 4.9 1.2 6.9-6.2-3.3-6.2 3.3 1.2-6.9-5-4.9 6.9-1L12 2Z" />
              </svg>
            </label>
          ))}
        </div>
        <label className={styles.field}>Отзыв (необязательно)
          <textarea name="review" rows={3} maxLength={1000} />
        </label>
      </fieldset>
      <button type="submit" className={styles.submit} disabled={!score || submit.isPending}>
        {submit.isPending ? "Сохраняем…" : "Отправить оценку"}
      </button>
      {submit.isError && <p role="alert">Не удалось сохранить оценку.</p>}
    </form>
  );
}
