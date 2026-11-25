import { bookingFailure, forwardBooking } from "../../upstream";

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!/^[\da-f]{8}(?:-[\da-f]{4}){3}-[\da-f]{12}$/i.test(id)) return bookingFailure(400);
  return forwardBooking(request, `/api/v1/bookings/${id}/rating`);
}
