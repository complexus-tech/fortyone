/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import { StrategyMapCanvas } from "./strategy-map-canvas";
import type { StrategyMap } from "./types";

const mockUseStrategyKeyResults = jest.fn();
let mockViewport = {
  height: 1100,
  left: 0,
  top: 0,
  width: 4000,
};

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "scale-test" }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({ data: [] }),
}));
jest.mock("@/modules/objectives/hooks", () => ({
  useUpdateObjectiveMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("@/modules/key-results/components/key-result-context-menu", () => ({
  KeyResultContextMenu: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}));
jest.mock("@/modules/teams/public/client", () => ({
  useTeams: () => ({ data: [] }),
}));
jest.mock("./strategy-map-canvas-renderers", () => ({
  CanvasConnections: () => null,
  getDefaultNodeDimensions: (nodeId: string) =>
    nodeId.startsWith("key-result:")
      ? { height: 154, width: 280 }
      : { height: 174, width: 340 },
  getNodePosition: (
    positions: Record<string, { x: number; y: number }>,
    nodeId: string,
  ) => positions[nodeId],
  isInteractiveTarget: () => false,
}));
jest.mock("./strategy-map-canvas-nodes", () => ({
  StrategyCanvasNodes: ({
    keyResultNodes,
    objectives,
  }: {
    keyResultNodes: readonly { keyResult: KeyResult }[];
    objectives: readonly Objective[];
  }) => (
    <div>
      {objectives.map((objective) => (
        <div
          data-testid={`mounted-objective-${objective.id}`}
          key={objective.id}
        />
      ))}
      {keyResultNodes.map(({ keyResult }) => (
        <div data-testid="mounted-key-result" key={keyResult.id} />
      ))}
    </div>
  ),
}));
jest.mock("./use-strategy-key-results", () => ({
  useStrategyKeyResults: (...args: unknown[]) =>
    mockUseStrategyKeyResults(...args),
}));
jest.mock("./use-strategy-map-canvas-viewport", () => ({
  useStrategyMapCanvasNodeDimensions: () => ({}),
  useStrategyMapCanvasViewport: () => mockViewport,
}));

const createObjective = (index: number) =>
  ({
    id: `objective-${index}`,
    keyResultCount: 2,
    stats: { completed: 0, total: 0 },
  }) as Objective;

const createKeyResult = (objectiveId: string, index: number) =>
  ({
    id: `${objectiveId}-key-result-${index}`,
    objectiveId,
  }) as KeyResult;

const objectives = Array.from({ length: 50 }, (_, index) =>
  createObjective(index),
);
const strategy: StrategyMap = {
  description: null,
  pillars: [
    {
      description: null,
      id: "pillar-1",
      name: "Scale",
      objectiveIds: objectives.map((objective) => objective.id),
      orderIndex: 0,
    },
  ],
  ultimateGoal: "Scale strategy execution",
};

describe("StrategyMapCanvas scale window", () => {
  it("keeps mounted cards and key-result requests bounded for fifty objectives", () => {
    mockViewport = { height: 1100, left: 0, top: 0, width: 4000 };
    mockUseStrategyKeyResults.mockImplementation(
      (
        _objectives: readonly Objective[],
        activeObjectiveIds: ReadonlySet<string>,
      ) => {
        const keyResultsByObjective = new Map(
          Array.from(activeObjectiveIds, (objectiveId) => [
            objectiveId,
            [createKeyResult(objectiveId, 1), createKeyResult(objectiveId, 2)],
          ]),
        );

        return {
          isPending: false,
          keyResultsByObjective,
          loadedObjectiveIds: new Set(activeObjectiveIds),
        };
      },
    );

    const renderMap = () => (
      <StrategyMapCanvas
        canEdit
        objectives={objectives}
        onAddPillar={jest.fn()}
        onAlign={jest.fn()}
        onDeletePillar={jest.fn()}
        onSelectGoal={jest.fn()}
        onSelectKeyResult={jest.fn()}
        onSelectObjective={jest.fn()}
        onSelectPillar={jest.fn()}
        onZoomChange={jest.fn()}
        showUnaligned
        strategy={strategy}
        zoom={1}
      />
    );
    const { rerender } = render(renderMap());

    expect(screen.getAllByTestId(/^mounted-objective-/)).toHaveLength(4);
    expect(screen.getAllByTestId("mounted-key-result")).toHaveLength(8);

    mockViewport = { height: 500, left: 0, top: 3650, width: 4000 };
    rerender(renderMap());

    const requestedKeyResultCounts = mockUseStrategyKeyResults.mock.calls.map(
      ([, activeObjectiveIds]) =>
        (activeObjectiveIds as ReadonlySet<string>).size,
    );

    expect(Math.max(...requestedKeyResultCounts)).toBe(4);
    expect(screen.getAllByTestId(/^mounted-objective-/)).toHaveLength(4);
    expect(screen.getAllByTestId("mounted-key-result")).toHaveLength(8);
    expect(
      mockUseStrategyKeyResults.mock.calls.some(([, activeObjectiveIds]) =>
        (activeObjectiveIds as ReadonlySet<string>).has("objective-24"),
      ),
    ).toBe(true);
    expect(
      screen.getByTestId("mounted-objective-objective-24"),
    ).toBeInTheDocument();
    expect(
      screen.queryByTestId("mounted-objective-objective-0"),
    ).not.toBeInTheDocument();
  });
});
