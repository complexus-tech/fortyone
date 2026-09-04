import { randomUUID } from "node:crypto";
import type { InfiniteData, QueryKey } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  QueryObserver,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { createObjective } from "../actions/create-objective";
import type { NewObjective, Objective, ObjectivesPage } from "../types";
import { objectiveKeys } from "../constants";
import { useCreateObjectiveMutation } from "./create-mutation";

jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useWorkspacePath: jest.fn(),
}));
jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("../actions/create-objective", () => ({
  createObjective: jest.fn(),
}));

const WORKSPACE_SLUG = "forty-one";
const TEAM_ID = "team-1";
const mockCreateObjective = jest.mocked(createObjective);
const randomUUIDDescriptor = Object.getOwnPropertyDescriptor(
  globalThis.crypto,
  "randomUUID",
);

beforeAll(() => {
  Object.defineProperty(globalThis.crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
});

afterAll(() => {
  if (randomUUIDDescriptor) {
    Object.defineProperty(
      globalThis.crypto,
      "randomUUID",
      randomUUIDDescriptor,
    );
  } else {
    Reflect.deleteProperty(globalThis.crypto, "randomUUID");
  }
});

const objective = (id: string, name = id): Objective => ({
  id,
  name,
  sequenceId: 1,
  description: "",
  shortSummary: null,
  leadUser: "user-1",
  teamId: TEAM_ID,
  workspaceId: "workspace-1",
  startDate: "2026-09-04",
  endDate: "2026-12-31",
  isPrivate: false,
  createdAt: "2026-09-04T10:00:00Z",
  updatedAt: "2026-09-04T10:00:00Z",
  createdBy: "user-1",
  statusId: "status-1",
  keyResultCount: 0,
  health: null,
  color: "#4A90E2",
  forecastStartDate: null,
  forecastEndDate: null,
  scheduleStatus: "no_schedule",
  forecastDaysDelta: 0,
  forecastCauseStory: null,
});

const newObjective = (name: string): NewObjective => ({
  name,
  teamId: TEAM_ID,
  statusId: "status-1",
  isPrivate: true,
});

const paginatedObjectives = (
  objectives: Objective[],
): InfiniteData<ObjectivesPage> => ({
  pages: [
    {
      objectives,
      pagination: { page: 1, pageSize: 15, hasMore: false, nextPage: 0 },
    },
  ],
  pageParams: [1],
});

const deferredResponse = () => {
  let resolve!: (value: Awaited<ReturnType<typeof createObjective>>) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<Awaited<ReturnType<typeof createObjective>>>(
    (resolvePromise, rejectPromise) => {
      resolve = resolvePromise;
      reject = rejectPromise;
    },
  );
  return { promise, resolve, reject };
};

describe("useCreateObjectiveMutation", () => {
  let queryClient: QueryClient;
  let unsubscribe: (() => void)[];

  const observe = (queryKey: QueryKey, data: unknown) => {
    queryClient.setQueryData(queryKey, data);
    const observer = new QueryObserver(queryClient, {
      queryKey,
      queryFn: () =>
        new Promise(() => {
          // Keep reconciliation pending so assertions observe the cache writes.
        }),
      staleTime: Infinity,
    });
    unsubscribe.push(observer.subscribe(() => undefined));
  };

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    jest.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        mutations: { retry: false },
        queries: { retry: false, gcTime: Infinity },
      },
    });
    unsubscribe = [];
    jest.mocked(useWorkspacePath).mockReturnValue({
      workspaceSlug: WORKSPACE_SLUG,
    } as ReturnType<typeof useWorkspacePath>);
    jest.mocked(useAnalytics).mockReturnValue({
      analytics: { track: jest.fn() },
    } as unknown as ReturnType<typeof useAnalytics>);
    jest.mocked(useSession).mockReturnValue({
      data: { user: { id: "user-1" } },
    } as ReturnType<typeof useSession>);
  });

  afterEach(() => {
    unsubscribe.forEach((stop) => {
      stop();
    });
    queryClient.clear();
  });

  it("updates only the matching workspace and team's collection shapes", async () => {
    const existing = objective("existing");
    const listKey = [...objectiveKeys.list(WORKSPACE_SLUG), ""];
    const teamKey = [...objectiveKeys.team(WORKSPACE_SLUG, TEAM_ID), ""];
    const infiniteKey = [
      ...objectiveKeys.team(WORKSPACE_SLUG, TEAM_ID),
      "infinite",
      "",
      15,
    ];
    const untouchedEntries: [QueryKey, unknown][] = [
      [objectiveKeys.objective(WORKSPACE_SLUG, existing.id), existing],
      [
        objectiveKeys.activitiesInfinite(WORKSPACE_SLUG, existing.id),
        { pages: [{ activities: [] }], pageParams: [1] },
      ],
      [[...objectiveKeys.list("other-workspace"), ""], [existing]],
      [[...objectiveKeys.team(WORKSPACE_SLUG, "other-team"), ""], []],
      [[...objectiveKeys.list(WORKSPACE_SLUG), "unrelated search"], []],
    ];
    observe(listKey, [existing]);
    observe(teamKey, [existing]);
    observe(infiniteKey, paginatedObjectives([existing]));
    untouchedEntries.forEach(([key, data]) => {
      observe(key, data);
    });
    const response = deferredResponse();
    mockCreateObjective.mockReturnValueOnce(response.promise);
    const { result } = renderHook(useCreateObjectiveMutation, { wrapper });

    act(() => {
      result.current.mutate(newObjective("Launch"));
    });
    await waitFor(() => {
      expect(mockCreateObjective).toHaveBeenCalledTimes(1);
    });

    const optimistic = queryClient.getQueryData<Objective[]>(listKey)?.at(-1);
    expect(optimistic).toMatchObject({ name: "Launch", isPrivate: true });
    expect(optimistic?.id).toMatch(/^optimistic:/);
    expect(queryClient.getQueryData<Objective[]>(teamKey)?.at(-1)).toEqual(
      optimistic,
    );
    const optimisticPage =
      queryClient.getQueryData<InfiniteData<ObjectivesPage>>(infiniteKey);
    expect(optimisticPage?.pages[0]?.objectives).toEqual([
      existing,
      optimistic,
    ]);
    expect(optimisticPage?.pageParams).toEqual([1]);
    untouchedEntries.forEach(([key, data]) => {
      expect(queryClient.getQueryData(key)).toEqual(data);
    });

    const created = objective("created", "Launch");
    act(() => {
      response.resolve({ data: { objective: created } });
    });
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(queryClient.getQueryData(listKey)).toEqual([existing, created]);
    expect(
      queryClient.getQueryData<InfiniteData<ObjectivesPage>>(infiniteKey)
        ?.pages[0]?.objectives,
    ).toEqual([existing, created]);
    expect(
      queryClient.getQueryState([...objectiveKeys.list("other-workspace"), ""])
        ?.isInvalidated,
    ).toBe(false);
  });

  it("rolls back only the failed create while preserving concurrent and server updates", async () => {
    const listKey = [...objectiveKeys.list(WORKSPACE_SLUG), ""];
    const infiniteKey = [
      ...objectiveKeys.team(WORKSPACE_SLUG, TEAM_ID),
      "infinite",
      "",
      15,
    ];
    const existing = objective("existing");
    observe(listKey, [existing]);
    observe(infiniteKey, paginatedObjectives([existing]));
    const first = deferredResponse();
    const second = deferredResponse();
    mockCreateObjective
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise);
    const invalidate = jest.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(useCreateObjectiveMutation, { wrapper });

    act(() => {
      result.current.mutate(newObjective("First"));
      result.current.mutate(newObjective("Second"));
    });
    await waitFor(() => {
      expect(mockCreateObjective).toHaveBeenCalledTimes(2);
    });

    const pending = queryClient.getQueryData<Objective[]>(listKey) ?? [];
    expect(pending.map(({ name }) => name)).toEqual([
      "existing",
      "First",
      "Second",
    ]);
    expect(new Set(pending.map(({ id }) => id)).size).toBe(3);

    const created = objective("second-created", "Second");
    act(() => {
      second.resolve({ data: { objective: created } });
    });
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });
    expect(invalidate).not.toHaveBeenCalled();

    const external = objective("external-update");
    queryClient.setQueryData<Objective[]>(listKey, (data = []) => [
      ...data,
      external,
    ]);
    act(() => {
      first.reject(new Error("Creation rejected"));
    });
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledTimes(1);
    });

    expect(queryClient.getQueryData(listKey)).toEqual([
      existing,
      created,
      external,
    ]);
    expect(
      queryClient.getQueryData<InfiniteData<ObjectivesPage>>(infiniteKey)
        ?.pages[0]?.objectives,
    ).toEqual([existing, created]);
    expect(invalidate).toHaveBeenCalledTimes(1);
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: objectiveKeys.list(WORKSPACE_SLUG),
    });
    expect(mockCreateObjective).toHaveBeenCalledTimes(2);
  });
});
