/* global afterAll, beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { RoadmapGanttBoard } from "@/components/ui/roadmap-gantt-board";
import type { KeyResult, Objective } from "@/modules/objectives/types";
import { DEFAULT_OBJECTIVE_VIEW_OPTIONS } from "../objective-board-utils";
import { ObjectivesBoard } from "./objectives-board";

const mockMutate = jest.fn();
const mockUseKeyResults = jest.fn();
const mockUseTeamMembers = jest.fn((_teamId: string) => ({ data: [] }));
const mockKeyResultsByObjective = new Map<string, KeyResult[]>();

jest.mock("@dnd-kit/core", () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children;

  return {
    DndContext: Passthrough,
    DragOverlay: Passthrough,
    PointerSensor: jest.fn(),
    useDraggable: () => ({
      attributes: {},
      isDragging: false,
      listeners: {},
      setNodeRef: jest.fn(),
    }),
    useDroppable: () => ({ isOver: false, setNodeRef: jest.fn() }),
    useSensor: () => ({}),
    useSensors: () => [],
  };
});

jest.mock("ui", () => {
  const actual = jest.requireActual("ui");
  const Passthrough = ({ children }: { children?: ReactNode }) => children;
  const Empty = () => null;
  const DatePicker = Object.assign(Passthrough, {
    Calendar: Empty,
    Trigger: Passthrough,
  });
  const Menu = Object.assign(Passthrough, {
    Button: Passthrough,
    Group: Passthrough,
    Item: Passthrough,
    Items: Passthrough,
  });
  const Popover = Object.assign(Passthrough, {
    Content: Passthrough,
    Trigger: Passthrough,
  });

  return {
    ...actual,
    DatePicker,
    Menu,
    Popover,
    Tooltip: Passthrough,
  };
});

jest.mock("@/components/ui", () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children;
  const Empty = () => null;
  const Compound = Object.assign(Passthrough, {
    Items: Empty,
    Trigger: Passthrough,
  });

  return {
    AssigneesMenu: Compound,
    ObjectiveHealthIcon: Empty,
    PrioritiesMenu: Compound,
    PriorityIcon: Empty,
  };
});

jest.mock("@/components/ui/story/assignees-menu", () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children;
  const AssigneesMenu = Object.assign(Passthrough, {
    Items: () => null,
    Trigger: Passthrough,
  });
  return { AssigneesMenu };
});

jest.mock("@/components/ui/story/priorities-menu", () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children;
  const PrioritiesMenu = Object.assign(Passthrough, {
    Items: () => null,
    Trigger: Passthrough,
  });
  return { PrioritiesMenu };
});

jest.mock("@/components/ui/objective-statuses-menu", () => {
  const Passthrough = ({ children }: { children?: ReactNode }) => children;
  const ObjectiveStatusesMenu = Object.assign(Passthrough, {
    Items: () => null,
    Trigger: Passthrough,
  });
  return { ObjectiveStatusesMenu };
});

jest.mock("@/components/ui/objective-status-icon", () => ({
  ObjectiveStatusIcon: () => null,
}));

jest.mock("@/modules/objectives/components/objective-health-editor", () => ({
  ObjectiveHealthEditor: ({ children }: { children?: ReactNode }) => children,
}));

jest.mock("@/modules/objectives/components/objective-forecast-risk", () => ({
  ObjectiveForecastRiskBadge: ({
    label,
    size,
  }: {
    label?: string;
    size?: string;
  }) => <span data-forecast-label={label} data-forecast-size={size} />,
}));

jest.mock("@/modules/objectives/hooks", () => ({
  useCanUpdateObjective: () => true,
  useUpdateObjectiveMutation: () => ({ mutate: mockMutate }),
}));

jest.mock("@/modules/objectives/hooks/update-mutation", () => ({
  useUpdateObjectiveMutation: () => ({ mutate: mockMutate }),
}));

jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({ data: [] }),
}));

jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({
    data: [
      {
        category: "started",
        color: "#4A90E2",
        id: "status-1",
        isDefault: true,
        name: "In progress",
        orderIndex: 0,
      },
    ],
  }),
}));

jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: (teamId: string) => mockUseTeamMembers(teamId),
}));

jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({ data: [{ code: "TEAM", id: "team-1" }] }),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: null }),
}));

jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ prefetchQuery: jest.fn() }),
}));

jest.mock("@/hooks", () => {
  const React = jest.requireActual("react");
  return {
    useLocalStorage: <T,>(_key: string, initialValue: T) =>
      React.useState(initialValue),
    useTerminology: () => ({
      getTermDisplay: (term: string) => term,
    }),
    useWorkspacePath: () => ({ workspaceSlug: "workspace" }),
  };
});

jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "admin" }),
}));

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: (term: string) => term,
  }),
}));

jest.mock("./roadmap-key-results", () => ({
  RoadmapKeyResultSummary: ({ objective }: { objective: Objective }) => {
    const { data = [] } = mockUseKeyResults(
      objective.id,
      objective.keyResultCount > 0,
    ) as { data?: KeyResult[] };
    return <span data-key-result-count={data.length} />;
  },
  RoadmapObjectiveListItem: ({
    objective,
    onObjectiveSelect,
    selected,
  }: {
    objective: Objective;
    onObjectiveSelect: () => void;
    selected: boolean;
  }) => {
    const { data = [] } = mockUseKeyResults(
      objective.id,
      objective.keyResultCount > 0,
    ) as { data?: KeyResult[] };
    return (
      <button
        aria-pressed={selected}
        data-key-result-count={data.length}
        data-roadmap-list-objective={objective.id}
        onClick={onObjectiveSelect}
        type="button"
      >
        {objective.name}
      </button>
    );
  },
}));

const createKeyResult = (objectiveId: string, index: number): KeyResult => ({
  contributors: [],
  createdAt: "2026-01-01T00:00:00Z",
  createdBy: "member-1",
  currentValue: index * 10,
  endDate: "2026-12-31",
  id: `${objectiveId}-key-result-${index}`,
  lead: null,
  measurementType: "percentage",
  name: `Key result ${index}`,
  objectiveId,
  sequenceId: index,
  startDate: "2026-01-01",
  startValue: 0,
  targetValue: 100,
  updatedAt: "2026-01-01T00:00:00Z",
});

const objectives = Array.from({ length: 50 }, (_, index): Objective => {
  const id = `objective-${index}`;
  const hasKeyResults = index % 2 === 0;
  if (hasKeyResults) {
    mockKeyResultsByObjective.set(id, [
      createKeyResult(id, 1),
      createKeyResult(id, 2),
    ]);
  }

  return {
    color: "#4A90E2",
    createdAt: "2026-01-01T00:00:00Z",
    createdBy: "member-1",
    description: "",
    endDate: "2026-12-31",
    forecastCauseStory: null,
    forecastDaysDelta: 0,
    forecastEndDate: null,
    forecastStartDate: null,
    health: "On Track",
    id,
    isPrivate: false,
    keyResultCount: hasKeyResults ? 2 : 0,
    leadUser: "member-1",
    name: `Objective ${index + 1}`,
    priority: "High",
    scheduleStatus: "on_track",
    sequenceId: index + 1,
    shortSummary: null,
    startDate: "2026-01-01",
    statusId: "status-1",
    teamId: "team-1",
    updatedAt: "2026-01-01T00:00:00Z",
    workspaceId: "workspace-1",
  };
});

const middleObjective = objectives[24];
const viewOptions = {
  ...DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  showEmptyGroups: false,
};
const MAX_GANTT_MOUNTED_OBJECTIVES = 10;
const MAX_GANTT_HOOK_RENDER_CALLS = 30;
const VIEW_PERFORMANCE_BUDGETS = {
  kanban: { enabledKeyResultQueries: 8, virtualItems: 10 },
  list: { enabledKeyResultQueries: 8, virtualItems: 8 },
} as const;

const originalRequestAnimationFrame = globalThis.requestAnimationFrame;
const originalCancelAnimationFrame = globalThis.cancelAnimationFrame;

beforeAll(() => {
  globalThis.requestAnimationFrame = (callback: FrameRequestCallback) => {
    callback(0);
    return 1;
  };
  globalThis.cancelAnimationFrame = () => undefined;
});

afterAll(() => {
  globalThis.requestAnimationFrame = originalRequestAnimationFrame;
  globalThis.cancelAnimationFrame = originalCancelAnimationFrame;
});

beforeEach(() => {
  localStorage.clear();
  mockMutate.mockClear();
  mockUseKeyResults.mockReset();
  mockUseKeyResults.mockImplementation((objectiveId: string) => ({
    data: mockKeyResultsByObjective.get(objectiveId) ?? [],
  }));
  mockUseTeamMembers.mockClear();
});

describe("Roadmap large objective datasets", () => {
  it("keeps Gantt rows bounded and selects a pinned middle objective", () => {
    const onObjectiveSelect = jest.fn();
    const { container } = render(
      <RoadmapGanttBoard
        objectives={objectives}
        onObjectiveSelect={onObjectiveSelect}
        onZoomLevelChange={jest.fn()}
        selectedObjectiveId={middleObjective.id}
        zoomLevel="months"
      />,
    );

    const mountedIds = new Set(
      Array.from(container.querySelectorAll("[data-gantt-item-id]"), (node) =>
        node.getAttribute("data-gantt-item-id"),
      ),
    );
    expect(mountedIds.size).toBeLessThanOrEqual(MAX_GANTT_MOUNTED_OBJECTIVES);
    expect(mountedIds).toContain(middleObjective.id);
    expect(mockUseTeamMembers.mock.calls.length).toBeLessThanOrEqual(
      MAX_GANTT_HOOK_RENDER_CALLS,
    );

    const timelineRows = screen
      .getByRole("list", { name: "Timeline items" })
      .querySelectorAll("[data-gantt-item-id]");
    expect(timelineRows[0]).toHaveStyle({ height: "45.5px" });
    expect(timelineRows[1]).toHaveStyle({
      transform: "translateY(45.5px)",
    });

    const middleLabel = screen
      .getAllByText(middleObjective.name)
      .find((element) => element.closest("button"));
    expect(middleLabel).toBeDefined();
    fireEvent.click(middleLabel!.closest("button")!);
    expect(onObjectiveSelect).toHaveBeenCalledWith(middleObjective);
  });

  it("shows only the forecast icon and day delta when collapsed", () => {
    const atRiskObjective = {
      ...objectives[0],
      forecastDaysDelta: 15,
      forecastEndDate: "2027-01-15",
      scheduleStatus: "at_risk" as const,
    };

    render(
      <RoadmapGanttBoard
        objectives={[atRiskObjective]}
        onObjectiveSelect={jest.fn()}
        onZoomLevelChange={jest.fn()}
        zoomLevel="months"
      />,
    );

    fireEvent.click(
      screen.getByRole("button", { name: "Collapse objectives panel" }),
    );

    expect(
      document.querySelector(
        '[data-forecast-label="delta"][data-forecast-size="control"]',
      ),
    ).toBeInTheDocument();
  });

  it.each(["list", "kanban"] as const)(
    "keeps %s rows and key-result subscriptions bounded",
    (layout) => {
      const onObjectiveSelect = jest.fn();
      const { container } = render(
        <ObjectivesBoard
          layout={layout}
          objectives={objectives}
          onCreateObjective={jest.fn()}
          onKeyResultSelect={jest.fn()}
          onObjectiveSelect={onObjectiveSelect}
          selectedObjectiveId={middleObjective.id}
          setViewOptions={jest.fn()}
          viewOptions={viewOptions}
        />,
      );

      const mountedVirtualItems = container.querySelectorAll(
        "[data-virtual-index]",
      );
      const enabledKeyResultObjectiveIds = new Set(
        mockUseKeyResults.mock.calls.flatMap(([objectiveId, enabled]) =>
          enabled ? [objectiveId as string] : [],
        ),
      );
      expect(mountedVirtualItems.length).toBeLessThanOrEqual(
        VIEW_PERFORMANCE_BUDGETS[layout].virtualItems,
      );
      expect(enabledKeyResultObjectiveIds.size).toBeLessThanOrEqual(
        VIEW_PERFORMANCE_BUDGETS[layout].enabledKeyResultQueries,
      );
      expect(mockUseKeyResults).toHaveBeenCalledWith(middleObjective.id, true);

      const middleLabel = screen.getByText(middleObjective.name);
      fireEvent.click(middleLabel.closest("button")!);
      expect(onObjectiveSelect).toHaveBeenCalledWith(middleObjective);
    },
  );
});
