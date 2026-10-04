import type { DehydratedState } from "@tanstack/react-query";
import {
  defaultShouldDehydrateQuery,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { getQueryClient } from "@/app/get-query-client";
import { memberKeys, teamKeys } from "@/constants/keys";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { getTeamSettings } from "@/modules/teams/queries/get-team-settings";
import TeamManagementPage from "./page";

jest.mock("@/auth", () => ({
  auth: async () => ({ token: "session" }),
}));
jest.mock("@/app/get-query-client", () => ({
  getQueryClient: jest.fn(),
}));
jest.mock("@/lib/queries/members/get-members", () => ({
  getTeamMembers: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-team-settings", () => ({
  getTeamSettings: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-team", () => ({
  getTeam: jest.fn(),
}));
jest.mock("@/modules/settings/workspace/teams/management", () => ({
  TeamManagement: () => <div>Team settings</div>,
}));

const teamId = "team-one";
const workspaceSlug = "first";
const teamMembersKey = memberKeys.team(workspaceSlug, teamId);
const teamSettingsKey = teamKeys.settings(workspaceSlug, teamId);
const pendingKey = ["shared-streaming-query", workspaceSlug];
const timestamp = "2026-10-04T09:00:00Z";
const teamSettings: Awaited<ReturnType<typeof getTeamSettings>> = {
  sprintSettings: {
    autoCreateSprints: false,
    upcomingSprintsCount: 1,
    sprintDurationWeeks: 2,
    sprintStartDay: "monday",
    workingDays: [1, 2, 3, 4, 5],
    moveIncompleteStoriesEnabled: false,
    nextAutoSprintNumber: 1,
    autoCreateDisabledAt: null,
    autoCreateDisabledReason: null,
    createdAt: timestamp,
    updatedAt: timestamp,
  },
  storyAutomationSettings: {
    autoCloseInactiveEnabled: false,
    autoCloseInactiveMonths: 3,
    autoArchiveEnabled: false,
    autoArchiveMonths: 3,
    createdAt: timestamp,
    updatedAt: timestamp,
  },
  estimationSettings: {
    scheme: "points",
    createdAt: timestamp,
    updatedAt: timestamp,
  },
};
let serverClient: QueryClient;
let browserClient: QueryClient;

beforeEach(() => {
  jest.clearAllMocks();
  serverClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      dehydrate: {
        shouldDehydrateQuery: (query) =>
          defaultShouldDehydrateQuery(query) ||
          query.state.status === "pending",
      },
    },
  });
  browserClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  jest.mocked(getQueryClient).mockReturnValue(serverClient);
  jest.mocked(getTeamMembers).mockResolvedValue([]);
  jest.mocked(getTeamSettings).mockResolvedValue(teamSettings);
});

afterEach(() => {
  serverClient.clear();
  browserClient.clear();
});

it("hydrates only the settled queries this page awaited, without repeating workspace streams", async () => {
  serverClient.setQueryData(["workspace-shell", workspaceSlug], {
    name: "First workspace",
  });
  void serverClient.prefetchQuery({
    queryKey: pendingKey,
    queryFn: () => new Promise(() => {}),
  });
  void browserClient.prefetchQuery({
    queryKey: pendingKey,
    queryFn: () => new Promise(() => {}),
  });

  const page = await TeamManagementPage({
    params: Promise.resolve({ workspaceSlug, teamId }),
  });
  const state = page.props.state as DehydratedState;
  expect(state.queries.map((query) => query.queryKey)).toEqual([
    teamMembersKey,
    teamSettingsKey,
  ]);
  render(
    <QueryClientProvider client={browserClient}>{page}</QueryClientProvider>,
  );
  expect(screen.getByText("Team settings")).toBeInTheDocument();
  expect(browserClient.getQueryData(teamMembersKey)).toEqual([]);
  expect(browserClient.getQueryData(teamSettingsKey)).toEqual(teamSettings);
  expect(browserClient.getQueryState(pendingKey)?.fetchStatus).toBe("fetching");
});

it("omits failed page prefetches so client queries can recover normally", async () => {
  jest.mocked(getTeamSettings).mockRejectedValue(new Error("API unavailable"));
  const page = await TeamManagementPage({
    params: Promise.resolve({ workspaceSlug, teamId }),
  });
  const state = page.props.state as DehydratedState;
  expect(state.queries.map((query) => query.queryKey)).toEqual([
    teamMembersKey,
  ]);
});
