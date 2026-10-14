/* global afterEach, beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, render, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import { objectiveKeys } from "@/modules/objectives/constants";
import { getKeyResults } from "@/modules/objectives/queries/get-key-results";
import { StrategyMapCanvas } from "./strategy-map-canvas";
import { CanvasConnections } from "./strategy-map-canvas-renderers";
import { useStrategyKeyResults } from "./use-strategy-key-results";
import {
  createConnectionPath,
  createStrategyMapLayout,
  getStrategyConnections,
} from "./strategy-map-layout";
import type { StrategyMap } from "./types";
import type * as StrategyMapLayoutModule from "./strategy-map-layout";

const mockUpdateObjective = jest.fn();
const mockEmptyArray: never[] = [];
const mockDimensions = {};
let mockWorkspace = "map-performance";
let mockViewport = { height: 1100, left: 0, top: 0, width: 4000 };

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspace }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { token: "session" } }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({ data: mockEmptyArray }),
}));
jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({ data: mockEmptyArray }),
}));
jest.mock("@/modules/objectives/hooks", () => ({
  useUpdateObjectiveMutation: () => ({ mutate: mockUpdateObjective }),
}));
jest.mock("@/modules/teams/public/client", () => ({
  useTeams: () => ({ data: mockEmptyArray }),
}));
jest.mock("@/modules/key-results/components/key-result-context-menu", () => ({
  KeyResultContextMenu: ({ children }: { children: ReactNode }) => children,
}));
jest.mock("@/modules/objectives/stories/overview/new-key-result", () => ({
  NewKeyResultDialog: () => null,
}));
jest.mock("@/modules/objectives/queries/get-key-results", () => ({
  getKeyResults: jest.fn(),
}));
jest.mock("./strategy-map-layout", () => {
  const actual = jest.requireActual<typeof StrategyMapLayoutModule>(
    "./strategy-map-layout",
  );
  return {
    ...actual,
    createConnectionPath: jest.fn(actual.createConnectionPath),
    createStrategyMapLayout: jest.fn(actual.createStrategyMapLayout),
    getStrategyConnections: jest.fn(actual.getStrategyConnections),
  };
});
jest.mock("./strategy-map-canvas-nodes", () => ({
  StrategyCanvasNodes: () => null,
}));
jest.mock("./use-strategy-map-canvas-viewport", () => ({
  useStrategyMapCanvasNodeDimensions: () => mockDimensions,
  useStrategyMapCanvasViewport: () => mockViewport,
}));

const objectives = Array.from({ length: 500 }, (_, index) => ({
  id: `objective-${index}`,
  keyResultCount: 2,
  stats: { completed: 0, total: 0 },
})) as Objective[];
const strategy: StrategyMap = {
  description: null,
  pillars: [
    {
      description: null,
      id: "pillar",
      name: "Scale",
      objectiveIds: objectives.map(({ id }) => id),
      orderIndex: 0,
    },
  ],
  ultimateGoal: "Scale strategy execution",
};
const keyResultsFor = (objectiveId: string) =>
  [1, 2].map((index) => ({
    id: `${objectiveId}-key-result-${index}`,
    objectiveId,
  })) as KeyResult[];
let queryClient: QueryClient;
const Wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

beforeEach(() => {
  window.localStorage.clear();
  mockWorkspace = "map-performance";
  mockViewport = { height: 1100, left: 0, top: 0, width: 4000 };
  queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  objectives.forEach(({ id }) => {
    queryClient.setQueryData(
      objectiveKeys.keyResults(mockWorkspace, id),
      keyResultsFor(id),
    );
  });
  jest.mocked(getKeyResults).mockReset().mockResolvedValue([]);
  jest.mocked(createStrategyMapLayout).mockClear();
  jest.mocked(createConnectionPath).mockClear();
  jest.mocked(getStrategyConnections).mockClear();
});

afterEach(() => {
  queryClient.clear();
  jest.restoreAllMocks();
});

it("keeps layout data stable across camera windows containing the same objectives", async () => {
  const activeObjectiveIds = new Set(["objective-0", "objective-1"]);
  const { result, rerender } = renderHook(
    ({ activeIds }) => useStrategyKeyResults(objectives, activeIds),
    { initialProps: { activeIds: activeObjectiveIds }, wrapper: Wrapper },
  );
  const original = result.current;

  for (let frame = 0; frame < 30; frame++) {
    rerender({ activeIds: new Set(activeObjectiveIds) });
  }
  expect(result.current.keyResultsByObjective).toBe(
    original.keyResultsByObjective,
  );
  expect(result.current.loadedObjectiveIds).toBe(original.loadedObjectiveIds);
  expect(getKeyResults).not.toHaveBeenCalled();

  act(() => {
    queryClient.setQueryData(
      objectiveKeys.keyResults(mockWorkspace, "objective-0"),
      [{ ...keyResultsFor("objective-0")[0], name: "Updated result" }],
    );
  });
  await waitFor(() => {
    expect(
      result.current.keyResultsByObjective.get("objective-0"),
    ).toHaveLength(1);
  });
  expect(result.current.keyResultsByObjective).not.toBe(
    original.keyResultsByObjective,
  );

  rerender({ activeIds: new Set(["objective-2"]) });
  expect([...result.current.keyResultsByObjective.keys()]).toEqual([
    "objective-0",
    "objective-1",
    "objective-2",
  ]);
  expect([...result.current.loadedObjectiveIds]).toEqual([
    "objective-0",
    "objective-1",
    "objective-2",
  ]);

  const afterEnteringWindow = result.current.keyResultsByObjective;
  rerender({ activeIds: activeObjectiveIds });
  expect(result.current.keyResultsByObjective).toBe(afterEnteringWindow);
  expect(getKeyResults).not.toHaveBeenCalled();
});

it("retains offscreen graph data without fetching inactive objectives and prunes deleted owners", async () => {
  const { result, rerender } = renderHook(
    ({ activeIds, currentObjectives }) =>
      useStrategyKeyResults(currentObjectives, activeIds),
    {
      initialProps: {
        activeIds: new Set(["objective-0"]),
        currentObjectives: objectives,
      },
      wrapper: Wrapper,
    },
  );
  rerender({
    activeIds: new Set(["objective-1"]),
    currentObjectives: objectives,
  });

  await act(async () => {
    await queryClient.invalidateQueries({
      queryKey: objectiveKeys.keyResults(mockWorkspace, "objective-0"),
    });
  });
  expect(getKeyResults).not.toHaveBeenCalled();
  expect(result.current.keyResultsByObjective.has("objective-0")).toBe(true);
  expect(result.current.keyResultsByObjective.has("objective-2")).toBe(false);

  act(() => {
    queryClient.setQueryData(
      objectiveKeys.keyResults(mockWorkspace, "objective-0"),
      [{ ...keyResultsFor("objective-0")[0], name: "Updated while offscreen" }],
    );
  });
  await waitFor(() => {
    expect(result.current.keyResultsByObjective.get("objective-0")).toEqual([
      expect.objectContaining({ name: "Updated while offscreen" }),
    ]);
  });

  rerender({
    activeIds: new Set(["objective-1"]),
    currentObjectives: objectives.filter(({ id }) => id !== "objective-0"),
  });
  expect(result.current.keyResultsByObjective.has("objective-0")).toBe(false);
});

it("keeps loaded key-result connector geometry when revisiting a previous window", () => {
  const { result, rerender } = renderHook(
    ({ activeIds }) => useStrategyKeyResults(objectives, activeIds),
    {
      initialProps: { activeIds: new Set(["objective-0"]) },
      wrapper: Wrapper,
    },
  );
  const firstResult = result.current.keyResultsByObjective.get("objective-0");
  rerender({ activeIds: new Set(["objective-24"]) });
  const retained = result.current.keyResultsByObjective;
  const layout = createStrategyMapLayout(
    strategy,
    objectives,
    retained,
    new Set(["objective-0", "objective-24"]),
  );
  const connections = getStrategyConnections(
    strategy,
    new Set(objectives.map(({ id }) => id)),
    retained,
    new Set(["objective-0", "objective-24"]),
  );

  expect(retained.get("objective-0")).toBe(firstResult);
  expect(layout.positions["key-result:objective-0-key-result-1"]).toBeDefined();
  expect(
    connections.some(
      ({ targetId }) => targetId === "key-result:objective-0-key-result-1",
    ),
  ).toBe(true);
  rerender({ activeIds: new Set(["objective-0"]) });
  expect(result.current.keyResultsByObjective).toBe(retained);
});

it("refreshes invalidated retained data when its objective becomes active again", async () => {
  const { result, rerender } = renderHook(
    ({ activeIds }) => useStrategyKeyResults(objectives, activeIds),
    {
      initialProps: { activeIds: new Set(["objective-0"]) },
      wrapper: Wrapper,
    },
  );
  rerender({ activeIds: new Set(["objective-1"]) });
  await act(async () => {
    await queryClient.invalidateQueries({
      queryKey: objectiveKeys.keyResults(mockWorkspace, "objective-0"),
    });
  });
  expect(getKeyResults).not.toHaveBeenCalled();

  jest
    .mocked(getKeyResults)
    .mockResolvedValue([
      { ...keyResultsFor("objective-0")[0], name: "Refreshed" },
    ]);
  rerender({ activeIds: new Set(["objective-0"]) });
  await waitFor(() => {
    expect(result.current.keyResultsByObjective.get("objective-0")).toEqual([
      expect.objectContaining({ name: "Refreshed" }),
    ]);
  });
  expect(getKeyResults).toHaveBeenCalledTimes(1);
});

it("does not retain key results when moving into an uncached workspace", async () => {
  const activeIds = new Set(["objective-0"]);
  const { result, rerender } = renderHook(
    () => useStrategyKeyResults(objectives, activeIds),
    { wrapper: Wrapper },
  );
  expect(result.current.loadedObjectiveIds.has("objective-0")).toBe(true);

  mockWorkspace = "another-workspace";
  rerender();
  expect(result.current.keyResultsByObjective.get("objective-0")).toEqual([]);
  expect(result.current.loadedObjectiveIds.size).toBe(0);
  await waitFor(() => {
    expect(result.current.loadedObjectiveIds.has("objective-0")).toBe(true);
  });
  expect(getKeyResults).toHaveBeenCalledWith("objective-0", {
    session: { token: "session" },
    workspaceSlug: "another-workspace",
  });
});

it("does no layout, edge geometry, or persistence work during thirty camera-only updates of a 500-objective map", () => {
  const onAction = jest.fn();
  const renderMap = () => (
    <StrategyMapCanvas
      canEdit
      objectives={objectives}
      onAddPillar={onAction}
      onAlign={onAction}
      onDeletePillar={onAction}
      onSelectGoal={onAction}
      onSelectKeyResult={onAction}
      onSelectObjective={onAction}
      onSelectPillar={onAction}
      onZoomChange={onAction}
      showUnaligned
      strategy={strategy}
      zoom={1}
    />
  );
  const { rerender } = render(renderMap(), { wrapper: Wrapper });
  expect(createStrategyMapLayout).toHaveBeenCalled();
  expect(createConnectionPath).toHaveBeenCalled();
  const persist = jest.spyOn(Storage.prototype, "setItem");
  jest.mocked(createStrategyMapLayout).mockClear();
  jest.mocked(createConnectionPath).mockClear();
  jest.mocked(getStrategyConnections).mockClear();

  for (let frame = 1; frame <= 30; frame++) {
    mockViewport = { ...mockViewport, left: frame, top: frame };
    rerender(renderMap());
  }

  expect(createStrategyMapLayout).not.toHaveBeenCalled();
  expect(getStrategyConnections).not.toHaveBeenCalled();
  expect(createConnectionPath).not.toHaveBeenCalled();
  expect(persist).not.toHaveBeenCalled();
  expect(getKeyResults).not.toHaveBeenCalled();

  // Entering a different objective window still rebuilds from the new data.
  mockViewport = { height: 500, left: 0, top: 3650, width: 4000 };
  rerender(renderMap());
  expect(createStrategyMapLayout).toHaveBeenCalled();
  expect(createConnectionPath).toHaveBeenCalled();

  jest.mocked(createStrategyMapLayout).mockClear();
  jest.mocked(createConnectionPath).mockClear();
  persist.mockClear();
  mockViewport = { height: 1100, left: 0, top: 0, width: 4000 };
  rerender(renderMap());
  expect(createStrategyMapLayout).not.toHaveBeenCalled();
  expect(createConnectionPath).not.toHaveBeenCalled();
  expect(persist).not.toHaveBeenCalled();
});

it("updates connector geometry when nodes move while retaining crossing and pinned edges", () => {
  const singleObjectiveStrategy = {
    ...strategy,
    pillars: [{ ...strategy.pillars[0], objectiveIds: ["objective-0"] }],
  };
  const props = {
    alwaysVisibleNodeIds: new Set<string>(),
    dimensions: {},
    expandedObjectiveIds: new Set<string>(),
    keyResultsByObjective: new Map<string, KeyResult[]>(),
    objectiveIds: new Set(["objective-0"]),
    positions: {
      goal: { x: 200, y: -500 },
      "pillar:pillar": { x: 200, y: -100 },
      "objective:objective-0": { x: 200, y: 800 },
    },
    strategy: singleObjectiveStrategy,
    viewport: { height: 200, left: 200, top: 300, width: 400 },
  };
  const { container, rerender } = render(<CanvasConnections {...props} />);
  expect(container.querySelectorAll("path")).toHaveLength(1);
  const originalPath = container.querySelector("path")?.getAttribute("d");
  const moved = {
    ...props,
    positions: {
      ...props.positions,
      "objective:objective-0": { x: 350, y: 900 },
    },
  };
  rerender(<CanvasConnections {...moved} />);
  expect(container.querySelector("path")?.getAttribute("d")).not.toBe(
    originalPath,
  );

  const offscreen = { ...moved, viewport: { ...props.viewport, left: 10000 } };
  rerender(<CanvasConnections {...offscreen} />);
  expect(container.querySelectorAll("path")).toHaveLength(0);
  rerender(
    <CanvasConnections
      {...offscreen}
      alwaysVisibleNodeIds={new Set(["objective:objective-0"])}
    />,
  );
  expect(container.querySelectorAll("path")).toHaveLength(1);
});
