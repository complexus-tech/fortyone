/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError, get } from "api-client";
import {
  AuthContractError,
  AuthSessionLookupError,
  CurrentUserLookupError,
  getCurrentUser,
  getSessionFromRequest,
} from "auth";

jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {
    constructor(
      message: string,
      readonly status: number,
      readonly data: unknown,
    ) {
      super(message);
      this.name = "ApiError";
    }
  },
  get: jest.fn(),
}));

const mockedGet = jest.mocked(get);
const originalApiUrl = process.env.NEXT_PUBLIC_API_URL;
const originalFetch = globalThis.fetch;

const currentUser = {
  avatarUrl: null,
  createdAt: "2026-08-30T12:00:00Z",
  email: "joseph@example.com",
  fullName: "Joseph",
  hasSeenWalkthrough: true,
  id: "user-1",
  isActive: true,
  isInternal: false,
  lastUsedWorkspaceId: "workspace-1",
  timezone: "Africa/Harare",
  updatedAt: "2026-08-30T12:00:00Z",
  username: "joseph",
};

beforeEach(() => {
  jest.clearAllMocks();
  process.env.NEXT_PUBLIC_API_URL = "https://api.fortyone.test";
});

afterEach(() => {
  globalThis.fetch = originalFetch;
  if (originalApiUrl === undefined) {
    delete process.env.NEXT_PUBLIC_API_URL;
  } else {
    process.env.NEXT_PUBLIC_API_URL = originalApiUrl;
  }
});

describe("current-user authentication contract", () => {
  it("returns null only for an explicit unauthorized response", async () => {
    mockedGet.mockRejectedValue(new ApiError("Unauthorized", 401, null));

    await expect(getCurrentUser()).resolves.toBeNull();
  });

  it("surfaces upstream failures without retaining their response data", async () => {
    const upstreamError = new ApiError("Unavailable", 503, {
      providerToken: "secret-value",
    });
    mockedGet.mockRejectedValue(upstreamError);

    try {
      await getCurrentUser();
      throw new Error("Expected current-user lookup to fail");
    } catch (error) {
      expect(error).toEqual(
        expect.objectContaining({
          message: "Current-user lookup failed with status 503",
          name: "CurrentUserLookupError",
          status: 503,
        }),
      );
      expect(error).toBeInstanceOf(CurrentUserLookupError);
      expect((error as Error).cause).toBeUndefined();
      expect(String(error)).not.toContain("secret-value");
    }
  });

  it("rejects malformed current-user data", async () => {
    mockedGet.mockResolvedValue({ data: { id: "user-1" } });

    await expect(getCurrentUser()).rejects.toBeInstanceOf(AuthContractError);
  });

  it("normalizes a user who has not selected a workspace yet", async () => {
    mockedGet.mockResolvedValue({
      data: { ...currentUser, lastUsedWorkspaceId: null },
    });

    await expect(getCurrentUser()).resolves.toMatchObject({
      lastUsedWorkspaceId: "",
    });
  });

  it("does not interpret a successful null user as an authenticated session", async () => {
    mockedGet.mockResolvedValue({ data: null });

    await expect(getCurrentUser()).rejects.toBeInstanceOf(AuthContractError);
  });
});

describe("request session lookup", () => {
  const request = (cookie = "fortyone_session=session") =>
    ({
      headers: {
        get: (name: string) =>
          name.toLowerCase() === "cookie" ? cookie : null,
      },
    }) as Request;

  const response = (status: number, body: unknown): Response =>
    ({
      json: async () => body,
      ok: status >= 200 && status < 300,
      status,
    }) as Response;

  it("returns the validated user", async () => {
    const mockedFetch = jest.fn(async () =>
      response(200, { data: currentUser }),
    );
    globalThis.fetch = mockedFetch as typeof fetch;

    await expect(getSessionFromRequest(request())).resolves.toEqual(
      currentUser,
    );
    expect(mockedFetch).toHaveBeenCalledWith(
      "https://api.fortyone.test/auth/me",
      {
        cache: "no-store",
        headers: { cookie: "fortyone_session=session" },
        method: "GET",
      },
    );
  });

  it("forwards only the session cookie to the auth upstream", async () => {
    const mockedFetch = jest.fn(async () =>
      response(200, { data: currentUser }),
    );
    globalThis.fetch = mockedFetch as typeof fetch;

    await getSessionFromRequest(
      request(
        "theme=dark; fortyone_session=session%3Dvalue; posthog=anonymous",
      ),
    );

    expect(mockedFetch).toHaveBeenCalledWith(
      "https://api.fortyone.test/auth/me",
      expect.objectContaining({
        headers: { cookie: "fortyone_session=session%3Dvalue" },
      }),
    );
  });

  it.each([
    ["no cookies", ""],
    ["an unrelated cookie", "theme=dark; posthog=anonymous"],
    ["an empty session cookie", "fortyone_session="],
    ["a similarly named cookie", "not_fortyone_session=session"],
  ])("does not contact the auth upstream with %s", async (_case, cookie) => {
    const mockedFetch = jest.fn();
    globalThis.fetch = mockedFetch as typeof fetch;

    await expect(getSessionFromRequest(request(cookie))).resolves.toBeNull();
    expect(mockedFetch).not.toHaveBeenCalled();
  });

  it("returns null for 401 and surfaces 5xx responses", async () => {
    globalThis.fetch = jest.fn(async () =>
      response(401, { data: null }),
    ) as typeof fetch;
    await expect(getSessionFromRequest(request())).resolves.toBeNull();

    globalThis.fetch = jest.fn(async () =>
      response(503, { error: { message: "unavailable" } }),
    ) as typeof fetch;
    await expect(getSessionFromRequest(request())).rejects.toMatchObject({
      name: "AuthSessionLookupError",
      status: 503,
    });
  });

  it("wraps network failures as typed lookup errors", async () => {
    const networkError = new TypeError("fetch failed");
    globalThis.fetch = jest.fn(async () => {
      throw networkError;
    }) as typeof fetch;

    await expect(getSessionFromRequest(request())).rejects.toMatchObject({
      cause: networkError,
      message: "Current-user lookup failed",
      name: "AuthSessionLookupError",
      status: undefined,
    });
  });

  it("rejects a successful response that does not contain a user", async () => {
    globalThis.fetch = jest.fn(async () =>
      response(200, { data: null }),
    ) as typeof fetch;

    await expect(getSessionFromRequest(request())).rejects.toMatchObject({
      message: "Current-user response did not match its contract",
      name: "AuthSessionLookupError",
      status: 200,
    });
  });

  it("wraps malformed responses without retaining their body-derived cause", async () => {
    globalThis.fetch = jest.fn(
      async () =>
        ({
          json: async () => {
            throw new SyntaxError("provider-secret");
          },
          ok: true,
          status: 200,
        }) as unknown as Response,
    ) as typeof fetch;

    try {
      await getSessionFromRequest(request());
      throw new Error("Expected current-user lookup to fail");
    } catch (error) {
      expect(error).toEqual(
        expect.objectContaining({
          message: "Current-user response was not valid JSON",
          name: "AuthSessionLookupError",
          status: 200,
        }),
      );
      expect((error as Error).cause).toBeUndefined();
      expect(String(error)).not.toContain("provider-secret");
    }
  });

  it("fails closed when the API origin is not configured", async () => {
    delete process.env.NEXT_PUBLIC_API_URL;

    await expect(getSessionFromRequest(request())).rejects.toBeInstanceOf(
      AuthSessionLookupError,
    );
  });
});
