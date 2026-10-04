import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { statusKeys } from "@/constants/keys";
import { storyKeys } from "@/shared/story/cache-keys";
import { wipCapacityToastId } from "@/shared/story/wip-capacity";
import type { State } from "@/types/states";
import type { GroupedStoriesResponse } from "@/modules/stories/types";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import type { DetailedStory } from "../types";
import { duplicateStoryAction } from "../actions/duplicate-story";
import { useDuplicateStoryMutation } from "./duplicate-mutation";

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));
jest.mock("next/navigation", () => ({ useRouter: jest.fn() }));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn(), warning: jest.fn() },
}));
jest.mock("posthog-js", () => ({
  __esModule: true,
  default: { captureException: jest.fn() },
}));
jest.mock("../actions/duplicate-story", () => ({
  duplicateStoryAction: jest.fn(),
}));
jest.mock("@/modules/stories/queries/get-grouped-stories", () => ({
  getGroupedStories: jest.fn(),
}));

const WORKSPACE = "forty-one";
const TEAM = "team-one";
const STATUS = "started";
const params = { groupBy: "status", teamIds: [TEAM] } as const;
const countKey = [...storyKeys.grouped(WORKSPACE), params];
const sourceStory: DetailedStory = {
  id: "source",
  sequenceId: 1,
  title: "Source task",
  estimateLabel: null,
  estimateValue: null,
  estimatedDurationMinutes: null,
  minimumFocusBlockMinutes: null,
  autoSchedulingEnabled: false,
  autoSchedulingLocked: false,
  autoSchedulingStatus: "off",
  autoSchedulingReason: null,
  autoSchedulingUpdatedAt: null,
  description: "",
  descriptionHTML: "",
  teamId: TEAM,
  teamCode: "ENG",
  workspaceId: "workspace-one",
  objectiveId: null,
  keyResultId: null,
  statusId: STATUS,
  assigneeId: null,
  collaboratorIds: [],
  collaborators: [],
  collaboratorCount: 0,
  watcherCount: 0,
  watchers: [],
  isWatching: false,
  watchingReason: null,
  reporterId: "user-one",
  priority: "Medium",
  sprintId: null,
  epicId: null,
  startDate: null,
  endDate: null,
  createdAt: "2026-10-04T00:00:00.000Z",
  updatedAt: "2026-10-04T00:00:00.000Z",
  completedAt: null,
  labels: [],
  associations: [],
  parentId: "",
  estimateScheme: "points",
  subStories: [],
  archivedAt: null,
  deletedAt: null,
};
const duplicate = {
  ...sourceStory,
  id: "duplicate",
  title: "Source task (Copy)",
  teamCode: "ENG",
  sequenceId: 12,
};
const serverCounts = (count: number): GroupedStoriesResponse => ({
  groups: [
    {
      key: STATUS,
      totalCount: count,
      loadedCount: 0,
      stories: [],
      hasMore: true,
      nextPage: 1,
    },
  ],
  meta: {
    groupBy: "status",
    totalGroups: 1,
    filters: {},
    orderBy: "created",
    orderDirection: "desc",
  },
});
const push = jest.fn();
const track = jest.fn();
const createClient = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  client.setQueryData(statusKeys.team(WORKSPACE, TEAM), [
    { id: STATUS, teamId: TEAM, name: "In progress", wipLimit: 5 } as State,
  ]);
  client.setQueryData(countKey, serverCounts(4));
  return client;
};
const wrapperFor = (client: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  };

beforeEach(() => {
  jest.clearAllMocks();
  jest
    .mocked(useRouter)
    .mockReturnValue({ push } as unknown as ReturnType<typeof useRouter>);
  jest.mocked(useWorkspacePath).mockReturnValue({
    workspaceSlug: WORKSPACE,
    withWorkspace: (path: string) => `/${WORKSPACE}${path}`,
  } as ReturnType<typeof useWorkspacePath>);
  jest
    .mocked(useAnalytics)
    .mockReturnValue({ analytics: { track } } as unknown as ReturnType<
      typeof useAnalytics
    >);
  jest.mocked(useTerminology).mockReturnValue({
    getTermDisplay: (_term, options) => (options?.capitalize ? "Task" : "task"),
  } as ReturnType<typeof useTerminology>);
  jest.mocked(getGroupedStories).mockResolvedValue(serverCounts(6));
});

