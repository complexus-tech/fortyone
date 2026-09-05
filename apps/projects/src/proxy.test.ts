/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { NextRequest } from "next/server";
import { AuthSessionLookupError, getSessionFromRequest } from "auth";
import {
  getCanonicalPublicPath,
  getInternalPublicPath,
  isPublicPath,
} from "./public-portal-routes";
import proxy from "./proxy";

jest.mock("auth", () => ({
  AuthSessionLookupError: class AuthSessionLookupError extends Error {
    readonly status?: number;

    constructor(message: string, status?: number) {
      super(message);
      this.name = "AuthSessionLookupError";
      this.status = status;
    }
  },
  getSessionFromRequest: jest.fn(),
}));
jest.mock("next/server", () => {
  class MockNextResponse {
    readonly headers: { get: (name: string) => string | null };
    readonly status: number;
    readonly body: string;

    static next = jest.fn(() => ({ status: 200 }));
    static redirect = jest.fn();
    static rewrite = jest.fn();

    constructor(
      body = "",
      init: { headers?: Record<string, string>; status?: number } = {},
    ) {
      const headers = new Map(
        Object.entries(init.headers ?? {}).map(([name, value]) => [
          name.toLowerCase(),
          value,
        ]),
      );

      this.body = body;
      this.headers = {
        get: (name: string) => headers.get(name.toLowerCase()) ?? null,
      };
      this.status = init.status ?? 200;
    }

    async text() {
      return this.body;
    }
  }

  return { NextResponse: MockNextResponse };
});

const { NextResponse: mockedResponse } = jest.requireMock<{
  NextResponse: { redirect: jest.Mock; rewrite: jest.Mock };
}>("next/server");

const mockedGetSessionFromRequest = jest.mocked(getSessionFromRequest);

beforeEach(() => {
  jest.clearAllMocks();
  mockedGetSessionFromRequest.mockReset();
});

describe("workspace public portal routes", () => {
  it("does not depend on the auth upstream for public routes", async () => {
    const response = await proxy({
      headers: { get: () => null },
      nextUrl: {
        hostname: "fortyone.app",
        pathname: "/feedback",
        search: "",
      },
      url: "https://fortyone.app/feedback",
    } as unknown as NextRequest);

    expect(response.status).toBe(200);
    expect(mockedGetSessionFromRequest).not.toHaveBeenCalled();
  });

  it("rewrites the short feedback URL to the internal portal route", () => {
    expect(getInternalPublicPath("/feedback", "art-circles")).toBe(
      "/portal/art-circles/feedback",
    );
  });

  it("redirects the legacy portal URL to the canonical short URL", () => {
    expect(
      getCanonicalPublicPath("/portal/art-circles/feedback", "art-circles"),
    ).toBe("/feedback");
  });

  it("recognizes short feedback routes as public", () => {
    expect(isPublicPath("/feedback")).toBe(true);
    expect(isPublicPath("/feedback/roadmap")).toBe(true);
    expect(isPublicPath("/feedback/improve-mobile-navigation")).toBe(true);
    expect(isPublicPath("/roadmap")).toBe(false);
    expect(isPublicPath("/feedback-private")).toBe(false);
    expect(isPublicPath("/embed/feedback/art-circles")).toBe(true);
    expect(isPublicPath("/embed/private")).toBe(false);
  });

  it("maps legacy request paths onto canonical feedback paths", () => {
    expect(
      getCanonicalPublicPath(
        "/portal/art-circles/requests/improve-mobile-navigation",
        "art-circles",
      ),
    ).toBe("/feedback/improve-mobile-navigation");
  });

  it("rewrites feedback roadmap and update paths through the current workspace", () => {
    expect(getInternalPublicPath("/feedback/roadmap", "art-circles")).toBe(
      "/portal/art-circles/feedback/roadmap",
    );
    expect(getInternalPublicPath("/updates", "art-circles")).toBe(
      "/portal/art-circles/updates",
    );
    expect(getInternalPublicPath("/roadmap", "art-circles")).toBeNull();
  });

  it("does not redirect the removed legacy portal roadmap URL", () => {
    expect(
      getCanonicalPublicPath("/portal/art-circles/roadmap", "art-circles"),
    ).toBeNull();
    expect(
      getCanonicalPublicPath(
        "/portal/art-circles/feedback/roadmap",
        "art-circles",
      ),
    ).toBe("/feedback/roadmap");
  });

  it("keeps portal account settings outside workspace application routes", () => {
    expect(getInternalPublicPath("/account", "art-circles")).toBe(
      "/portal/art-circles/account",
    );
    expect(
      getCanonicalPublicPath("/portal/art-circles/account", "art-circles"),
    ).toBe("/account");
  });

  it("keeps feedback detail paths on the workspace subdomain", () => {
    expect(
      getInternalPublicPath(
        "/feedback/improve-mobile-navigation",
        "art-circles",
      ),
    ).toBe("/portal/art-circles/feedback/improve-mobile-navigation");
  });

  it("keeps public contributor profiles on the workspace subdomain", () => {
    expect(
      getInternalPublicPath(
        "/people/00000000-0000-4000-8000-000000000001",
        "art-circles",
      ),
    ).toBe("/portal/art-circles/people/00000000-0000-4000-8000-000000000001");
    expect(
      getCanonicalPublicPath(
        "/portal/art-circles/people/00000000-0000-4000-8000-000000000001",
        "art-circles",
      ),
    ).toBe("/people/00000000-0000-4000-8000-000000000001");
  });
});

