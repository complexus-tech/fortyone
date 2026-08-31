/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this exact file-level pragma.

/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { getCurrentUser } from "auth";
import { cookies } from "next/headers";
import { unstable_rethrow as rethrowNextControlFlow } from "next/navigation";
import { auth } from "./auth";

jest.mock("auth", () => ({
  SESSION_COOKIE_NAME: "fortyone_session",
  getCurrentUser: jest.fn(),
}));
jest.mock("next/headers", () => ({
  cookies: jest.fn(),
}));
jest.mock("next/navigation", () => ({
  unstable_rethrow: jest.fn(),
}));

const mockedGetCurrentUser = jest.mocked(getCurrentUser);
const mockedCookies = jest.mocked(cookies);
const mockedRethrowNextControlFlow = jest.mocked(rethrowNextControlFlow);

beforeEach(() => {
  jest.clearAllMocks();
  mockedCookies.mockResolvedValue({
    get: jest.fn(() => ({ value: "session" })),
  } as never);
});

describe("server auth session", () => {
  it("does not contact the auth upstream without a session cookie", async () => {
    mockedCookies.mockResolvedValue({
      get: jest.fn(() => undefined),
    } as never);

    await expect(auth()).resolves.toBeNull();
    expect(mockedGetCurrentUser).not.toHaveBeenCalled();
  });

  it("resolves the current user inside a request scope", async () => {
    mockedGetCurrentUser.mockResolvedValue({
      avatarUrl: null,
      createdAt: "2026-08-31T12:00:00Z",
      email: "joseph@example.com",
      fullName: "Joseph",
      hasSeenWalkthrough: true,
      id: "user-1",
      isActive: true,
      isInternal: false,
      lastUsedWorkspaceId: "workspace-1",
      timezone: "Africa/Harare",
      updatedAt: "2026-08-31T12:00:00Z",
      username: "joseph",
    });

    await expect(auth()).resolves.toMatchObject({
      user: { id: "user-1" },
    });
  });

  it("lets Next.js inspect wrapped rendering control-flow errors", async () => {
    const error = new Error("current-user lookup failed");
    mockedGetCurrentUser.mockRejectedValue(error);

    await expect(auth()).rejects.toBe(error);
    expect(mockedRethrowNextControlFlow).toHaveBeenCalledWith(error);
  });
});