it("refreshes only after a persisted duplicate and updates its existing toast while preserving View and analytics", async () => {
  const client = createClient();
  const competingFetch = jest.fn(async () => serverCounts(99));
  const observer = new QueryObserver(client, {
    queryKey: countKey,
    queryFn: competingFetch,
    staleTime: Infinity,
  });
  const unsubscribe = observer.subscribe(() => undefined);
  let persist: (() => void) | undefined;
  jest.mocked(duplicateStoryAction).mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        persist = () => {
          resolve({ data: duplicate });
        };
      }),
  );
  const { result } = renderHook(() => useDuplicateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  act(() => {
    result.current.mutate({ storyId: sourceStory.id, story: sourceStory });
  });
  await waitFor(() => {
    expect(duplicateStoryAction).toHaveBeenCalled();
  });
  expect(getGroupedStories).not.toHaveBeenCalled();
  expect(toast.warning).not.toHaveBeenCalled();
  act(() => {
    persist?.();
  });
  await waitFor(() => {
    expect(toast.warning).toHaveBeenCalledTimes(1);
  });
  expect(toast.success).toHaveBeenCalledTimes(1);
  const successOptions = jest.mocked(toast.success).mock.calls[0]?.[1];
  const warningOptions = jest.mocked(toast.warning).mock.calls[0]?.[1];
  expect(successOptions?.id).toBe(wipCapacityToastId(WORKSPACE, STATUS));
  expect(warningOptions?.id).toBe(successOptions?.id);
  expect(warningOptions?.action).toBe(successOptions?.action);
  expect(toast.warning).toHaveBeenCalledWith(
    "In progress is over capacity: 6/5",
    expect.objectContaining({
      description: "Reduce by 1 to get within the limit.",
    }),
  );
  const viewAction = warningOptions?.action as {
    onClick: () => void;
    label: string;
  };
  expect(viewAction.label).toBe("View task");
  act(() => {
    viewAction.onClick();
  });
  expect(push).toHaveBeenCalledWith(`/${WORKSPACE}/work/ENG-12`);
  expect(track).toHaveBeenCalledWith("story_created", {
    storyId: duplicate.id,
    title: duplicate.title,
    teamId: TEAM,
    hasObjective: false,
    hasSprint: false,
  });
  expect(
    client.getQueryData<GroupedStoriesResponse>(countKey)?.groups[0]
      ?.totalCount,
  ).toBe(6);
  expect(competingFetch).not.toHaveBeenCalled();
  expect(result.current.isSuccess).toBe(true);
  unsubscribe();
});

it("keeps the existing success toast without a warning when duplication reaches the limit", async () => {
  const client = createClient();
  jest.mocked(duplicateStoryAction).mockResolvedValueOnce({ data: duplicate });
  jest.mocked(getGroupedStories).mockResolvedValueOnce(serverCounts(5));
  const { result } = renderHook(() => useDuplicateStoryMutation(), {
    wrapper: wrapperFor(client),
  });
  await act(async () =>
    result.current.mutateAsync({ storyId: sourceStory.id, story: sourceStory }),
  );
  await waitFor(() => {
    expect(getGroupedStories).toHaveBeenCalledTimes(1);
  });
  expect(toast.success).toHaveBeenCalledWith(
    "Success",
    expect.objectContaining({ description: "Task duplicated successfully" }),
  );
  expect(toast.warning).not.toHaveBeenCalled();
});

it.each(["rejected request", "returned API error"])(
  "preserves failure feedback without refreshing or warning for a %s",
  async (failure) => {
    const client = createClient();
    if (failure === "rejected request") {
      jest
        .mocked(duplicateStoryAction)
        .mockRejectedValueOnce(new Error("Duplicate failed"));
    } else {
      jest.mocked(duplicateStoryAction).mockResolvedValueOnce({
        data: null,
        error: { message: "Duplicate failed" },
      });
    }
    const { result } = renderHook(() => useDuplicateStoryMutation(), {
      wrapper: wrapperFor(client),
    });
    act(() => {
      result.current.mutate({ storyId: sourceStory.id, story: sourceStory });
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(toast.error).toHaveBeenCalledWith(
      "Failed to duplicate task",
      expect.objectContaining({
        description: "Duplicate failed",
        action: expect.objectContaining({ label: "Retry" }),
      }),
    );
    expect(toast.success).not.toHaveBeenCalled();
    expect(toast.warning).not.toHaveBeenCalled();
    expect(getGroupedStories).not.toHaveBeenCalled();
    expect(track).not.toHaveBeenCalled();
  },
);
