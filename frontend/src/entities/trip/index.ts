import { api } from "../../shared/api";

export type TripDraft = {
  from_city: string;
  to_city: string;
  departure_at: string;
  seats_total: number;
  price_per_seat: number;
};

export type Trip = TripDraft & {
  id: string;
  driver_id: string;
  driver_name: string;
  driver_rating: number | null;
  driver_rating_count: number;
  seats_available: number;
  completed_at: string | null;
};

export const driverTripsKey = (driverId: string) => ["trips", "driver", driverId] as const;
export const tripSearchKey = (from: string, to: string) => ["trips", "search", from, to] as const;
export const tripDateFormat = new Intl.DateTimeFormat("ru-RU", { dateStyle: "medium", timeStyle: "short" });
export const tripPriceFormat = new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 2 });

export async function listTrips(from: string, to: string, signal: AbortSignal) {
  const { data } = await api.get<Trip[]>("/api/trips", { params: { from, to }, signal });
  return data
    .sort((first, second) => Date.parse(first.departure_at) - Date.parse(second.departure_at));
}

export async function listDriverTrips(userId: string, signal: AbortSignal) {
  return (await listTrips("", "", signal)).filter((trip) => trip.driver_id === userId);
}

export async function createTrip(draft: TripDraft, token: string) {
  const { data } = await api.post<Trip>("/api/trips", draft, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return data;
}

export async function completeTrip(tripId: string, token: string) {
  await api.post(`/api/trips/${tripId}/complete`, undefined, {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export const driverBalanceKey = (driverId: string) => ["driver-balance", driverId] as const;

export async function getDriverBalance(token: string, signal: AbortSignal) {
  const { data } = await api.get<{
    currency: string;
    locked_cents: number;
    available_cents: number;
    ready_trip_ids: string[];
  }>("/api/trips/balance", { headers: { Authorization: `Bearer ${token}` }, signal });
  return data;
}
