import { isAxiosError } from "axios";
import { api } from "../../shared/api";

type Booking = {
  id: string;
  trip_id: string;
  seats: number;
  status: string;
};

const statusLabels = new Map([
  ["pending", "Ожидает подтверждения"],
  ["confirmed", "Подтверждена"],
  ["cancelled", "Отменена"],
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

export function bookingErrorStatus(error: unknown) {
  return isAxiosError(error) ? error.response?.status : undefined;
}
