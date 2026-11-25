import { isAxiosError } from "axios";
import { api } from "../../shared/api";

type Booking = {
  id: string;
  trip_id: string;
  seats: number;
  status: string;
  paid: boolean;
  driver_rating: number | null;
  driver_review: string | null;
  payment_available: boolean;
  arrived_at: string | null;
  arrival_available: boolean;

};

const statusLabels = new Map([
  ["requested", "Ожидает решения водителя"],
  ["rejected", "Водитель отклонил заявку"],
  ["pending", "Ожидает оплаты"],
  ["confirmed", "Подтверждена"],
  ["cancelled", "Отменена"],
  ["expired", "Срок брони истёк"],
  ["completed", "Поездка завершена"],
]);

export const bookingStatusLabel = (status: string) => statusLabels.get(status) ?? status;
export const bookingsKey = (userId: string) => ["bookings", userId] as const;

function auth(token: string) {
  return { Authorization: `Bearer ${token}` };
}

export async function listBookings(token: string, signal: AbortSignal) {
  const { data } = await api.get<Booking[]>("/api/bookings", { headers: auth(token), signal });
  return data;
}

export async function bookTrip(tripId: string, seats: number, token: string) {
  await api.post("/api/bookings", { trip_id: tripId, seats }, { headers: auth(token) });
}

export async function cancelBooking(bookingId: string, token: string) {
  await api.delete(`/api/bookings/${bookingId}`, { headers: auth(token) });
}

export async function startBookingPayment(bookingId: string, token: string) {
  const { data } = await api.post<{ url: string }>(`/api/bookings/${bookingId}/pay`, undefined, { headers: auth(token) });
  const checkout = new URL(data.url);
  if (checkout.origin !== "https://checkout.stripe.com" || checkout.username || checkout.password) {
    throw new Error("Invalid Stripe Checkout URL");
  }
  return checkout.href;
}

export function bookingErrorStatus(error: unknown) {
  return isAxiosError(error) ? error.response?.status : undefined;
}

export async function confirmArrival(bookingId: string, token: string) {
  await api.post(`/api/bookings/${bookingId}/arrive`, undefined, { headers: auth(token) });
}

type DriverBooking = {
  seats: number;
  id: string;
  trip_id: string;
  passenger_name: string;
  passenger_rating: number | null;
  passenger_rating_count: number;
  rating_given: number | null;
  review_given: string | null;
  status: string;
  decision_available: boolean;
};

export const driverBookingsKey = (driverId: string) => ["driver-bookings", driverId] as const;

export async function listDriverBookings(token: string, signal: AbortSignal) {
  const { data } = await api.get<DriverBooking[]>("/api/bookings/driver", { headers: auth(token), signal });
  return data;
}

export async function decideBooking(id: string, decision: "approve" | "reject", token: string) {
  await api.post(`/api/bookings/${id}/decision`, { decision }, { headers: auth(token) });
}

export async function rateBooking(id: string, score: number, review: string, token: string) {
  await api.post(`/api/bookings/${id}/rating`, { score, review }, { headers: auth(token) });
}

const ratingFormat = new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 1 });
export function formatRating(score: number | null, count: number) {
  return count && score !== null ? `${ratingFormat.format(score)} из 5 (${count})` : "Нет оценок";
}
