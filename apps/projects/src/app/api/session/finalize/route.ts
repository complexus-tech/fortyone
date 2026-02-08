import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";

export async function POST() {
  const session = await auth();

  if (!session?.token) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const response = await fetch(`${getApiUrl()}/users/session`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${session.token}`,
      },
      cache: "no-store",
    });

    const nextResponse = NextResponse.json(
      { ok: response.ok },
      { status: response.status },
    );

    const sessionCookie = response.headers.get("set-cookie");
    if (sessionCookie) {
      nextResponse.headers.append("set-cookie", sessionCookie);
    }

    return nextResponse;
  } catch {
    return NextResponse.json(
      { error: "Session finalize failed" },
      { status: 500 },
    );
  }
}
