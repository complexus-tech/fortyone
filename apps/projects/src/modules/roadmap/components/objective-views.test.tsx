import { fireEvent, render, screen } from "@testing-library/react";
import type { Objective, KeyResult } from "@/modules/objectives/types";
import { DEFAULT_OBJECTIVE_VIEW_OPTIONS } from "../objective-board-utils";
import type { RoadmapLayoutType } from "../types";
import { ObjectiveViews } from "./objective-views";

const OBJECTIVE = {
  id: "objective-1",
  name: "Improve reliability",
  teamId: "team-1",
} as Objective;
const KEY_RESULT = { id: "key-result-1", name: "Reduce errors" } as KeyResult;
const mockBoard = jest.fn();
const mockTimeline = jest.fn();
jest.mock("@/components/ui/roadmap-gantt-board", () => ({
  RoadmapGanttBoard: (props: {
    onObjectiveSelect: (objective: Objective) => void;
  }) => {
    mockTimeline(props);
    return (
      <button
        onClick={() => {
          props.onObjectiveSelect(OBJECTIVE);
        }}
        type="button"
      >
        Select timeline objective
      </button>
    );
  },
}));
jest.mock("@/components/ui/board-skeleton", () => ({
  BoardSkeleton: () => null,
}));
jest.mock("./objectives-board", () => ({
  ObjectivesBoard: (props: {
    onKeyResultSelect: (objective: Objective, keyResult: KeyResult) => void;
  }) => {
    mockBoard(props);
    return (
      <button
        onClick={() => {
          props.onKeyResultSelect(OBJECTIVE, KEY_RESULT);
        }}
        type="button"
      >
        Select key result
      </button>
    );
  },
}));
jest.mock("./objective-details", () => ({
  RoadmapObjectiveDetails: ({
    onClose,
    onKeyResultSelect,
  }: {
    onClose: () => void;
    onKeyResultSelect: (keyResult: KeyResult) => void;
  }) => (
    <aside aria-label="Objective preview">
      <button onClick={onClose} type="button">
        Close objective preview
      </button>
      <button
        onClick={() => {
          onKeyResultSelect(KEY_RESULT);
        }}
        type="button"
      >
        Select preview key result
      </button>
    </aside>
  ),
}));
jest.mock("@/modules/key-results/components/key-result-details", () => ({
  KeyResultDetails: ({ onClose }: { onClose: () => void }) => (
    <aside aria-label="Key result preview">
      <button onClick={onClose} type="button">
        Close key result preview
      </button>
    </aside>
  ),
}));
const props = {
  emptyState: null,
  isPending: false,
  objectives: [OBJECTIVE],
  onCreateObjective: jest.fn(),
  onZoomLevelChange: jest.fn(),
  setViewOptions: jest.fn(),
  viewOptions: DEFAULT_OBJECTIVE_VIEW_OPTIONS,
  zoomLevel: "months" as const,
};
describe("objective preview scope", () => {
  beforeEach(() => jest.clearAllMocks());
  it("exposes the objective panel ID only while that preview is rendered", () => {
    render(<ObjectiveViews {...props} layout="gantt" />);
    expect(mockTimeline).toHaveBeenLastCalledWith(
      expect.objectContaining({ objectiveDetailsPanelId: undefined }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Select timeline objective" }),
    );
    expect(mockTimeline).toHaveBeenLastCalledWith(
      expect.objectContaining({
        selectedObjectiveId: OBJECTIVE.id,
        objectiveDetailsPanelId: `objective-details-${OBJECTIVE.id}`,
      }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Select preview key result" }),
    );
    expect(
      screen.queryByRole("complementary", { name: "Objective preview" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("complementary", { name: "Key result preview" }),
    ).toBeInTheDocument();
    expect(mockTimeline).toHaveBeenLastCalledWith(
      expect.objectContaining({
        selectedObjectiveId: OBJECTIVE.id,
        objectiveDetailsPanelId: undefined,
      }),
    );
  });

  it.each(["list", "kanban"] as const)(
    "keeps objective previews on the timeline after switching to %s",
    (layout: RoadmapLayoutType) => {
      const { rerender } = render(<ObjectiveViews {...props} layout="gantt" />);
      fireEvent.click(
        screen.getByRole("button", { name: "Select timeline objective" }),
      );
      expect(
        screen.getByRole("complementary", { name: "Objective preview" }),
      ).toBeInTheDocument();
      rerender(<ObjectiveViews {...props} layout={layout} />);
      expect(
        screen.queryByRole("complementary", { name: "Objective preview" }),
      ).not.toBeInTheDocument();
      expect(mockBoard).toHaveBeenLastCalledWith(
        expect.not.objectContaining({
          onObjectiveSelect: expect.any(Function),
        }),
      );
      expect(mockBoard).toHaveBeenLastCalledWith(
        expect.objectContaining({ selectedObjectiveId: undefined }),
      );
    },
  );
  it.each(["list", "kanban"] as const)(
    "preserves independent key-result details in %s",
    (layout: RoadmapLayoutType) => {
      render(<ObjectiveViews {...props} layout={layout} />);
      fireEvent.click(
        screen.getByRole("button", { name: "Select key result" }),
      );
      expect(
        screen.getByRole("complementary", { name: "Key result preview" }),
      ).toBeInTheDocument();
      expect(mockBoard).toHaveBeenLastCalledWith(
        expect.objectContaining({ selectedObjectiveId: OBJECTIVE.id }),
      );
      fireEvent.click(
        screen.getByRole("button", { name: "Close key result preview" }),
      );
      expect(
        screen.queryByRole("complementary", { name: "Key result preview" }),
      ).not.toBeInTheDocument();
    },
  );
});
