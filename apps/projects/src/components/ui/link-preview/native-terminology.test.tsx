import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { teamKeys, workspaceKeys } from "@/constants/keys";
import { useTerminology } from "@/hooks/use-terminology-display";
import { getWorkspaceSettings } from "@/lib/queries/workspaces/get-settings";
import { getTeamDetails } from "@/modules/teams/queries/get-team";
import type { Team } from "@/modules/teams/types";
import type { DetailedStory } from "@/shared/story/types";
import type { Objective } from "@/shared/objectives/types";
import type { WorkspaceSettings } from "@/types";
import type { NativeLinkResource } from "./native-query";
import type { NativeLinkTarget } from "./native-target";
import { useNativePreviewTerminology } from "./native-terminology";

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: jest.fn(),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "member" } } }),
}));
jest.mock("@/lib/queries/workspaces/get-settings", () => ({
  getWorkspaceSettings: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-team", () => ({
  getTeamDetails: jest.fn(),
}));

const localTarget: NativeLinkTarget = {
  kind: "story",
  workspaceSlug: "acme",
  id: "ENG-42",
  href: "https://acme.fortyone.app/work/ENG-42",
};
const remoteTarget: NativeLinkTarget = {
  ...localTarget,
  workspaceSlug: "other",
  href: "https://other.fortyone.app/work/ENG-42",
};
const story: NativeLinkResource = {
  kind: "story",
  value: { teamId: "engineering" } as DetailedStory,
};
const settings = {
  storyTerm: "task",
  objectiveTerm: "goal",
} as WorkspaceSettings;
let client: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);

beforeEach(() => {
  jest.clearAllMocks();
  client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  jest.mocked(useTerminology).mockImplementation((teamId) => ({
    getTermDisplay: (key) => {
      if (key === "objectiveTerm") return "project";
      return teamId === "engineering" ? "ticket" : "task";
    },
  }));
  jest.mocked(getWorkspaceSettings).mockResolvedValue(settings);
  jest.mocked(getTeamDetails).mockResolvedValue({
    id: "engineering",
    storyTerm: "deal",
  } as Team);
});

afterEach(() => {
  client.clear();
});

it("uses the resource team in the active workspace and avoids destination requests", () => {
  const { result } = renderHook(
    () => useNativePreviewTerminology(localTarget, story),
    { wrapper },
  );
  expect(result.current).toEqual({ label: "Ticket", term: "ticket" });
  expect(useTerminology).toHaveBeenCalledWith("engineering");
  expect(getWorkspaceSettings).not.toHaveBeenCalled();
  expect(getTeamDetails).not.toHaveBeenCalled();
});

it("uses the workspace default while the resource team is unresolved", () => {
  const { result } = renderHook(
    () => useNativePreviewTerminology(localTarget, null),
    { wrapper },
  );
  expect(result.current).toEqual({ label: "Task", term: "task" });
  expect(useTerminology).toHaveBeenCalledWith(null);
});

it("respects objective terminology without using the story team override", () => {
  const target: NativeLinkTarget = {
    kind: "objective",
    id: "objective",
    teamId: "engineering",
    workspaceSlug: "acme",
    href: "https://acme.fortyone.app/teams/engineering/objectives/objective",
  };
  const { result } = renderHook(
    () =>
      useNativePreviewTerminology(target, {
        kind: "objective",
        value: { teamId: "engineering" } as Objective,
      }),
    { wrapper },
  );
  expect(result.current).toEqual({ label: "Project", term: "project" });
});

it("uses neutral labels before authorization even if destination metadata is cached", () => {
  client.setQueryData(workspaceKeys.settings("other"), settings);
  client.setQueryData(teamKeys.lists("other"), [
    { id: "engineering", storyTerm: "deal" },
  ]);
  const { result } = renderHook(
    () => useNativePreviewTerminology(remoteTarget, null),
    { wrapper },
  );
  expect(result.current).toEqual({ label: "Item", term: "item" });
  expect(getWorkspaceSettings).not.toHaveBeenCalled();
  expect(getTeamDetails).not.toHaveBeenCalled();
});

it("resolves authorized destination settings and team overrides through scoped queries", async () => {
  const { result } = renderHook(
    () => useNativePreviewTerminology(remoteTarget, story),
    { wrapper },
  );
  expect(result.current).toEqual({ label: "Item", term: "item" });
  await waitFor(() => {
    expect(result.current).toEqual({ label: "Deal", term: "deal" });
  });
  expect(getWorkspaceSettings).toHaveBeenCalledWith(
    expect.objectContaining({ workspaceSlug: "other" }),
  );
  expect(getTeamDetails).toHaveBeenCalledWith(
    "engineering",
    expect.objectContaining({ workspaceSlug: "other" }),
    expect.any(AbortSignal),
  );
  expect(client.getQueryData(workspaceKeys.settings("other"))).toEqual(
    settings,
  );
  expect(client.getQueryData(teamKeys.detail("other", "engineering"))).toEqual({
    id: "engineering",
    storyTerm: "deal",
  });
});

it("reuses destination metadata and reacts to team terminology changes", () => {
  client.setQueryData(workspaceKeys.settings("other"), settings);
  client.setQueryData(teamKeys.lists("other"), [
    { id: "engineering", storyTerm: null },
  ]);
  const { result } = renderHook(
    () => useNativePreviewTerminology(remoteTarget, story),
    { wrapper },
  );
  expect(result.current).toEqual({ label: "Task", term: "task" });
  act(() => {
    client.setQueryData(teamKeys.lists("other"), [
      { id: "engineering", storyTerm: "work item" },
    ]);
  });
  expect(result.current).toEqual({ label: "Work item", term: "work item" });
  expect(getWorkspaceSettings).not.toHaveBeenCalled();
  expect(getTeamDetails).not.toHaveBeenCalled();
});

it("uses destination objective terminology without fetching a team", async () => {
  const { result } = renderHook(
    () =>
      useNativePreviewTerminology(
        {
          ...remoteTarget,
          kind: "objective",
          teamId: "engineering",
        },
        { kind: "objective", value: {} as Objective },
      ),
    { wrapper },
  );
  await waitFor(() => {
    expect(result.current).toEqual({ label: "Goal", term: "goal" });
  });
  expect(getTeamDetails).not.toHaveBeenCalled();
});

it("keeps a neutral label if destination terminology cannot be resolved", async () => {
  jest.mocked(getWorkspaceSettings).mockRejectedValue(new Error("Unavailable"));
  const { result } = renderHook(
    () => useNativePreviewTerminology(remoteTarget, story),
    { wrapper },
  );
  await waitFor(() => {
    expect(client.getQueryState(workspaceKeys.settings("other"))?.status).toBe(
      "error",
    );
  });
  expect(result.current).toEqual({ label: "Item", term: "item" });
});
