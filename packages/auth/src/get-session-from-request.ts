import { decodeCurrentUserResponse } from "./current-user-contract";
import { getSessionCookieHeader } from "./session-cookie";
import type { CurrentUser } from "./types";

export class AuthSessionLookupError extends Error {
  readonly status?: number;

  constructor(message: string, status?: number, options?: ErrorOptions) {
    super(message, options);
    this.name = "AuthSessionLookupError";
    this.status = status;
  }
}

export const getSessionFromRequest = async (
  request: Request,
): Promise<CurrentUser | null> => {
  const sessionCookieHeader = getSessionCookieHeader(
    request.headers.get("cookie"),
  );

  if (!sessionCookieHeader) return null;

  const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL;
  if (!apiBaseUrl) {
    throw new AuthSessionLookupError("NEXT_PUBLIC_API_URL is not configured");
  }

  try {
    const response = await fetch(`${apiBaseUrl}/auth/me`, {
      method: "GET",
      headers: {
        cookie: sessionCookieHeader,
      },
      cache: "no-store",
    });

    if (response.status === 401) return null;
    if (!response.ok) {
      throw new AuthSessionLookupError(
        `Current-user lookup failed with status ${response.status}`,
        response.status,
      );
    }

    let body: unknown;
    try {
      body = (await response.json()) as unknown;
    } catch {
      throw new AuthSessionLookupError(
        "Current-user response was not valid JSON",
        response.status,
      );
    }

    try {
      const user = decodeCurrentUserResponse(body).data;
      if (!user) {
        throw new AuthSessionLookupError(
          "Authenticated current-user response did not include a user",
          response.status,
        );
      }
      return user;
    } catch {
      throw new AuthSessionLookupError(
        "Current-user response did not match its contract",
        response.status,
      );
    }
  } catch (cause) {
    if (cause instanceof AuthSessionLookupError) throw cause;

    throw new AuthSessionLookupError("Current-user lookup failed", undefined, {
      cause,
    });
  }
};
