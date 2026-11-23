const headers = { "Cache-Control": "no-store" };

function failure(status: number) {
  return Response.json({ error: "trips_request_failed" }, { status, headers });
}

async function forward(method: "GET" | "POST", request: Request) {
  const authorization = request.headers.get("authorization");
  if (method === "POST" && !authorization?.startsWith("Bearer ")) return failure(401);

  try {
    const url = new URL("/api/v1/trips", process.env.API_BASE_URL || "http://localhost:8080");
    if (method === "GET") {
      const search = new URL(request.url).searchParams;
      for (const field of ["from", "to"]) {
        const value = search.get(field);
        if (value) url.searchParams.set(field, value);
      }
    }
    const upstream = await fetch(url, {
      method,
      headers: method === "POST"
        ? { Authorization: authorization ?? "", "Content-Type": "application/json" }
        : undefined,
      body: method === "POST" ? await request.text() : undefined,
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(8_000),
    });

    if (!upstream.ok) {
      return failure([400, 401, 403].includes(upstream.status) ? upstream.status : 502);
    }
    return Response.json(await upstream.json(), { status: upstream.status, headers });
  } catch {
    return failure(502);
  }
}

export function GET(request: Request) {
  return forward("GET", request);
}

export function POST(request: Request) {
  return forward("POST", request);
}
