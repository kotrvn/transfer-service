const headers = { "Cache-Control": "no-store" };

function failure(status: number) {
  return Response.json({ error: "login_failed" }, { status, headers });
}

export async function POST(request: Request) {
  const body: unknown = await request.json().catch(() => null);
  if (
    !body || typeof body !== "object" ||
    !("email" in body) || typeof body.email !== "string" ||
    !("password" in body) || typeof body.password !== "string"
  ) {
    return failure(400);
  }

  try {
    const url = new URL("/api/v1/users/login", process.env.API_BASE_URL || "http://localhost:8080");
    const upstream = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: body.email, password: body.password }),
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(8_000),
    });

    if (!upstream.ok) {
      return failure(upstream.status === 400 || upstream.status === 401 ? upstream.status : 502);
    }

    const result: unknown = await upstream.json();
    if (!result || typeof result !== "object" || !("token" in result) || typeof result.token !== "string") {
      return failure(502);
    }
    return Response.json({ token: result.token }, { headers });
  } catch {
    return failure(502);
  }
}
