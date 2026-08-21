/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Session } from "@/auth";
import { getJoinedTeams } from "@/modules/teams/queries/get-teams";
import type { Team } from "@/modules/teams/types";
import { resolveJoinedTeams } from "./resolve-joined-teams";

jest.mock("@/modules/teams/queries/get-teams", () => ({
  getJoinedTeams: jest.fn(),
}));

const getJoinedTeamsMock = jest.mocked(getJoinedTeams);
const session: Session = {
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
const product: Team = {
  id: "team-1",
  name: "Product",
  code: "PROD",
  color: "#6366F1",
  isPrivate: false,
  workspaceId: "workspace-1",
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-01T08:00:00.000Z",
  memberCount: 2,
  sprintsEnabled: true,
};

describe("resolveJoinedTeams", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("loads authoritative memberships for the current workspace", async () => {
    getJoinedTeamsMock.mockResolvedValue([product]);

    await expect(
      resolveJoinedTeams({ session, workspaceSlug: "complexus" }),
    ).resolves.toEqual([product]);
    expect(getJoinedTeamsMock).toHaveBeenCalledWith({
      session,
      workspaceSlug: "complexus",
    });
  });

  it("returns unknown when request context cannot identify a workspace user", async () => {
    await expect(
      resolveJoinedTeams({ session: null, workspaceSlug: "complexus" }),
    ).resolves.toBeNull();
    await expect(
      resolveJoinedTeams({ session, workspaceSlug: undefined }),
    ).resolves.toBeNull();
    expect(getJoinedTeamsMock).not.toHaveBeenCalled();
  });

  it("returns unknown and records the failure when memberships cannot load", async () => {
    const error = new Error("backend unavailable");
    const consoleError = jest.spyOn(console, "error").mockImplementation();
    getJoinedTeamsMock.mockRejectedValue(error);

    await expect(
      resolveJoinedTeams({ session, workspaceSlug: "complexus" }),
    ).resolves.toBeNull();
    expect(consoleError).toHaveBeenCalledWith(
      "Failed to resolve joined teams for Maya",
      error,
    );

    consoleError.mockRestore();
  });
});
