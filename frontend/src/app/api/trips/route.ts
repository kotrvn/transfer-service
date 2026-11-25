import { forwardTrip } from "./upstream";

export function GET(request: Request) {
  return forwardTrip("GET", request);
}

export function POST(request: Request) {
  return forwardTrip("POST", request);
}
