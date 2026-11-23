import { forwardBooking } from "./upstream";

export function GET(request: Request) {
  return forwardBooking(request, "/api/v1/bookings/my");
}

export function POST(request: Request) {
  return forwardBooking(request, "/api/v1/bookings");
}
