import { BodyTooLarge, readBody } from "../request-body";

const headers = { "Cache-Control": "no-store" };

export function tripFailure(status: number) {
  return Response.json({ error: "trips_request_failed" }, { status, headers });
}

export async function forwardTrip(method: "GET" | "POST", request: Request, path = "/api/v1/trips") {
  const authorization = request.headers.get("authorization");
  const requiresAuth = method === "POST" || path === "/api/v1/trips/balance";
  if (requiresAuth && !authorization?.startsWith("Bearer ")) return tripFailure(401);

  try {
    const url = new URL(path, process.env.API_BASE_URL || "http://localhost:8080");
    if (method === "GET") {
      const search = new URL(request.url).searchParams;
      for (const field of ["from", "to"]) {
        const value = search.get(field);
        if (value) url.searchParams.set(field, value);
      }
    }
    const upstream = await fetch(url, {
      method,
      headers: requiresAuth
        ? { Authorization: authorization ?? "", "Content-Type": "application/json" }
        : undefined,
      body: method === "POST" ? await readBody(request) : undefined,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(8_000),
    });

    if (!upstream.ok) {
      return tripFailure([400, 401, 403, 404, 409].includes(upstream.status) ? upstream.status : 502);
    }
    if (upstream.status === 204) return new Response(null, { status: 204, headers });
    return Response.json(await upstream.json(), { status: upstream.status, headers });
  } catch (error) {
    return tripFailure(error instanceof BodyTooLarge ? 413 : 502);
  }
}
