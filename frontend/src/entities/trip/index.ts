import { api } from "../../shared/api";

export type TripDraft = {
  from_city: string;
  to_city: string;
  departure_at: string;
  seats_total: number;
  price_per_seat: number;
};

type Trip = TripDraft & {
  id: string;
  driver_id: string;
  seats_available: number;
};

export const driverTripsKey = (driverId: string) => ["trips", "driver", driverId] as const;

export async function listDriverTrips(userId: string, signal: AbortSignal) {
  const { data } = await api.get<Trip[]>("/api/trips", { signal });
  return data
    .filter((trip) => trip.driver_id === userId)
    .sort((first, second) => Date.parse(first.departure_at) - Date.parse(second.departure_at));
}

export async function createTrip(draft: TripDraft, token: string) {
  const { data } = await api.post<Trip>("/api/trips", draft, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return data;
}
