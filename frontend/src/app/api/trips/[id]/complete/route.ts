import { forwardTrip, tripFailure } from "../../upstream";

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!/^[\da-f]{8}(?:-[\da-f]{4}){3}-[\da-f]{12}$/i.test(id)) return tripFailure(400);
  return forwardTrip("POST", request, `/api/v1/trips/${id}/complete`);
}
