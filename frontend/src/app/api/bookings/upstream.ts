const headers = { "Cache-Control": "no-store" };

export function bookingFailure(status: number) {
  return Response.json({ error: "booking_request_failed" }, { status, headers });
}

export async function forwardBooking(request: Request, path: string) {
  const authorization = request.headers.get("authorization");
  if (!authorization?.startsWith("Bearer ")) return bookingFailure(401);

  try {
    const upstream = await fetch(new URL(path, process.env.API_BASE_URL || "http://localhost:8080"), {
      method: request.method,
      headers: { Authorization: authorization, "Content-Type": "application/json" },
      body: request.method === "POST" ? await request.text() : undefined,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(8_000),
    });
    if (!upstream.ok) {
      return bookingFailure([400, 401, 403, 404].includes(upstream.status) ? upstream.status : 502);
    }
    if (upstream.status === 204) return new Response(null, { status: 204, headers });
    return Response.json(await upstream.json(), { status: upstream.status, headers });
  } catch {
    return bookingFailure(502);
  }
}
