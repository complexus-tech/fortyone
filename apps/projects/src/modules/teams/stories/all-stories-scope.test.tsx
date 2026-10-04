import { fireEvent, render, screen } from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import type { SavedViewConfiguration } from "@/shared/story/view-configuration";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { useTeamStoriesGrouped } from "@/modules/stories/hooks/use-team-stories-grouped";
import { AllStories } from "./all-stories";

let mockMetadata: Pick<SavedViewConfiguration, "scope"> = {};
jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "ownership-team" }),
}));
jest.mock("./provider", () => ({
  useTeamOptions: () => ({
    viewMetadata: mockMetadata,
    filters: { ...DEFAULT_STORIES_FILTER, teamIds: ["product", "engineering"] },
    viewOptions: {
      groupBy: "priority",
      orderBy: "deadline",
      orderDirection: "asc",
      showSubStories: true,
    },
    setViewOptions: jest.fn(),
  }),
}));
jest.mock("@/components/ui", () => ({
  StoriesBoard: () => <div>Task board</div>,
}));
jest.mock("./stories-skeleton", () => ({
  StoriesSkeleton: () => <div role="status">Loading tasks</div>,
}));
jest.mock("@/modules/stories/hooks/use-my-stories-grouped", () => ({
  useMyStoriesGrouped: jest.fn(),
}));
jest.mock("@/modules/stories/hooks/use-team-stories-grouped", () => ({
  useTeamStoriesGrouped: jest.fn(),
}));
const mockRetry = jest.fn();

describe("saved-view task source", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockMetadata = { scope: { kind: "my-work", tab: "created" } };
    jest
      .mocked(useMyStoriesGrouped)
      .mockReturnValue({
        isFetchedAfterMount: true,
        isError: false,
        data: {},
        refetch: mockRetry,
      } as unknown as ReturnType<typeof useMyStoriesGrouped>);
    jest
      .mocked(useTeamStoriesGrouped)
      .mockReturnValue({ isPending: false, data: {} } as ReturnType<
        typeof useTeamStoriesGrouped
      >);
  });
  it("preserves relationship and multiple-team filters without appending the storage ownership team", () => {
    const { rerender } = render(<AllStories layout="kanban" />);
    expect(useMyStoriesGrouped).toHaveBeenCalledWith(
      "priority",
      expect.objectContaining({
        createdByMe: true,
        teamIds: ["product", "engineering"],
        orderBy: "deadline",
        orderDirection: "asc",
        showSubStories: true,
      }),
      { accountScoped: true },
    );
    expect(useTeamStoriesGrouped).not.toHaveBeenCalled();
    mockMetadata = {};
    rerender(<AllStories layout="list" />);
    expect(useTeamStoriesGrouped).toHaveBeenCalledWith(
      "ownership-team",
      "priority",
      expect.objectContaining({ teamIds: ["ownership-team"] }),
    );
  });
  it("withholds retained task data until the scoped query has freshly resolved", () => {
    jest
      .mocked(useMyStoriesGrouped)
      .mockReturnValue({
        isFetchedAfterMount: false,
        isError: false,
        data: { groups: [{ name: "Previous account" }] },
        refetch: mockRetry,
      } as unknown as ReturnType<typeof useMyStoriesGrouped>);
    render(<AllStories layout="list" />);
    expect(screen.getByRole("status")).toHaveTextContent("Loading tasks");
    expect(screen.queryByText("Task board")).not.toBeInTheDocument();
  });
  it("reports scoped query failure and retries without presenting retained tasks", () => {
    jest
      .mocked(useMyStoriesGrouped)
      .mockReturnValue({
        isFetchedAfterMount: true,
        isError: true,
        data: {},
        refetch: mockRetry,
      } as unknown as ReturnType<typeof useMyStoriesGrouped>);
    render(<AllStories layout="list" />);
    expect(screen.getByRole("alert")).toHaveTextContent("could not be loaded");
    expect(screen.queryByText("Task board")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockRetry).toHaveBeenCalledTimes(1);
  });
});
