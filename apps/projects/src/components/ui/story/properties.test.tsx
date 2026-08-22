/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import { render } from "@testing-library/react";
import type { Story } from "@/modules/stories/types";
import { StoryProperties } from "./properties";

const mockIsColumnVisible = jest.fn();
const mockStoryStrategyProperties = jest.fn((_props: unknown) => null);

jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactModule>("react");
  const Primitive = ({ children }: { children?: ReactModule.ReactNode }) =>
    React.createElement("div", null, children);

  return {
    Box: Primitive,
    Button: Primitive,
    DatePicker: Primitive,
    Divider: Primitive,
    Flex: Primitive,
    Text: Primitive,
    Tooltip: Primitive,
  };
});
jest.mock("icons", () => ({
  ArrowRight2Icon: () => null,
  CalendarIcon: () => null,
  EstimateIcon: () => null,
  SprintsIcon: () => null,
  SubStoryIcon: () => null,
  Time02Icon: () => null,
}));
jest.mock("lib", () => ({ cn: () => "" }));
jest.mock("@/utils", () => ({ hexToRgba: () => "transparent" }));
jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children }: { children?: ReactModule.ReactNode }) => children,
}));
jest.mock("@/components/ui/board-context", () => ({
  useBoard: () => ({ isColumnVisible: mockIsColumnVisible }),
}));

jest.mock("@/components/ui/confirm-dialog", () => ({
  ConfirmDialog: () => null,
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "story",
  }),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => path,
    workspaceSlug: "first",
  }),
}));

jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => ({ data: [] }),
}));
jest.mock("@/components/ui/story/sprints-menu", () => ({
  SprintsMenu: () => null,
}));
jest.mock("@/components/ui/story/estimate-menu", () => ({
  EstimateMenu: () => null,
}));
jest.mock("@/components/ui/story/time-needed-menu", () => ({
  TimeNeededMenu: () => null,
}));
jest.mock("@/components/ui/story/labels", () => ({ Labels: () => null }));
jest.mock("@/components/ui/story/statuses-menu", () => ({
  StatusesMenu: () => null,
}));
jest.mock("@/components/ui/story/priorities-menu", () => ({
  PrioritiesMenu: () => null,
}));
jest.mock("@/components/ui/story-status-icon", () => ({
  StoryStatusIcon: () => null,
}));
jest.mock("@/components/ui/priority-icon", () => ({
  PriorityIcon: () => null,
}));

jest.mock("./strategy-properties", () => ({
  StoryStrategyProperties: (props: unknown) =>
    mockStoryStrategyProperties(props),
}));

const story: Story = {
  archivedAt: null,
  assignee: null,
  assigneeId: null,
  autoSchedulingEnabled: false,
  autoSchedulingLocked: false,
  autoSchedulingReason: null,
  autoSchedulingStatus: "off",
  autoSchedulingUpdatedAt: null,
  collaboratorCount: 0,
  completedAt: null,
  createdAt: "2026-08-22T08:00:00.000Z",
  deletedAt: null,
  endDate: null,
  epicId: null,
  estimateLabel: null,
  estimateScheme: "points",
  estimateValue: null,
  estimatedDurationMinutes: null,
  id: "story-1",
  keyResultId: null,
  labels: null,
  minimumFocusBlockMinutes: null,
  objective: null,
  objectiveId: null,
  priority: "No Priority",
  reporter: null,
  reporterId: "member-1",
  sequenceId: 1,
  sprint: null,
  sprintId: null,
  startDate: null,
  statusId: "status-1",
  subStories: [],
  team: null,
  teamId: "team-1",
  title: "Story",
  updatedAt: "2026-08-22T08:00:00.000Z",
  workspaceId: "workspace-1",
};

const renderProperties = (overrides: Partial<Story> = {}) =>
  render(
    <StoryProperties {...story} {...overrides} handleUpdate={jest.fn()} />,
  );

describe("StoryProperties strategy metadata", () => {
  const visibleColumns = new Set<string>();

  beforeEach(() => {
    visibleColumns.clear();
    mockIsColumnVisible.mockImplementation((column: string) =>
      visibleColumns.has(column),
    );
    mockStoryStrategyProperties.mockClear();
  });

  it("does not mount strategy properties without strategy metadata", () => {
    visibleColumns.add("Objective");
    visibleColumns.add("Key Result");

    renderProperties();

    expect(mockStoryStrategyProperties).not.toHaveBeenCalled();
  });

  it("mounts strategy properties for a visible objective", () => {
    visibleColumns.add("Objective");

    renderProperties({ objectiveId: "objective-1" });

    expect(mockStoryStrategyProperties).toHaveBeenCalledTimes(1);
  });

  it("mounts strategy properties for a visible key result", () => {
    visibleColumns.add("Key Result");

    renderProperties({
      keyResultId: "key-result-1",
      objectiveId: "objective-1",
    });

    expect(mockStoryStrategyProperties).toHaveBeenCalledTimes(1);
  });

  it("does not mount hidden strategy metadata", () => {
    renderProperties({
      keyResultId: "key-result-1",
      objectiveId: "objective-1",
    });

    expect(mockStoryStrategyProperties).not.toHaveBeenCalled();
  });
});
