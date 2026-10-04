import { ApiError } from "api-client";
import type { WorkspaceCtx } from "@/lib/http";
import type { DetailedStory } from "@/shared/story/types";
import type { Team } from "@/modules/teams/public/types";
import type { SavedView } from "@/modules/work-presets/public/views";
import { getStory } from "@/modules/story/public/queries";
import { getTeamDetails } from "@/modules/teams/public/queries";
import { resolveSavedViews } from "@/modules/work-presets/public/views";
import { favoriteRequests, loadFavoriteSource } from "./favorites-resolution";

jest.mock("@/modules/story/public/queries", () => ({ getStory: jest.fn() }));
jest.mock("@/modules/teams/public/queries", () => ({
  getTeamDetails: jest.fn(),
}));
jest.mock("@/modules/work-presets/public/views", () => ({
  resolveSavedViews: jest.fn(),
  presetKey: (workspace: string, user: string, team: string, kind: string) => [
    "work-presets",
    workspace,
    user,
    team,
    kind,
  ],
  savedViewPath: (team: string, id: string) =>
    `/teams/${team}/stories?view=${id}`,
}));
const id = "00000000-0000-4000-8000-000000000001";
const teamId = "00000000-0000-4000-8000-000000000002";
const ctx = { workspaceSlug: "acme", session: null } as WorkspaceCtx;
const signal = new AbortController().signal;
const story = {
  id,
  title: "Current title",
  teamId,
  teamCode: "NEW",
  sequenceId: 26,
  statusId: "status",
  deletedAt: null,
  archivedAt: null,
} as DetailedStory;
const team = { id: teamId, name: "Current team", color: "#123456" } as Team;
const view = {
  id,
  kind: "view",
  teamId,
  name: "Current view",
  configuration: { layout: "kanban" },
} as SavedView;

describe("authorized favorite resolution", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("uses account/workspace-specific keys and current canonical story metadata", async () => {
    const request = favoriteRequests(
      [{ kind: "story", id, teamId: "old-team" }],
      "one",
      ctx,
    )[0];
    jest.mocked(getStory).mockResolvedValue(story);
    expect(await loadFavoriteSource(request, signal)).toBe(story);
    expect(request.project(story)).toEqual([
      {
        ref: request.refs[0],
        name: "Current title",
        path: "/work/NEW-26",
        icon: { kind: "story", statusId: "status" },
      },
    ]);
    expect(request.key).not.toEqual(
      favoriteRequests(request.refs, "two", ctx)[0].key,
    );
    expect(request.key).not.toEqual(
      favoriteRequests(request.refs, "one", {
        ...ctx,
        workspaceSlug: "other",
      })[0].key,
    );
    expect(request.project({ ...story, archivedAt: "today" })).toEqual(
      request.project(story),
    );
    expect(request.project({ ...story, deletedAt: "today" })).toEqual([]);
    expect(request.project({ ...story, id: teamId })).toEqual([]);
  });

  it("resolves a team with a throwing cancellable read and fresh color/name", async () => {
    const request = favoriteRequests(
      [{ kind: "team", id: teamId }],
      "one",
      ctx,
    )[0];
    jest.mocked(getTeamDetails).mockResolvedValue(team);
    await loadFavoriteSource(request, signal);
    expect(getTeamDetails).toHaveBeenCalledWith(teamId, ctx, signal);
    expect(request.project([team])).toEqual([
      {
        ref: request.refs[0],
        name: "Current team",
        path: `/teams/${teamId}/stories`,
        icon: { kind: "team", color: "#123456" },
      },
    ]);
    expect(request.project([{ ...team, id }])).toEqual([]);
  });

  it("batches view lookups by team, rejects wrong-team definitions and defaults an unset layout to list", async () => {
    const refs = [
      { kind: "view" as const, id, teamId },
      { kind: "view" as const, id: teamId, teamId },
    ];
    const request = favoriteRequests(refs, "one", ctx)[0];
    jest.mocked(resolveSavedViews).mockResolvedValue([view]);
    await loadFavoriteSource(request, signal);
    expect(resolveSavedViews).toHaveBeenCalledWith(
      teamId,
      [id, teamId],
      ctx,
      signal,
    );
    expect(
      request.project([view, { ...view, id: teamId, teamId: "wrong-team" }]),
    ).toHaveLength(1);
    expect(request.project([view])[0].icon).toEqual({
      kind: "view",
      layout: "kanban",
    });
    expect(
      request.project([
        { ...view, configuration: { ...view.configuration, layout: null } },
      ])[0].icon,
    ).toEqual({ kind: "view", layout: "list" });
  });

  it.each([403, 404, 410])(
    "marks only confirmed missing/denied responses unavailable (%s)",
    async (status) => {
      const request = favoriteRequests(
        [{ kind: "team", id: teamId }],
        "one",
        ctx,
      )[0];
      jest
        .mocked(getTeamDetails)
        .mockRejectedValue(new ApiError("Unavailable", status, null));
      expect(await loadFavoriteSource(request, signal)).toBeNull();
    },
  );

  it.each([
    new ApiError("Unauthenticated", 401, null),
    new ApiError("Server failure", 500, null),
    new Error("Offline"),
  ])(
    "preserves transient/authentication failures for honest retry feedback",
    async (error) => {
      const request = favoriteRequests(
        [{ kind: "team", id: teamId }],
        "one",
        ctx,
      )[0];
      jest.mocked(getTeamDetails).mockRejectedValue(error);
      await expect(loadFavoriteSource(request, signal)).rejects.toBe(error);
    },
  );
  it("resolves the saved view's fresh custom icon without persisting display metadata in refs", async () => {
    const ref = { kind: "view" as const, id, teamId };
    jest
      .mocked(resolveSavedViews)
      .mockResolvedValue([
        { ...view, configuration: { ...view.configuration, icon: "calendar" } },
      ]);
    const request = favoriteRequests([ref], "owner", ctx)[0];
    const source = await request.load(signal);
    expect(request.project(source)[0].icon).toEqual({
      kind: "view",
      layout: "kanban",
      icon: "calendar",
    });
    expect(ref).toEqual({ kind: "view", id, teamId });
  });
});
