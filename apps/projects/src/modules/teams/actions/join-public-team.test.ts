/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { post } from "@/lib/http";
import { joinPublicTeamAction } from "./join-public-team";

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http", () => ({
  post: jest.fn(),
}));

jest.mock("@/utils", () => ({
  getApiError: jest.fn(),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);

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

describe("joinPublicTeamAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    postMock.mockResolvedValue({ data: { teamId: "team-1" } });
  });

  it("sends no caller-selected user ID", async () => {
    await joinPublicTeamAction("team-1", "complexus");

    expect(postMock).toHaveBeenCalledWith(
      "teams/team-1/join",
      {},
      { session, workspaceSlug: "complexus" },
    );
  });

  it("does not call the API without an authenticated session", async () => {
    authMock.mockResolvedValue(null);

    await expect(joinPublicTeamAction("team-1", "complexus")).resolves.toEqual({
      data: null,
      error: { message: "Authentication required to join teams" },
    });
    expect(postMock).not.toHaveBeenCalled();
  });
});
