import { forwardTrip } from "../upstream";

export function GET(request: Request) {
  return forwardTrip("GET", request, "/api/v1/trips/balance");
}
