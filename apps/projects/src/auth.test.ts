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
});

describe("browser auth session", () => {
  it("does not access Next request cookies in the browser", async () => {
    mockedGetCurrentUser.mockResolvedValue(null);

    await expect(auth()).resolves.toBeNull();
    expect(mockedGetCurrentUser).toHaveBeenCalledTimes(1);
    expect(mockedCookies).not.toHaveBeenCalled();
  });

  it("copies the current user's internal flag into the session", async () => {
    mockedGetCurrentUser.mockResolvedValue({
      avatarUrl: null,
      createdAt: "2026-06-18T00:00:00Z",
      email: "maya@fortyone.app",
      fullName: "Maya Internal",
      hasSeenWalkthrough: true,
      id: "user-1",
      isActive: true,
      isInternal: true,
      lastUsedWorkspaceId: "workspace-1",
      timezone: "Africa/Harare",
      updatedAt: "2026-06-18T00:00:00Z",
      username: "maya",
    });

    await expect(auth()).resolves.toMatchObject({
      user: {
        isInternal: true,
      },
    });
    expect(mockedCookies).not.toHaveBeenCalled();
  });

  it("surfaces browser lookup failures without server control-flow handling", async () => {
    const error = new Error("current-user lookup failed");
    mockedGetCurrentUser.mockRejectedValue(error);

    await expect(auth()).rejects.toBe(error);
    expect(mockedRethrowNextControlFlow).not.toHaveBeenCalled();
    expect(mockedCookies).not.toHaveBeenCalled();
  });
});
