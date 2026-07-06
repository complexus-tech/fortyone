/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { getCurrentUser } from "auth";
import { auth } from "./auth";

jest.mock("auth", () => ({
  getCurrentUser: jest.fn(),
}));

const mockedGetCurrentUser = jest.mocked(getCurrentUser);

describe("auth session", () => {
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
  });
});
