/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { remove } from "@/lib/http";
import { leaveTeamAction } from "./leave-team";

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http", () => ({
  remove: jest.fn(),
}));

jest.mock("@/utils", () => ({
  getApiError: jest.fn(),
}));

const authMock = jest.mocked(auth);
const removeMock = jest.mocked(remove);

const session = {
  user: {
    id: "user-1",
    name: "Joseph Mukorivo",
    email: "joseph@example.com",
    image: null,
    username: "joseph",
    fullName: "Joseph Mukorivo",
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
  },
};

describe("leaveTeamAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    removeMock.mockResolvedValue({ data: null });
  });

  it("uses the authenticated self-membership route without a user ID", async () => {
    await leaveTeamAction("team-1", "complexus");

    expect(removeMock).toHaveBeenCalledWith("teams/team-1/membership", {
      session,
      workspaceSlug: "complexus",
    });
  });

  it("does not call the API without an authenticated session", async () => {
    authMock.mockResolvedValue(null);

    await expect(leaveTeamAction("team-1", "complexus")).resolves.toEqual({
      data: null,
      error: { message: "Authentication required to leave teams" },
    });
    expect(removeMock).not.toHaveBeenCalled();
  });
});
