const headers = { "Cache-Control": "no-store" };

function failure(status: number) {
  return Response.json({ error: "registration_failed" }, { status, headers });
}

export async function POST(request: Request) {
  const body: unknown = await request.json().catch(() => null);
  if (
    !body || typeof body !== "object" ||
    !("name" in body) || typeof body.name !== "string" ||
    !("email" in body) || typeof body.email !== "string" ||
    !("password" in body) || typeof body.password !== "string" ||
    !("role" in body) || (body.role !== "passenger" && body.role !== "driver")
  ) {
    return failure(400);
  }

  try {
    const url = new URL("/api/v1/users/", process.env.API_BASE_URL || "http://localhost:8080");
    const upstream = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: body.name, email: body.email, password: body.password, role: body.role }),
      cache: "no-store",
      redirect: "error",
      signal: AbortSignal.timeout(8_000),
    });

    if (upstream.status === 201) {
      return Response.json({ created: true }, { status: 201, headers });
    }

    return failure(upstream.status === 400 || upstream.status === 409 ? upstream.status : 502);
  } catch {
    return failure(502);
  }
}
