import type { ApiResponse, CurrentUser } from "./types";

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL;

export const getSessionFromRequest = async (
  request: Request,
): Promise<CurrentUser | null> => {
  const cookieHeader = request.headers.get("cookie") ?? "";

  if (!cookieHeader || !apiBaseUrl) {
    return null;
  }

  try {
    const response = await fetch(`${apiBaseUrl}/auth/me`, {
      method: "GET",
      headers: {
        cookie: cookieHeader,
      },
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    const body = (await response.json()) as ApiResponse<CurrentUser>;
    return body.data ?? null;
  } catch {
    return null;
  }
};
