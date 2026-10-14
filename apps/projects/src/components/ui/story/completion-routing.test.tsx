/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { Story } from "@/modules/stories/types";
import type { StoryUpdate } from "@/modules/story/types";
import { StoryCard } from "./card";
import { StoryRow } from "./row";

const mockMutate = jest.fn<
  undefined,
  [{ storyId: string; payload: StoryUpdate }]
>();
const mockUpdateStory = jest.fn<undefined, [string, StoryUpdate]>();
const mockStatuses = [
  { id: "started", name: "In progress", category: "started" },
  { id: "unstarted", name: "Todo", category: "unstarted" },
  { id: "backlog", name: "Backlog", category: "backlog" },
  { id: "completed", name: "Completed", category: "completed" },
  { id: "cancelled", name: "Cancelled", category: "cancelled" },
  { id: "paused", name: "Paused", category: "paused" },
];

jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactModule>("react");
  const Primitive = ({ children }: { children?: ReactModule.ReactNode }) =>
    React.createElement("div", null, children);
  const Button = React.forwardRef<
    HTMLButtonElement,
    ReactModule.ButtonHTMLAttributes<HTMLButtonElement>
  >(({ children, disabled, onClick }, ref) =>
    React.createElement(
      "button",
      { disabled, onClick, ref, type: "button" },
      children,
    ),
  );
  Button.displayName = "Button";
  const Dialog = Object.assign(
    ({
      children,
      open,
    }: {
      children?: ReactModule.ReactNode;
      open: boolean;
    }) =>
      open ? React.createElement("div", { role: "dialog" }, children) : null,
    {
      Body: Primitive,
      Content: Primitive,
      Description: Primitive,
      Footer: Primitive,
      Header: Primitive,
      Title: Primitive,
    },
  );

  return {
    Avatar: Primitive,
    Badge: Primitive,
    Box: Primitive,
    Button,
    Checkbox: () => null,
    DatePicker: Primitive,
    Dialog,
    Divider: Primitive,
    Flex: Primitive,
    Input: () => null,
    Text: Primitive,
    Tooltip: Primitive,
  };
});
jest.mock("icons", () => ({
  ArrowRight2Icon: () => null,
  CalendarIcon: () => null,
  EstimateIcon: () => null,
  SprintsIcon: () => null,
  StoryIcon: () => null,
  SubStoryIcon: () => null,
  Time02Icon: () => null,
}));
jest.mock("lib", () => ({ cn: () => "" }));
jest.mock("@/utils", () => ({ hexToRgba: () => "transparent" }));
jest.mock("next/link", () => ({
  __esModule: true,
  default: ({ children }: { children?: ReactModule.ReactNode }) => children,
}));
jest.mock("next/navigation", () => ({ useRouter: () => ({}) }));
jest.mock("@dnd-kit/core", () => ({ useDraggable: () => ({}) }));
jest.mock("@tanstack/react-query", () => ({ useQueryClient: () => ({}) }));
jest.mock("@/lib/auth/client", () => ({ useSession: () => ({ data: null }) }));
jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: () => ({ data: null }),
}));
jest.mock("@/lib/hooks/figma", () => ({
  useFigmaHandoffStatuses: () => ({ data: null }),
}));
jest.mock("@/modules/settings/workspace/integrations/figma/icon", () => ({
  FigmaIcon: () => null,
}));
jest.mock("@/modules/story/queries/get-story", () => ({ getStory: jest.fn() }));
jest.mock("@/modules/story/queries/get-attachments", () => ({
  getStoryAttachments: jest.fn(),
}));
jest.mock("@/lib/queries/links/get-links", () => ({ getLinks: jest.fn() }));
jest.mock("@/shared/story/board-property-slots", () => ({
  useBoardPropertySlots: () => ({ Badges: () => null }),
}));
jest.mock("@/components/ui/board-context", () => ({
  useBoard: () => ({
    embedded: true,
    isColumnVisible: (column: string) => column === "Status",
    selectedStories: [],
    updateStory: mockUpdateStory,
  }),
  useBoardView: () => ({
    embedded: true,
    isColumnVisible: (column: string) => column === "Status",
    selectedStories: [],
    updateStory: mockUpdateStory,
  }),
}));
jest.mock("@/modules/story/hooks/update-mutation", () => ({
  useUpdateStoryMutation: () => ({ mutate: mockMutate }),
}));
jest.mock("@/hooks", () => ({
  useHydratedNow: () => null,
  useMediaQuery: () => true,
  useTerminology: () => ({ getTermDisplay: () => "story" }),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => path,
    workspaceSlug: "workspace",
  }),
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => ({ data: mockStatuses }),
}));
jest.mock("./statuses-menu", () => {
  const React = jest.requireActual<typeof ReactModule>("react");
  const Primitive = ({ children }: { children?: ReactModule.ReactNode }) =>
    children;
  return {
    StatusesMenu: Object.assign(Primitive, {
      Trigger: Primitive,
      Items: ({ setStatusId }: { setStatusId: (statusId: string) => void }) =>
        React.createElement(
          "button",
          {
            onClick: () => {
              setStatusId("completed");
            },
            type: "button",
          },
          "Select completed status",
        ),
    }),
  };
});
jest.mock("./context-menu", () => ({
  StoryContextMenu: ({ children }: { children?: ReactModule.ReactNode }) =>
    children,
}));
jest.mock("../row-wrapper", () => ({
  RowWrapper: ({ children }: { children?: ReactModule.ReactNode }) => children,
}));
jest.mock("../member-tooltip", () => ({ MemberTooltip: () => null }));
jest.mock("./drag-handle", () => ({ DragHandle: () => null }));
jest.mock("./people-avatars", () => ({ PeopleAvatars: () => null }));
jest.mock("./assignees-menu", () => ({ AssigneesMenu: () => null }));
jest.mock("./sprints-menu", () => ({ SprintsMenu: () => null }));
jest.mock("./estimate-menu", () => ({ EstimateMenu: () => null }));
jest.mock("./time-needed-menu", () => ({ TimeNeededMenu: () => null }));
jest.mock("./labels", () => ({ Labels: () => null }));
jest.mock("./priorities-menu", () => ({ PrioritiesMenu: () => null }));
jest.mock("../story-status-icon", () => ({ StoryStatusIcon: () => null }));
jest.mock("../priority-icon", () => ({ PriorityIcon: () => null }));
jest.mock("./strategy-properties", () => ({
  StoryStrategyProperties: () => null,
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
  createdAt: "2026-10-13T08:00:00.000Z",
  deletedAt: null,
  endDate: null,
  epicId: null,
  estimateLabel: null,
  estimateScheme: "points",
  estimateValue: null,
  estimatedDurationMinutes: null,
  id: "parent-story",
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
  statusId: "started",
  subStories: [],
  team: null,
  teamId: "team-1",
  title: "Parent story",
  updatedAt: "2026-10-13T08:00:00.000Z",
  workspaceId: "workspace-1",
};

describe.each(["list", "kanban"] as const)(
  "%s story completion routing",
  (view) => {
    beforeEach(() => {
      mockMutate.mockClear();
      mockUpdateStory.mockClear();
    });

    const renderStory = (childStatuses: string[]) => {
      const props = {
        handleStoryClick: jest.fn(),
        story: {
          ...story,
          subStories: childStatuses.map((statusId, index) => ({
            ...story,
            id: `child-${statusId}`,
            sequenceId: index + 2,
            statusId,
          })),
        },
      };
      render(
        view === "list" ? <StoryRow {...props} /> : <StoryCard {...props} />,
      );
    };

    const getUpdates = () =>
      view === "list"
        ? mockMutate.mock.calls.map(([update]) => update)
        : mockUpdateStory.mock.calls.map(([storyId, payload]) => ({
            storyId,
            payload,
          }));

    it("confirms completion for the parent and each active child using status-only payloads", () => {
      renderStory([
        "started",
        "unstarted",
        "backlog",
        "completed",
        "cancelled",
        "paused",
      ]);

      fireEvent.click(
        screen.getByRole("button", { name: "Select completed status" }),
      );

      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(getUpdates()).toEqual([]);

      fireEvent.click(
        screen.getByRole("button", { name: "Yes, mark as done" }),
      );

      expect(getUpdates()).toEqual(
        [
          "parent-story",
          "child-started",
          "child-unstarted",
          "child-backlog",
        ].map((storyId) => ({ storyId, payload: { statusId: "completed" } })),
      );
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("updates only the parent when child completion is declined", () => {
      renderStory(["started", "unstarted", "backlog"]);

      fireEvent.click(
        screen.getByRole("button", { name: "Select completed status" }),
      );
      fireEvent.click(screen.getByRole("button", { name: "No, leave as is" }));

      expect(getUpdates()).toEqual([
        { storyId: "parent-story", payload: { statusId: "completed" } },
      ]);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });

    it("updates the parent directly when no children need completion", () => {
      renderStory(["completed", "cancelled", "paused"]);

      fireEvent.click(
        screen.getByRole("button", { name: "Select completed status" }),
      );

      expect(getUpdates()).toEqual([
        { storyId: "parent-story", payload: { statusId: "completed" } },
      ]);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  },
);
