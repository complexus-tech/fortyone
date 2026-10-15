import { QueryClient } from "@tanstack/react-query";
import { memberKeys, teamKeys } from "@/constants/keys";
import type { WorkspaceCtx } from "@/lib/http";
import { getJoinedTeams, getTeams } from "@/modules/teams/queries/get-teams";
import { getTeamCustomFields } from "@/modules/custom-fields/api";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { customFieldKeys } from "@/modules/custom-fields/hooks";
import { warmDesktopSettings } from "./startup-queries";

jest.mock("@/lib/queries/users/profile", () => ({
  getProfile: async () => ({ id: "account" }),
}));
jest.mock("@/lib/queries/workspaces/get-settings", () => ({
  getWorkspaceSettings: async () => ({}),
}));
jest.mock("@/lib/queries/members/get-members", () => ({
  getMembers: async () => [],
  getTeamMembers: jest.fn(async () => []),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: async () => [],
}));
jest.mock("@/lib/queries/labels/get-labels", () => ({
  getLabels: async () => [],
}));
jest.mock("@/lib/queries/users/automation-preferences", () => ({
  getAutomationPreferences: async () => ({}),
}));
jest.mock("@/modules/notifications/queries/get-preferences", () => ({
  getNotificationPreferences: async () => ({}),
}));
jest.mock("@/modules/teams/queries/get-team-settings", () => ({
  getTeamSettings: async () => ({ estimationSettings: { scheme: "points" } }),
}));
jest.mock("@/modules/stories/queries/get-grouped-stories", () => ({
  getGroupedStories: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-teams", () => ({
  getJoinedTeams: jest.fn(),
  getTeams: jest.fn(),
}));
jest.mock("@/modules/custom-fields/api", () => ({
  getTeamCustomFields: jest.fn(),
}));
jest.mock("@/modules/custom-fields/hooks", () => ({
  customFieldKeys: {
    team: (slug: string, id: string) => ["custom-fields", slug, "team", id],
  },
}));

it("warms the task form's custom definitions, person choices and team settings using its actual keys", async () => {
  jest
    .mocked(getTeams)
    .mockResolvedValue([{ id: "joined" }, { id: "public" }] as Awaited<
      ReturnType<typeof getTeams>
    >);
  jest
    .mocked(getJoinedTeams)
    .mockResolvedValue([{ id: "joined" }] as Awaited<
      ReturnType<typeof getJoinedTeams>
    >);
  jest
    .mocked(getTeamCustomFields)
    .mockResolvedValue([{ id: "field", type: "person" }] as Awaited<
      ReturnType<typeof getTeamCustomFields>
    >);
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const ctx = {
    workspaceSlug: "acme",
    session: { user: { id: "account" } },
  } as WorkspaceCtx;
  expect(await warmDesktopSettings(client, ctx)).toEqual([]);
  expect(client.getQueryData(customFieldKeys.team("acme", "joined"))).toEqual([
    { id: "field", type: "person" },
  ]);
  expect(
    client.getQueryData([...memberKeys.team("acme", "joined"), ""]),
  ).toEqual([]);
  expect(client.getQueryData(teamKeys.settings("acme", "joined"))).toEqual({
    estimationSettings: { scheme: "points" },
  });
  expect(getTeamCustomFields).toHaveBeenCalledTimes(1);
  expect(getTeamMembers).toHaveBeenCalledWith("joined", ctx, "");
});