describe("authentication availability", () => {
  it("returns a controlled service-unavailable response", async () => {
    mockedGetSessionFromRequest.mockRejectedValue(
      new AuthSessionLookupError("Current-user lookup failed"),
    );

    const response = await proxy({
      headers: { get: () => null },
      nextUrl: {
        hostname: "localhost",
        pathname: "/acme/my-work",
        search: "",
      },
      url: "http://localhost:3000/acme/my-work",
    } as unknown as NextRequest);

    expect(response.status).toBe(503);
    expect(response.headers.get("cache-control")).toBe("no-store");
    expect(response.headers.get("retry-after")).toBe("5");
    await expect(response.text()).resolves.toBe(
      "Authentication service is temporarily unavailable.",
    );
  });
});

describe("workspace home routing", () => {
  const request = (path: string) => {
    const url = new URL(path, "https://acme.fortyone.app");
    return {
      headers: { get: () => null },
      nextUrl: url,
      url: url.href,
    } as unknown as NextRequest;
  };

  it("opens Maya at an authenticated workspace root and preserves the query", async () => {
    mockedGetSessionFromRequest.mockResolvedValue({
      id: "user-a",
    } as NonNullable<Awaited<ReturnType<typeof getSessionFromRequest>>>);
    await proxy(request("/?chatRef=conversation"));
    expect(mockedResponse.rewrite).toHaveBeenCalledWith(
      new URL("https://acme.fortyone.app/acme/maya?chatRef=conversation"),
    );
  });

  it("returns signed-out workspace visitors to Maya after login", async () => {
    mockedGetSessionFromRequest.mockResolvedValue(null);
    await proxy(request("/"));
    const redirectUrl = mockedResponse.redirect.mock.calls[0][0] as URL;
    expect(redirectUrl.searchParams.get("callbackUrl")).toBe(
      "https://acme.fortyone.app/maya",
    );
  });

  it("keeps explicit callbacks when signing in from a workspace root", async () => {
    mockedGetSessionFromRequest.mockResolvedValue(null);
    await proxy(request("/?callbackUrl=%2Fmy-work"));
    const redirectUrl = mockedResponse.redirect.mock.calls[0][0] as URL;
    expect(redirectUrl.searchParams.get("callbackUrl")).toBe("/my-work");
  });

  it("keeps the My Work route available", async () => {
    mockedGetSessionFromRequest.mockResolvedValue({
      id: "user-a",
    } as NonNullable<Awaited<ReturnType<typeof getSessionFromRequest>>>);
    await proxy(request("/my-work?view=board"));
    expect(mockedResponse.rewrite).toHaveBeenCalledWith(
      new URL("https://acme.fortyone.app/acme/my-work?view=board"),
    );
  });
});
