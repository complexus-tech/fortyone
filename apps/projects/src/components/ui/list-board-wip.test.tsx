import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import type {
  GroupedStoriesResponse,
  StoryGroup,
} from "@/modules/stories/types";
import type { State } from "@/types/states";
import {
  BoardPropertySlotsProvider,
  type WorkflowCountsProviderProps,
} from "@/shared/story/board-property-slots";
import type { StoriesViewOptions } from "./stories-view-options-button";
import type { StoriesHeader as StoriesHeaderComponent } from "./stories-header";
import { ListBoard } from "./list-board";

const mockWorkflowCounts = jest.fn();
const mockStatus: State = {
  id: "status-id",
  name: "In progress",
  color: "#F59E0B",
  category: "started",
  isDefault: false,
  orderIndex: 1,
  teamId: "team-id",
  workspaceId: "workspace-id",
  createdAt: "2026-10-03T00:00:00Z",
  updatedAt: "2026-10-03T00:00:00Z",
  wipLimit: 5,
  activeCount: 4,
};

jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team-id" }),
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => ({ data: [mockStatus] }),
  useStatuses: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/members", () => ({ useMembers: () => ({ data: [] }) }));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [] }),
}));
jest.mock("@/hooks", () => ({
  useUserRole: () => ({ userRole: "member" }),
  useTerminology: () => ({
    getTermDisplay: (_term: string, options?: { variant?: string }) =>
      options?.variant === "singular" ? "story" : "stories",
  }),
}));
jest.mock("./board-context", () => ({
  useBoard: () => ({
    newStoryDefaults: {},
    viewOptions: { showEmptyGroups: true },
    selectedStories: [],
    setSelectedStories: jest.fn(),
  }),
}));
jest.mock("./new-story-dialog", () => ({ NewStoryDialog: () => null }));
jest.mock("./story-status-icon", () => ({ StoryStatusIcon: () => null }));
jest.mock("./stories-group", () => {
  const { StoriesHeader } = jest.requireActual<{
    StoriesHeader: typeof StoriesHeaderComponent;
  }>("./stories-header");
  return {
    StoriesGroup: ({
      group,
      status,
    }: {
      group: StoryGroup;
      status?: State;
    }) => (
      <StoriesHeader
        group={group}
        groupBy="status"
        isCollapsed={false}
        setIsCollapsed={jest.fn()}
        status={status}
      />
    ),
  };
});

const passthrough = ({ children }: { children: ReactNode }) => <>{children}</>;
const WorkflowCounts = ({
  children,
  ...props
}: WorkflowCountsProviderProps) => {
  mockWorkflowCounts(props);
  return children(new Map(mockCounts));
};
let mockCounts: [string, number][] = [];
const viewOptions: StoriesViewOptions = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
  showEmptyGroups: true,
  showSubStories: false,
  displayColumns: ["ID", "Status"],
};
const groupedStories: GroupedStoriesResponse = {
  groups: [
    {
      key: mockStatus.id,
      totalCount: 2,
      loadedCount: 0,
      hasMore: true,
      stories: [],
      nextPage: 1,
    },
  ],
  meta: {
    filters: { assigneeIds: ["member-id"] },
    totalGroups: 1,
    groupBy: "status",
    orderBy: "created",
    orderDirection: "desc",
  },
};

const renderBoard = () =>
  render(
    <BoardPropertySlotsProvider
      slots={{
        Provider: passthrough,
        Badges: () => null,
        DisplayPicker: () => null,
        WorkflowCounts,
      }}
    >
      <ListBoard groupedStories={groupedStories} viewOptions={viewOptions} />
    </BoardPropertySlotsProvider>,
  );

beforeEach(() => {
  mockCounts = [];
  mockWorkflowCounts.mockClear();
});

it("uses the unfiltered workflow count for a filtered, partially loaded list", () => {
  mockCounts = [[mockStatus.id, 6]];
  renderBoard();
  expect(mockWorkflowCounts).toHaveBeenCalledWith({
    enabled: true,
    teamId: "team-id",
  });
  expect(screen.getByText("2 stories")).toBeInTheDocument();
  expect(screen.getByText("6/5")).toBeInTheDocument();
  expect(
    screen.getByText("6 stories, limit 5, over limit by 1"),
  ).toBeInTheDocument();
});

it("preserves a server count of zero instead of falling back to stale status counts", () => {
  mockCounts = [[mockStatus.id, 0]];
  renderBoard();
  expect(screen.getByText("0/5")).toBeInTheDocument();
});

it("uses the status API count while workflow counts are unavailable", () => {
  renderBoard();
  expect(screen.getByText("4/5")).toBeInTheDocument();
  expect(screen.queryByText("2/5")).not.toBeInTheDocument();
});
