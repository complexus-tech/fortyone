import { render, screen } from "@testing-library/react";
import type { StoryGroup } from "@/modules/stories/types";
import type { State } from "@/types/states";
import { StoriesKanbanHeader } from "./kanban-header";
import { StoriesHeader } from "./stories-header";

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

const group: StoryGroup = {
  key: "status-id",
  totalCount: 2,
  loadedCount: 0,
  hasMore: true,
  stories: [],
  nextPage: 1,
};
const status: State = {
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
};

describe.each([
  { name: "list", Component: StoriesHeader },
  { name: "Kanban", Component: StoriesKanbanHeader },
])("$name WIP header", ({ Component }) => {
  it.each([4, 5, 6])(
    "shows the team count %i instead of the filtered or loaded task count",
    (activeCount) => {
      render(
        <Component
          group={group}
          groupBy="status"
          isCollapsed={false}
          setIsCollapsed={jest.fn()}
          status={{ ...status, activeCount }}
        />,
      );
      const taskCount = screen.getByText("2 stories");
      const ratio = screen.getByText(`${activeCount}/5`);
      expect(taskCount.compareDocumentPosition(ratio)).toBe(
        Node.DOCUMENT_POSITION_FOLLOWING,
      );
      const indicator = ratio.parentElement;
      if (activeCount > 5) {
        expect(indicator).toHaveClass("text-warning");
        expect(indicator?.querySelector("svg")).not.toBeNull();
      } else {
        expect(indicator).not.toHaveClass("text-warning");
        expect(indicator?.querySelector("svg")).toBeNull();
      }
    },
  );

  it("does not show a status limit when grouping by priority", () => {
    render(
      <Component
        group={group}
        groupBy="priority"
        isCollapsed={false}
        priority="High"
        setIsCollapsed={jest.fn()}
        status={{ ...status, activeCount: 6 }}
      />,
    );
    expect(screen.queryByText("6/5")).not.toBeInTheDocument();
  });
});
